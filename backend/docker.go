// Copyright © 2019 Tobias Jeske
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package backend

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/docker/cli/cli/command"
	cmd_container "github.com/docker/cli/cli/command/container"
	cmd_build "github.com/docker/cli/cli/command/image"
	cliflags "github.com/docker/cli/cli/flags"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/tjeske/containerflight/util"
	"golang.org/x/net/context"
)

// "mock connectors" for unit-tesing
var filesystem = afero.NewOsFs()

type BackendDockerConfig struct {
}

type DockerContext interface {
	GetDockerFile(args map[string]string) string
	GetDockerFileContextFiles(args map[string]string) map[string]string
}

type dockerHttpApiClient interface {
	ImageList(ctx context.Context, options types.ImageListOptions) ([]types.ImageSummary, error)
	ImageRemove(ctx context.Context, imageID string, options types.ImageRemoveOptions) ([]types.ImageDeleteResponseItem, error)
}

type dockerCliClient interface {
	command.Cli
}

// DockerClient abstracts the containerflight communication with a moby daemon
type DockerClient struct {
	client    dockerHttpApiClient
	dockerCli dockerCliClient
	version   string
}

var notWordChar = regexp.MustCompile("\\W")

// NewDockerClient creates a new Docker client using API 1.25 (implemented by Docker 1.13)
func NewDockerClient(version string) *DockerClient {
	os.Setenv("DOCKER_API_VERSION", "1.25")

	// Docker HTTP API client
	var httpClient *http.Client
	client, err := client.NewClient(client.DefaultDockerHost, "1.30", httpClient, nil)
	util.CheckErr(err)

	// Docker cli client
	dockerCli, err := command.NewDockerCli(command.WithStandardStreams())
	util.CheckErr(err)
	opts := cliflags.NewClientOptions()
	err = dockerCli.Initialize(opts)
	util.CheckErr(err)

	return &DockerClient{client: client, dockerCli: dockerCli, version: version}
}

func NewDockerClientWithWriter(version string, w io.Writer) *DockerClient {
	os.Setenv("DOCKER_API_VERSION", "1.25")

	// Docker HTTP API client
	var httpClient *http.Client
	client, err := client.NewClient(client.DefaultDockerHost, "1.30", httpClient, nil)
	util.CheckErr(err)

	// Docker cli client
	dockerCli, err := command.NewDockerCli(command.WithCombinedStreams(w))
	util.CheckErr(err)
	opts := cliflags.NewClientOptions()
	err = dockerCli.Initialize(opts)
	util.CheckErr(err)

	return &DockerClient{client: client, dockerCli: dockerCli, version: version}
}

// build a Docker container
func (dc *DockerClient) build(label, containerName, userName, hashStr string, args map[string]string, dockerCtx DockerContext) {

	// remove all previous images
	dc.removeImages(label)

	tmpDir, err := afero.TempDir(filesystem, "", "udesk")
	util.CheckErr(err)
	defer filesystem.RemoveAll(tmpDir)

	err = os.Chmod(tmpDir, 0777)
	util.CheckErr(err)

	dockerFileName := tmpDir + "/Dockerfile"
	dockerFile, err := filesystem.Create(dockerFileName)
	util.CheckErr(err)

	dockerfileContent := dockerCtx.GetDockerFile(args)
	_, err = dockerFile.Write([]byte(dockerfileContent))
	util.CheckErr(err)
	defer dockerFile.Close()

	dockerCtxFiles := dockerCtx.GetDockerFileContextFiles(args)
	for fileName := range dockerCtxFiles {
		file, err := filesystem.Create(tmpDir + "/" + fileName)
		util.CheckErr(err)

		content := dockerCtxFiles[fileName]
		_, err = file.Write([]byte(content))
		util.CheckErr(err)
		defer file.Close()
	}

	cmdDockerRun := cmd_build.NewBuildCommand(dc.dockerCli)

	buildCmdArgs := dc.getBuildCmdArgs(dockerFileName, tmpDir, label, containerName, userName, hashStr)
	cmdDockerRun.SetArgs(buildCmdArgs)
	cmdDockerRun.SilenceErrors = true
	cmdDockerRun.SilenceUsage = true

	log.Debug("execute \"docker build " + strings.Join(buildCmdArgs, " ") + "\"")

	err = cmdDockerRun.Execute()
	util.CheckErr(err)
}

