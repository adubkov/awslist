package main

import (
	"fmt"
	"log"
	"net/http"
)

var (
	// @readonly
	rootHandlerMsg = "[INFO][%s]: %s %s request from %s. %d instances was returned.\n"
	runHttpMsg     = "[INFO] Runing awslist server on port: %d"
)

type HttpServer struct{}

// Default handler return instances as a respond
func (s *HttpServer) defaultHandler(res http.ResponseWriter, req *http.Request) {
	log.Printf(rootHandlerMsg,
		req.Host,
		req.Method,
		req.URL,
		req.RemoteAddr,
		len(instances))

	statusCode := 200
	res.WriteHeader(statusCode)
	for _, i := range instances {
		s := formatInstanceOutput(i.Profile.Name, i.Instance)
		fmt.Fprintf(res, s)
	}
}

// Default handler return instances as a respond
func (s *HttpServer) defaultHandlerV1(res http.ResponseWriter, req *http.Request) {
	log.Printf(rootHandlerMsg,
		req.Host,
		req.Method,
		req.URL,
		req.RemoteAddr,
		len(instances))

	statusCode := 200
	res.WriteHeader(statusCode)
	for _, i := range instances {
		s := formatInstanceOutputV1(i.Profile.Name, i.Instance)
		fmt.Fprintf(res, s)
	}
}

// Null handler do nothing but drop connection.
func (s *HttpServer) nullHandler(res http.ResponseWriter, req *http.Request) {}

// runHttpServer runs http listener on specific port
func (s *HttpServer) Run(port int) {
	// Set default http handler
	http.HandleFunc("/favicon.ico", s.nullHandler)
	http.HandleFunc("/", s.defaultHandler)
	http.HandleFunc("/v1", s.defaultHandlerV1)
	http.HandleFunc("/v2", s.defaultHandler)

	// Indicate port listening
	log.Printf(runHttpMsg, port)

	// Start listen port or die
	sockaddr := fmt.Sprintf(":%d", port)
	log.Fatal(http.ListenAndServe(sockaddr, nil))
}
