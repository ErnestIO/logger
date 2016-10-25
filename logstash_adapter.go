/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/nats-io/nats"
	"github.com/r3labs/nats_to_logstash"
)

// LogstashAdapter : Adapter for logging to logstash
type LogstashAdapter struct {
	Type        string                           `json:"type"`
	Hostname    string                           `json:"hostname"`
	Port        int                              `json:"port"`
	Timeout     int                              `json:"timeout"`
	Client      *nats_to_logstash.NatsToLogstash `json:"-"`
	Subscribers []*nats.Subscription             `json:"-"`
}

// NewLogstashAdapter : LogstashAdapter constructor
func NewLogstashAdapter(nc *nats.Conn, config []byte) (l LogstashAdapter, err error) {
	if err = json.Unmarshal(config, &l); err != nil {
		return l, err
	}

	l.Client = nats_to_logstash.New(l.Hostname, l.Port, l.Timeout, os.Getenv("NATS_URI"))

	return l, nil
}

// Manage : Manages the subscriptions
func (l *LogstashAdapter) Manage(subjects []string, fn nats_to_logstash.MessageProcessor) error {
	for _, subject := range subjects {
		s, _ := l.Client.SingleSubscription(subject, fn)
		l.Subscribers = append(l.Subscribers, s)
	}
	if err := l.Client.Writeln(`{"service":"init"}`); err != nil {
		log.Println(err.Error())
	}
	if _, err := l.Client.Connect(); err != nil {
		log.Println(err.Error())
	}

	return err
}

// Stop : stops current subscriptions
func (l *LogstashAdapter) Stop() {
	for _, s := range l.Subscribers {
		if err := s.Unsubscribe(); err != nil {
			log.Println(err.Error())
		}
	}
}
