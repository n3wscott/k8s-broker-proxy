package controller

import (
	bolt "github.com/coreos/bbolt"
)

type BrokerController struct {
	db *bolt.DB
}

func NewBrokerController(db *bolt.DB) *BrokerController {
	c := BrokerController{
		db: db,
	}
	return &c
}
