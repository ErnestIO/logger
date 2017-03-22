/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package adapters

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
func NewRollbarAdapter(nc *nats.Conn, config []byte) (Adapter, error) {
	var a RollbarAdapter
	var err error

	if err := json.Unmarshal(config, &a); err != nil {
		return &a, err
	}

	a.Client = nc
	log.Println("Logger set up")

	return &a, err
}

// Manage : Manages the subscriptions
func (l *RollbarAdapter) Manage(subjects []string, fn MessageProcessor) (err error) {
	rollbar.Token = l.Token
	rollbar.Environment = l.Environment

	for _, subject := range subjects {
		s, _ := l.Client.Subscribe(subject, func(m *nats.Msg) {
			if m.Subject == "logger.log" {
				return
			}
			level := "info"
			if strings.Contains(subject, ".error") {
				level = "error"
			}
			l.Log(m.Subject, fn(string(m.Data)), level, "system")
		})
		l.Subscribers = append(l.Subscribers, s)
	}
	return err
}

// Log : Writes a log line
func (l *RollbarAdapter) Log(subject, body, level, user string) {
	rollbar.Message(level, subject+" : '"+body+"'")
}

// Stop : stops current subscriptions
func (l *RollbarAdapter) Stop() {
	log.Println("Stopping rollbar logger")
	for _, s := range l.Subscribers {
		if err := s.Unsubscribe(); err != nil {
			log.Println(err.Error())
		}
	}
}

// Name : get the adapter name
func (l *RollbarAdapter) Name() string {
	return "rollbar"
}
