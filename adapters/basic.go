/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package adapters

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"

	"github.com/nats-io/nats"
)

// BasicAdapter : Will send logs to a plain file
type BasicAdapter struct {
	Type        string               `json:"type"`
	LogFile     string               `json:"logfile"`
	Subscribers []*nats.Subscription `json:"-"`
	Client      *nats.Conn           `json:"-"`
	File        *os.File             `json:"-"`
}

// NewBasicAdapter : Basic adapter constructor
func NewBasicAdapter(nc *nats.Conn, config []byte) (Adapter, error) {
	var a BasicAdapter
	var err error

	if err := json.Unmarshal(config, &a); err != nil {
		return &a, err
	}

	if _, err := os.Stat(a.LogFile); os.IsNotExist(err) {
		return &a, errors.New("Specified file '" + a.LogFile + "' does not exist")
	}

	a.File, err = os.OpenFile(a.LogFile, os.O_WRONLY|os.O_APPEND, 0640)
	if err != nil {
		log.Fatalln(err)
		return &a, errors.New("Seems I don't have permissions to write on " + a.LogFile)
	}
	log.SetOutput(io.MultiWriter(a.File, os.Stdout))

	a.Client = nc
	log.Println("Logger set up")

	return &a, err
}

// Manage : Manages the subscriptions
func (l *BasicAdapter) Manage(subjects []string, fn MessageProcessor) (err error) {
	for _, subject := range subjects {
		s, _ := l.Client.Subscribe(subject, func(m *nats.Msg) {
			if m.Subject == "logger.log" {
				return
			}
			l.Log(m.Subject, fn(string(m.Data)), "debug", "system")
		})
		l.Subscribers = append(l.Subscribers, s)
	}
	return err
}

// Log : Writes a log line
func (l *BasicAdapter) Log(subject, body, level, user string) {
	log.Println("level=" + level + " user=" + user + " : " + subject + "  '" + body + "'")
}

// Stop : stops current subscriptions
func (l *BasicAdapter) Stop() {
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

// Name : get the adapter name
func (l *BasicAdapter) Name() string {
	return "basic"
}
