package main

import (
	"crypto/tls"
	"fmt"
	"log"
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
			default:
				panic("Unknown proxy type")
			}
		}
	}

	registerRootHandler(backendMap)

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

func registerHttp(proxyPath string, host string, https bool) {
	hostname, port, err := net.SplitHostPort(host)
	if err != nil {
		panic(err)
	}
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = "http"
			if https {
				req.URL.Scheme = "https"
			}
			req.URL.Host = net.JoinHostPort(hostname, port)
			req.URL.Path = "/" + strings.TrimPrefix(req.URL.Path, proxyPath)
		},
		ModifyResponse: func(response *http.Response) error {
			response.Header.Set("Set-Cookie", "Proxy-Path="+proxyPath+"; Path=/;SameSite=Strict")
			return nil
		},
		Transport: &http.Transport{
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
			DisableKeepAlives: true, // Disable keep-alives to ensure new connections for each request
		},
	}

	http.Handle(proxyPath, proxy)
}

func registerRootHandler(backendMap *BackendMap) {
	proxyPathToHost := backendMap.proxyPathToHost()
	proxyPathToType := backendMap.proxyPathToType()

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			proxyPathCookie, err := req.Cookie("Proxy-Path")
			host := proxyPathToHost[proxyPathCookie.Value]
			if err != nil {
				log.Println(err)
				return
			}
			hostname, port, err := net.SplitHostPort(host)
			if err != nil {
				log.Println(err)
				return
			}
			req.URL.Scheme = proxyPathToType[proxyPathCookie.Value]
			req.URL.Host = net.JoinHostPort(hostname, port)
		},
		Transport: &http.Transport{
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
			DisableKeepAlives: true, // Disable keep-alives to ensure new connections for each request
		},
	}

	http.Handle("/", proxy)
}
