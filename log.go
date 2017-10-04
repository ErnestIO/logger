package main

import (
	"encoding/json"
	"log"

	"github.com/nats-io/nats"
)

// LogMessage holds the message payload
type LogMessage struct {
	Subject string `json:"subject"`
	Message string `json:"message"`
	Level   string `json:"level"`
	User    string `json:"user"`
}

var logListener = func(m *nats.Msg) {
	var l LogMessage
	if err := json.Unmarshal(m.Data, &l); err != nil {
		log.Println(err.Error())
		return
	}

	for _, adapter := range adapters {
		adapter.Log(l.Subject, l.Message, l.Level, l.User)
	}
}
