package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	bolt "github.com/coreos/bbolt"
	"github.com/gorilla/mux"
	"github.com/n3wscott/k8s-broker-proxy/pkg/controller"
	"github.com/n3wscott/k8s-broker-proxy/pkg/persistance"
	"github.com/n3wscott/osb-framework-go/pkg/apis/broker/v2"
	osb "github.com/n3wscott/osb-framework-go/pkg/server"
)

var version = "untagged"

type server struct {
	Router     *mux.Router
	Controller v2.BrokerController
	db         *bolt.DB
}

func CreateServer(db *bolt.DB) *server {
	broker := controller.NewBrokerController(db)

	osbServer := osb.CreateServer(broker)

	s := server{
		Router:     osbServer.Router,
		Controller: osbServer.Controller,
		db:         db,
	}

	persistance.SetupBolt(db)

	s.Router.HandleFunc("/", s.home).Methods("GET")

	return &s
}

func (s *server) home(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)

	var metrics bytes.Buffer

	s.db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(persistance.Metrics))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			metrics.WriteString(fmt.Sprintf("key=%s, value=%s\n", k, v))
		}

		return nil
	})

	json.NewEncoder(rw).Encode(map[string]interface{}{
		"version": version,
		"metrics": metrics.String(),
	})
}
