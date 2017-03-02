/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package adapters

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/nats-io/nats"
)

// LogstashAdapter : Adapter for logging to logstash
type LogstashAdapter struct {
	Type        string               `json:"type"`
	Hostname    string               `json:"hostname"`
	Port        int                  `json:"port"`
	Timeout     int                  `json:"timeout"`
	Subscribers []*nats.Subscription `json:"-"`
	Client      *nats.Conn           `json:"-"`
}

// LogMessage : Message to be sent to logstash
type LogMessage struct {
	Subject string      `json:"subject"`
	Message interface{} `json:"message"`
}

// NewLogstashAdapter : LogstashAdapter constructor
func NewLogstashAdapter(nc *nats.Conn, config []byte) (Adapter, error) {
	var l LogstashAdapter
	var err error

	if err = json.Unmarshal(config, &l); err != nil {
		return &l, err
	}

	l.Client = nc

	if err := l.writeln([]byte(`{"service":"initial"}`)); err != nil {
		log.Println(err.Error())
	}

	return &l, nil
}

// Manage : Manages the subscriptions
func (l *LogstashAdapter) Manage(subjects []string, fn MessageProcessor) (err error) {
	for _, subject := range subjects {
		s, _ := l.Client.Subscribe(subject, func(m *nats.Msg) {

			lg := LogMessage{
				Subject: m.Subject,
				Message: fn(string(m.Data)),
			}
			if body, err := json.Marshal(lg); err != nil {
				log.Println(err.Error())
			} else {
				if err = l.writeln(body); err != nil {
					log.Println(err.Error())
				}
			}
		})
		l.Subscribers = append(l.Subscribers, s)
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

func (l *LogstashAdapter) writeln(message []byte) (err error) {
	port := strconv.Itoa(l.Port)
	url := "http://" + l.Hostname + ":" + port
	println(string(url))
	println(string(message))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(message))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

	if err := resp.Body.Close(); err != nil {
		return err
	}
	return nil
}

// Name : get the adapter name
func (l *LogstashAdapter) Name() string {
	return "logstash"
}