// removeImages destroys all Docker images with the specific label
func (dc *DockerClient) removeImages(label string) {
	client := dc.client
	images, err := client.ImageList(context.Background(), types.ImageListOptions{})
	util.CheckErr(err)

	for _, image := range images {
		tagFound := false
		for _, tag := range image.RepoTags {
			if tag == label {
				tagFound = true
				break
			}
		}
		if tagFound {
			// remove image
			options := types.ImageRemoveOptions{Force: true, PruneChildren: true}
			client.ImageRemove(context.Background(), image.ID, options)
			util.CheckErr(err)
		}
	}
}

// get Docker build command args
func (dc *DockerClient) getBuildCmdArgs(dockerfile string, dockerBuildCtx string, label, containerName, userName, hashStr string) []string {
	buildCmd := []string{
		dockerBuildCtx,
		"-f", dockerfile,
		"--label", "udesk=" + "true",
		"--label", "udesk_hash=" + hashStr,
		"--label", "udesk_name=" + containerName,
		"--label", "udesk_owner=" + userName,
		"-t", label,
	}

	return buildCmd
}

func (dc *DockerClient) Run(containerName, userName string, args2 map[string]string, dockerBuildCtx DockerContext, dockerRunArgs []string, args []string, cb func()) {
	// defer cb()
	cmdDockerRun := cmd_container.NewRunCommand(dc.dockerCli)

	imageID := dc.getImageID(containerName, userName, args2, dockerBuildCtx)

	dockerRunCmdArgs := dc.getRunCmdArgs(dockerRunArgs, imageID, args)
	cmdDockerRun.SetArgs(dockerRunCmdArgs)
	cmdDockerRun.SilenceErrors = true
	cmdDockerRun.SilenceUsage = true

	log.Info("execute \"docker run " + strings.Join(dockerRunCmdArgs, " ") + "\"")

	err := cmdDockerRun.Execute()
	util.CheckErr(err)
}

func (dc *DockerClient) Create(containerName, userName string, args2 map[string]string, dockerBuildCtx DockerContext, dockerRunArgs []string, args []string, cb func()) {
	// defer cb()
	cmdDockerRun := cmd_container.NewCreateCommand(dc.dockerCli)

	imageID := dc.getImageID(containerName, userName, args2, dockerBuildCtx)

	dockerRunCmdArgs := dc.getRunCmdArgs(dockerRunArgs, imageID, args)
	cmdDockerRun.SetArgs(dockerRunCmdArgs)
	cmdDockerRun.SilenceErrors = true
	cmdDockerRun.SilenceUsage = true

	log.Info("execute \"docker create " + strings.Join(dockerRunCmdArgs, " ") + "\"")

	err := cmdDockerRun.Execute()
	util.CheckErr(err)
}

func (dc *DockerClient) Start(containerName string) {
	// defer cb()
	cmdDockerRun := cmd_container.NewStartCommand(dc.dockerCli)

	dockerRunCmdArgs := []string{containerName} //dc.getRunCmdArgs(dockerRunArgs, imageID, args)
	cmdDockerRun.SetArgs(dockerRunCmdArgs)
	cmdDockerRun.SilenceErrors = true
	cmdDockerRun.SilenceUsage = true

	log.Info("execute \"docker start " + containerName + "\"")

	err := cmdDockerRun.Execute()
	util.CheckErr(err)
}

func (dc *DockerClient) GetStatus() []types.Container {
	opts := types.ContainerListOptions{All: true,
		Filters: filters.NewArgs(filters.Arg("label", "udesk")),
	}
	containers, err := dc.dockerCli.Client().ContainerList(context.Background(), opts)
	util.CheckErr(err)
	return containers
}

