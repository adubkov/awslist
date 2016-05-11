package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

var (
	// @readonly
	rootHandlerMsg = "[INFO][%s]: %s %s request from %s. %d instances was returned.\n"
	runHttpMsg     = "[INFO] Runing awslist server on port: %d"
)

type HttpServer struct{}

// Default handler return screen buffer as a respond
func (s *HttpServer) defaultHandler(res http.ResponseWriter, req *http.Request) {
	log.Printf(rootHandlerMsg,
		req.Host,
		req.Method,
		req.URL,
		req.RemoteAddr,
		len(screen_buffer))

	statusCode := 200
	res.WriteHeader(statusCode)

	fmt.Fprintf(res, strings.Join(screen_buffer, "\n"))
}

// Null handler do nothing but drop connection.
func (s *HttpServer) nullHandler(res http.ResponseWriter, req *http.Request) {}

// runHttpServer runs http listener on specific port
func (s *HttpServer) Run(port int) {
	// Set default http handler
	http.HandleFunc("/favicon.ico", s.nullHandler)
	http.HandleFunc("/", s.defaultHandler)

	// Indicate port listening
	log.Printf(runHttpMsg, port)

	// Start listen port or die
	sockaddr := fmt.Sprintf(":%d", port)
	log.Fatal(http.ListenAndServe(sockaddr, nil))
}
