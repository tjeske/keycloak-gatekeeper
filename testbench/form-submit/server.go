// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
)

var addr = flag.String("addr", "localhost:9090", "http service address")

func updateTemplate(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	// name := chi.URLParam(req, "query")
	// app := storageProvider.GetTemplateByName(name)
	js, err := json.Marshal(req.Form)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println(req.ParseForm())

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func main() {
	fs := http.FileServer(http.Dir("./dist"))
	http.Handle("/", fs)
	http.HandleFunc("/udesk/updateTemplate/", updateTemplate)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
