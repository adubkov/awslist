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
	ec2HandlerMsg = "[INFO][%s]: %s %s request from %s. %d Instances was returned.\n"
	elbHandlerMsg = "[INFO][%s]: %s %s request from %s. %d ELB was returned.\n"
	runHttpMsg    = "[INFO] Runing awslist server on port: %d"
)

type HttpServer struct{}

// V1 default handler return instances in old style as a respond.
// For backward compatibility.
func (s *HttpServer) ec2HandlerV1(res http.ResponseWriter, req *http.Request) {
	statusCode := 200
	res.WriteHeader(statusCode)
	data := []string{}
	for _, i := range instances {
		data = append(data, formatInstanceOutputV1(i.Profile.Name, i.Instance))
	}
	content := strings.Join(data, "")

	log.Printf(ec2HandlerMsg,
		req.Host,
		req.Method,
		req.URL,
		req.RemoteAddr,
		len(data))
	printText(res, content)
}

// Default handler return instances as a respond
func (s *HttpServer) ec2Handler(res http.ResponseWriter, req *http.Request) {
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

		log.Printf(ec2HandlerMsg,
			req.Host,
			req.Method,
			req.URL,
			req.RemoteAddr,
			len(data))
		printText(res, content)
	}
}

func (s *HttpServer) elbHandler(res http.ResponseWriter, req *http.Request) {
	var profile, elb string
	var result []Elb

	params := mux.Vars(req)
	statusCode := 200
	res.WriteHeader(statusCode)

	profile = strings.Trim(params["profile"], "/")
	elb = strings.Trim(params["elb"], "/")

	if elb != "" {
		for _, i := range elbs {
			if *i.Elb.LoadBalancerName == elb && i.Profile.Name == profile {
				result = append(result, i)
				break
			}
		}
	} else if profile != "" {
		for _, i := range elbs {
			if i.Profile.Name == profile {
				result = append(result, i)
			}
		}
	} else {
		for _, i := range elbs {
			result = append(result, i)
		}
	}

	switch params["format"] {
	case ".json":
		printJson(res, result)
	default:
		for _, r := range result {
			if elb != "" {
				printText(res, formatElbInstancesOutput(r.Profile.Name, r.Elb))
			} else {
				printText(res, formatElbOutput(r.Profile.Name, r.Elb))
			}
		}
	}

	log.Printf(elbHandlerMsg,
		req.Host,
		req.Method,
		req.URL,
		req.RemoteAddr,
		len(result))

}

func (s *HttpServer) debugHandler(res http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	statusCode := 200
	res.WriteHeader(statusCode)
	text := fmt.Sprintf("%+v", params)
	printText(res, text)
}

// Null handler do nothing but drop connection.
func (s *HttpServer) nullHandler(res http.ResponseWriter, req *http.Request) {}

// runHttpServer runs http listener on specific port
func (s *HttpServer) Run(port int) {

	r := mux.NewRouter()

	// For backward compatibility we should add V1 handler as default
	if *compat {
		r.HandleFunc("/", s.ec2HandlerV1)
	}

	r.HandleFunc(`/favicon.ico`, s.nullHandler)
	r.HandleFunc(`/{ver:(v1|v2)?/?}{type:/?(ec2/?)?}{format:(.json)?}`, s.ec2Handler)
	r.HandleFunc(`/{ver:(v2)?/?}{type:/?(elb/?)?}{format:(.json)?}`, s.elbHandler)
	r.HandleFunc(`/{ver:(v2)?/?}{type:(elb)?}/{profile:([\w\s-]*/?)?}{format:(.json)?}`, s.elbHandler)
	r.HandleFunc(`/{ver:(v2)?/?}{type:(elb)?}/{profile}/{elb:([\w\s-]*/?)?}{format:(.json)?}`, s.elbHandler)

	// Indicate port listening
	log.Printf(runHttpMsg, port)

	// Start listen port or die
	sockaddr := fmt.Sprintf(":%d", port)
	log.Fatal(http.ListenAndServe(sockaddr, r))
}
