// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:9090", "http service address")

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
} // use default options

var rb = NewRingBuffer(100)

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		data := <-rb.Output
		err := c.WriteMessage(websocket.TextMessage, bytes.ReplaceAll(data.([]byte), []byte("\n"), []byte("\n\r")))
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func main() {
	go func() {
		for {
			rb.Write([]byte("ABC"))
			time.Sleep(1 * time.Second)
		}
	}()
	http.HandleFunc("/echo", echo)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
