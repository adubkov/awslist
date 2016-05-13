package main

import (
	"fmt"
	"github.com/gorilla/mux"
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

// Default handler return instances as a respond
func (s *HttpServer) ec2Handler(res http.ResponseWriter, req *http.Request) {

	log.Printf(rootHandlerMsg,
		req.Host,
		req.Method,
		req.URL,
		req.RemoteAddr,
		len(instances))

	params := mux.Vars(req)

	statusCode := 200
	res.WriteHeader(statusCode)
	switch params["format"] {
	case ".json":
		printJson(res, instances)
		return
	default:
		data := []string{}
		for _, i := range instances {
			switch strings.Trim(params["ver"], "/") {
			case "v1":
				data = append(data, formatInstanceOutputV1(i.Profile.Name, i.Instance))
			default:
				data = append(data, formatInstanceOutput(i.Profile.Name, i.Instance))
			}
		}
		content := strings.Join(data, "")
		printText(res, content)
	}
}

func (s *HttpServer) elbHandler(res http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)

	statusCode := 200
	res.WriteHeader(statusCode)

	switch params["format"] {
	case ".json":
		printJson(res, elbs)
		return
	default:
		data := []string{}
		for _, i := range elbs {
			data = append(data, formatElbOutput(i.Profile.Name, i.Elb))
		}
		content := strings.Join(data, "")
		printText(res, content)
	}
}

// Null handler do nothing but drop connection.
func (s *HttpServer) nullHandler(res http.ResponseWriter, req *http.Request) {}

// runHttpServer runs http listener on specific port
func (s *HttpServer) Run(port int) {

	r := mux.NewRouter()

	r.HandleFunc("/favicon.ico", s.nullHandler)
	r.HandleFunc("/{ver:(v1|v2)?/?}{type:/?(ec2/?)?}{format:(.json)?}", s.ec2Handler)
	r.HandleFunc("/{ver:(v2)?/?}{type:/?(elb/?)?}{format:(.json)?}", s.elbHandler)

	// Indicate port listening
	log.Printf(runHttpMsg, port)

	// Start listen port or die
	sockaddr := fmt.Sprintf(":%d", port)
	log.Fatal(http.ListenAndServe(sockaddr, r))
}
