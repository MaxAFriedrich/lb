package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
)

type BackendHost struct {
	Host       string
	Port       string
	Proxy      string
	TargetPath string
}

func handleError(w http.ResponseWriter, err error, code int) {
	w.WriteHeader(code)
	_, err2 := w.Write([]byte(err.Error()))
	if err2 != nil {
		return
	}
	log.Println(err)
}

func requestHandler(w http.ResponseWriter, r *http.Request) {
	// TODO parse the url to get the stuff for the proxy
	pathComponents := strings.Split(r.URL.Path, "/")
	// expected format: /instance_id/box_id/service_id/target_path
	if len(pathComponents) < 5 {
		handleError(w, fmt.Errorf("invalid path, expected format: /instance_id/box_id/service_id/target_path"), http.StatusBadRequest)
		return
	}

	instanceId := pathComponents[1]
	boxId := pathComponents[2]
	serviceId := pathComponents[3]
	targetPath := pathComponents[4]

	host, err := getHostPort(instanceId, boxId, serviceId)
	if err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}

	proxy, err := getProxyType(boxId, serviceId)
	if err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}

	if proxy == "" || host == "" {
		handleError(w, fmt.Errorf("host or proxy not found"), http.StatusNotFound)
		return
	}

	hostname, port, err := net.SplitHostPort(host)
	if err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}

	backHost := BackendHost{
		Host:       hostname,
		Port:       port,
		Proxy:      proxy,
		TargetPath: targetPath,
	}

	proxyRequest(w, r, backHost)
}

func proxyRequest(w http.ResponseWriter, r *http.Request, host BackendHost) {
	switch host.Proxy {
	case "http":
		proxyHTTP(w, r, host, false)
		break
	case "https":
		proxyHTTP(w, r, host, true)
		break
	case "ssh":
		proxySSH(w, r, host)
		break
	case "udp":
		proxyUDP(w, r, host)
		break
	case "tcp":
		proxyTCP(w, r, host)
		break
	default:
		handleError(w, fmt.Errorf("unsupported proxy type"), http.StatusNotImplemented)
	}
}

func proxySSH(w http.ResponseWriter, r *http.Request, host BackendHost) {
	handleError(w, fmt.Errorf("ssh proxy not implemented"), http.StatusNotImplemented)
}

func proxyUDP(w http.ResponseWriter, r *http.Request, host BackendHost) {
	handleError(w, fmt.Errorf("udp proxy not implemented"), http.StatusNotImplemented)
}

func proxyTCP(w http.ResponseWriter, r *http.Request, host BackendHost) {
	handleError(w, fmt.Errorf("tcp proxy not implemented"), http.StatusNotImplemented)
}
