package server

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/n3wscott/k8s-broker-proxy/pkg/controller"
	"github.com/n3wscott/osb-framework-go/pkg/apis/broker/v2"
	osb "github.com/n3wscott/osb-framework-go/pkg/server"
)

var version = "untagged"

type server struct {
	Router     *mux.Router
	Controller v2.BrokerController
}

func CreateServer() *server {
	broker := controller.NewBrokerController()

	osbServer := osb.CreateServer(broker)

	s := server{
		Router:     osbServer.Router,
		Controller: osbServer.Controller,
	}

	s.Router.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)

		json.NewEncoder(rw).Encode(map[string]interface{}{
			"version": version,
		})
	}).Methods("GET")

	return &s
}
