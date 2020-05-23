// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"context"
	"encoding/binary"
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:9090", "http service address")

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
} // use default options

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	os.Setenv("DOCKER_API_VERSION", "1.25")

	// Docker HTTP API client
	var httpClient *http.Client
	client, _ := client.NewClient(client.DefaultDockerHost, "1.30", httpClient, nil)
	reader, err := client.ContainerLogs(context.Background(), "test", types.ContainerLogsOptions{
		ShowStderr: true,
		ShowStdout: true,
		Timestamps: false,
		Follow:     true,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	// scanner := bufio.NewScanner(reader)
	// for scanner.Scan() {
	// 	err := c.WriteMessage(websocket.TextMessage, []byte(scanner.Text()+"\n\r"))
	// 	if err != nil {
	// 		log.Println("write:", err)
	// 		break
	// 	}
	// 	fmt.Println(scanner.Text())
	// }
	hdr := make([]byte, 8)
	for {
		_, err := reader.Read(hdr)
		if err != nil {
			log.Fatal(err)
		}
		count := binary.BigEndian.Uint32(hdr[4:])
		dat := make([]byte, count)
		_, err = reader.Read(dat)
		err = c.WriteMessage(websocket.TextMessage, []byte(string(dat)+"\r"))
		if err != nil {
			return
		}
	}
}

func main() {
	http.HandleFunc("/echo", echo)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
