/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package adapters

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/nats-io/nats"
	"github.com/r3labs/sse"
)

// SseAdapter : Will send logs to a plain file
type SseAdapter struct {
	Type        string               `json:"type"`
	UUID        string               `json:"uuid"`
	Pipe        *sse.Server          `json:"-"`
	Subscribers []*nats.Subscription `json:"-"`
	Client      *nats.Conn           `json:"-"`
	File        *os.File             `json:"-"`
}

type sseMessage struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
	Level   string `json:"level"`
}

// NewSseAdapter : Basic adapter constructor
func NewSseAdapter(nc *nats.Conn, config []byte, s *sse.Server) (Adapter, error) {
	var l SseAdapter
	var err error

	if err := json.Unmarshal(config, &l); err != nil {
		return &l, err
	}

	l.Pipe = s
	l.Client = nc
	log.Println("SSE logger set up")

	return &l, err
}

// Manage : Manages the subscriptions
func (l *SseAdapter) Manage(subjects []string, fn MessageProcessor) (err error) {
	for _, subject := range subjects {
		log.Println("Subscribed to " + subject)
		s, _ := l.Client.Subscribe(subject, func(m *nats.Msg) {
			output := fmt.Sprintf("%s", fn(string(m.Data)))
			log.Println("Publising to " + l.UUID + " : " + output)
			body, _ := json.Marshal(sseMessage{
				Subject: m.Subject,
				Body:    output,
				Level:   "info",
			})
			l.Pipe.Publish(l.UUID, body)
		})
		l.Subscribers = append(l.Subscribers, s)
	}
	return err
}

// Stop : stops current subscriptions
func (l *SseAdapter) Stop() {
	log.Println("Stopping sse logger")
	for _, s := range l.Subscribers {
		if err := s.Unsubscribe(); err != nil {
			log.Println(err.Error())
		}
	}
	l.Pipe.Close()
}

// Name : get the adapter name
func (l *SseAdapter) Name() string {
	return "sse"
}
