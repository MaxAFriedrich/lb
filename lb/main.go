package main

import (
	"net/http"
	"os"
	"strings"
)

var backendMap *BackendMap

func main() {
	args := os.Args
	backendMapPath := "backend-map.yaml"
	if len(args) == 2 {
		backendMapPath = args[1]
	}
	err := parseBackendMap(backendMapPath)
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/", requestHandler)

	hostPort := strings.Split(backendMap.LbEndpoint, "//")[1]

	err = http.ListenAndServe(hostPort, nil)
	if err != nil {
		panic(err)
	}
}
