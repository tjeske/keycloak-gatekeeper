// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gatekeeper

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/gorilla/websocket"
)

var upgrader2 = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	// hub *Hub

	// The websocket connection.
	conn *websocket.Conn
}

func (c *Client) closeStreamLog() {
	defer func() {
		// c.hub.unregister <- c
		c.conn.Close()
	}()
	for {
		c.conn.ReadMessage()
		break
	}
}

func (c *Client) streamLog(containerID string, config *types.ContainerJSON) {
	os.Setenv("DOCKER_API_VERSION", "1.25")

	// Docker HTTP API client
	var httpClient *http.Client
	client, _ := client.NewClient(client.DefaultDockerHost, "1.30", httpClient, nil)
	reader, err := client.ContainerLogs(context.Background(), containerID, types.ContainerLogsOptions{
		ShowStderr: true,
		ShowStdout: true,
		Timestamps: false,
		Follow:     true,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()
	if config.Config.Tty {
		scanner := bufio.NewScanner(reader)
		for {
			scanned := scanner.Scan()
			if scanned {
				fmt.Println(c.conn)
				err := c.conn.WriteMessage(websocket.TextMessage, []byte(scanner.Text()+"\n\r"))
				if err != nil {
					return
				}
			} else {
				return
			}
		}
	} else {
		hdr := make([]byte, 8)
		for {
			_, err := reader.Read(hdr)
			if err != nil {
				return
			}
			count := binary.BigEndian.Uint32(hdr[4:])
			dat := make([]byte, count)
			_, err = reader.Read(dat)
			err = c.conn.WriteMessage(websocket.TextMessage, []byte(string(dat)+"\r"))
			if err != nil {
				return
			}
		}
	}
}