func (dc *DockerClient) GetContainer(uuid string) (*types.ContainerJSON, error) {
	opts := types.ContainerListOptions{All: true,
		Filters: filters.NewArgs(filters.Arg("label", "udesk_uuid="+uuid)),
	}
	container, err := dc.dockerCli.Client().ContainerList(context.Background(), opts)
	if err != nil {
		return nil, err
	}
	if len(container) > 0 {
		c, err := dc.dockerCli.Client().ContainerInspect(context.Background(), container[0].ID)
		if err != nil {
			return nil, err
		}
		return &c, nil
	}
	return nil, errors.New("found more than one container")
}

func (dc *DockerClient) RemoveContainer(uuid string) error {
	container, err := dc.GetContainer(uuid)
	if err != nil {
		return err
	}
	err = dc.dockerCli.Client().ContainerRemove(context.Background(), container.Name, types.ContainerRemoveOptions{Force: true})
	return err
}

func (dc *DockerClient) PauseContainer(uuid string) error {
	container, err := dc.GetContainer(uuid)
	if err != nil {
		return err
	}
	err = dc.dockerCli.Client().ContainerPause(context.Background(), container.Name)
	return err
}

func (dc *DockerClient) UnpauseContainer(uuid string) error {
	container, err := dc.GetContainer(uuid)
	if err != nil {
		return err
	}
	err = dc.dockerCli.Client().ContainerUnpause(context.Background(), container.Name)
	return err
}

// return Docker image Id, if image does not exists build it
func (dc *DockerClient) getImageID(containerName, userName string, args map[string]string, dockerCtx DockerContext) string {

	containerLabel := dc.getDockerContainerLabel(containerName, userName)
	hashStr := dc.getDockerContainerHash(args, dockerCtx)

	imageID, err := dc.getDockerContainerImageID(hashStr)
	if err != nil {
		dc.build(containerLabel, containerName, userName, hashStr, args, dockerCtx)
		imageID, err = dc.getDockerContainerImageID(hashStr)
		util.CheckErr(err)
	}
	return imageID
}

// get Docker run command args
func (dc *DockerClient) getRunCmdArgs(dockerRunArgs []string, imageID string, args []string) []string {
	runCmdArgs := []string{
		"--rm",
	}

	runCmdArgs = append(runCmdArgs, dockerRunArgs...)
	runCmdArgs = append(runCmdArgs, imageID)
	runCmdArgs = append(runCmdArgs, args...)

	return runCmdArgs
}

// getDockerContainerImageID returns the Docker image ID for an app hash value
func (dc *DockerClient) getDockerContainerImageID(hashStr string) (string, error) {
	fmt.Println("ABC")
	images, err := dc.client.ImageList(context.Background(), types.ImageListOptions{})
	util.CheckErr(err)
	imageID := ""
	for _, image := range images {
		imgHash := image.Labels["udesk_hash"]
		if hashStr == imgHash {
			imageID = image.ID
			break
		}
	}
	if imageID != "" {
		return imageID, nil
	}
	return "", fmt.Errorf("cannot find image with ID `%s`", hashStr)
}

// generate a container label
func (dc *DockerClient) getDockerContainerLabel(containerName, userName string) string {
	return "udesk_" + strings.ToLower(containerName) + "_" + strings.ToLower(userName)
}

// get the corresponding hash value for an app file
func (dc *DockerClient) getDockerContainerHash(args map[string]string, dockerCtx DockerContext) string {

	hash := sha256.New()

	// hash containerflight version
	hash.Write([]byte(dc.version))

	// hash config file
	appConfigBytes := []byte(dockerCtx.GetDockerFile(args))
	hash.Write(appConfigBytes)

	// hash Docker build context
	for _, file := range dockerCtx.GetDockerFileContextFiles(args) {
		fileBytes := []byte(file)
		hash.Write(fileBytes)
	}

	hashStr := hex.EncodeToString(hash.Sum(nil))
	return hashStr
}
