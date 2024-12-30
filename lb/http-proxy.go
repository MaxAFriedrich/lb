package main

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httputil"
)

func proxyHTTP(w http.ResponseWriter, r *http.Request, host BackendHost, isSecure bool) {
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = "http"
			if isSecure {
				req.URL.Scheme = "https"
			}
			req.URL.Host = net.JoinHostPort(host.Host, host.Port)
			req.URL.Path = host.TargetPath
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	proxy.ServeHTTP(w, r)
}
