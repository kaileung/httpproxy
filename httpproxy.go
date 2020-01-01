package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
)

func main() {
	log.Fatalln(http.ListenAndServe(":80",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if "CONNECT" != r.Method {
				// handle http
				proxy := httputil.ReverseProxy{
					Director: func(req *http.Request) {
						req = r
						req.Method = "http"
						req.URL.Host = r.Host
					},
				}
				proxy.ServeHTTP(w, r)
			} else {
				// handle https
				host := r.URL.Host
				hij, ok := w.(http.Hijacker)
				if !ok {
					r.Close = true
					log.Printf("HTTP Server does not support hijacking")
					return
				}
				client, _, err := hij.Hijack()
				if err != nil {
					r.Close = true
					return
				}
				server, err := net.Dial("tcp", host)
				if err != nil {
					r.Close = true
					return
				}
				client.Write([]byte("HTTP/1.0 200 Connection Established\r\n\r\n"))
				// double channel copy
				go io.Copy(server, client)
				go io.Copy(client, server)
			}
		}),
	))
}