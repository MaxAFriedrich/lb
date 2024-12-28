package main

import (
	"fmt"
	"net/http"
	"strings"
)

func requestHandler(w http.ResponseWriter, r *http.Request) {
	pathComponents := strings.Split(r.URL.Path, "/")
	// expected format: /instance_id/box_id/service_id
	if len(pathComponents) < 4 {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("Invalid path, expected format: /instance_id/box_id/service_id"))
		if err != nil {
			return
		}
		return
	}

	instanceId := pathComponents[1]
	boxId := pathComponents[2]
	serviceId := pathComponents[3]

	host, err := getHostPort(instanceId, boxId, serviceId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err2 := w.Write([]byte("Error getting host"))
		if err2 != nil {
			return
		}
		return
	}

	proxy, err := getProxyType(boxId, serviceId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err2 := w.Write([]byte("Error getting proxy"))
		if err2 != nil {
			return
		}
		return
	}

	if proxy == "" || host == "" {
		w.WriteHeader(http.StatusNotFound)
		_, err2 := w.Write([]byte("Host or proxy not found"))
		if err2 != nil {
			return
		}
		return
	}

	out := fmt.Sprintf("instance_id: %s, box_id: %s, service_id: %s, host: %s, proxy: %s\n", instanceId, boxId, serviceId, host, proxy)

	_, err2 := w.Write([]byte(out))
	if err2 != nil {
		return
	}
}
