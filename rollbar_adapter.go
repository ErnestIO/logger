/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/nats-io/nats"
	"github.com/stvp/rollbar"
)

// RollbarAdapter : Will send logs to a plain file
type RollbarAdapter struct {
	Type        string               `json:"type"`
	Token       string               `json:"token"`
	Environment string               `json:"environment"`
	Subscribers []*nats.Subscription `json:"-"`
	Client      *nats.Conn           `json:"-"`
	File        *os.File             `json:"-"`
}

// NewRollbarAdapter : Rollbar adapter constructor
func NewRollbarAdapter(nc *nats.Conn, config []byte) (a RollbarAdapter, err error) {
	if err := json.Unmarshal(config, &a); err != nil {
		return a, err
	}

	a.Client = nc
	log.Println("Logger set up")

	return a, err
}

// Manage : Manages the subscriptions
func (l *RollbarAdapter) Manage(subjects []string, fn MessageProcessor) (err error) {
	rollbar.Token = l.Token
	rollbar.Environment = l.Environment

	for _, subject := range subjects {
		s, _ := l.Client.Subscribe(subject, func(m *nats.Msg) {
			if strings.Contains(m.Subject, ".error") {
				rollbar.Message("error", m.Subject+" : '"+fn(string(m.Data))+"'")
			} else {
				rollbar.Message("info", m.Subject+" : '"+fn(string(m.Data))+"'")
			}
		})
		l.Subscribers = append(l.Subscribers, s)
	}
	return err
}

// Stop : stops current subscriptions
func (l *RollbarAdapter) Stop() {
	log.Println("Stopping basic logger")
	for _, s := range l.Subscribers {
		if err := s.Unsubscribe(); err != nil {
			log.Println(err.Error())
		}
	}
	if err := l.File.Close(); err != nil {
		log.Println("An error occurred trying to close the file")
		log.Println(err.Error())
	}
	log.SetOutput(os.Stdout)
}
