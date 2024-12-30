package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
)

func main() {
	args := os.Args
	backendMapPath := "backend-map.yaml"
	if len(args) == 2 {
		backendMapPath = args[1]
	}

	var backendMap *BackendMap
	err := parseBackendMap(backendMapPath, &backendMap)
	if err != nil {
		panic(err)
	}

	for _, backend := range backendMap.Backends {
		for _, service := range backend.Services {
			proxyPath := fmt.Sprintf("/%s/%s/%s/", backend.ID, service.BoxID, service.ServiceID)
			cleanPath := fmt.Sprintf("/%s/%s/%s", backend.ID, service.BoxID, service.ServiceID)
			proxyType, err := backendMap.getProxyType(service.BoxID, service.ServiceID)
			if err != nil {
				panic(err)
			}

			http.HandleFunc(cleanPath, addCleanRedirect)

			switch proxyType {
			case "http":
				registerHttp(proxyPath, service.Host, false)
				break
			case "https":
				registerHttp(proxyPath, service.Host, true)
				break
			case "ssh":
				registerSsh(proxyPath, service.Host)
			default:
				panic("Unknown proxy type")
			}
		}
	}

	hostPort := strings.Split(backendMap.LbEndpoint, "//")[1]

	err = http.ListenAndServe(hostPort, nil)
	if err != nil {
		panic(err)
	}
}

func addCleanRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, r.URL.Path+"/", http.StatusMovedPermanently)
	//http.Error(w, "Redirecting to: "+r.URL.Path+"/", http.StatusMovedPermanently)
	_, err := http.ResponseWriter(w).Write([]byte("Redirecting to: " + r.URL.Path + "/"))
	if err != nil {
		return
	}
}

func registerSsh(path string, host string) {
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "SSH not supported, would connect to: "+host, http.StatusNotImplemented)
	})
}

func registerHttp(proxyPath string, host string, https bool) {
	hostname, port, err := net.SplitHostPort(host)
	if err != nil {
		panic(err)
	}
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			origin := req.URL.Host
			ctx := context.WithValue(req.Context(), "origin", origin)
			req = req.WithContext(ctx)

			req.URL.Scheme = "http"
			if https {
				req.URL.Scheme = "https"
			}
			req.URL.Host = net.JoinHostPort(hostname, port)
			req.URL.Path = "/" + strings.TrimPrefix(req.URL.Path, proxyPath)
		},
		ModifyResponse: func(response *http.Response) error {
			origin, ok := response.Request.Context().Value("origin").(string)
			if !ok || origin == "" {
				return nil // or handle the error appropriately
			}

			// make sure if the origin is in the body, it is replaced with the proxy path
			// TODO shim document.location, document.URL, window.location to use the proxy path
			// TODO check for links that start with / and replace them with the proxy path
			bodyBytes, err := io.ReadAll(response.Body)
			if err != nil {
				return err
			}
			bodyString := string(bodyBytes)
			bodyString = strings.ReplaceAll(bodyString, origin, proxyPath)
			response.Body = io.NopCloser(strings.NewReader(bodyString))

			return nil
		},
		Transport: &http.Transport{
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
			DisableKeepAlives: true, // Disable keep-alives to ensure new connections for each request
		},
	}

	http.Handle(proxyPath, proxy)
}
