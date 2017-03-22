/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package adapters

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/nats-io/nats"
)

var users2service map[string]string

func getUsernameLevel(m *nats.Msg, n *nats.Conn) (username, level string) {
	if users2service == nil {
		users2service = make(map[string]string)
	}
	username = "system"
	level = "debug"
	parts := strings.Split(m.Subject, ".")
	if len(parts) < 4 {
		return
	}
	providers := []string{"aws", "vcloud", "azure", "fake"}
	sw := false
	for _, p := range providers {
		if parts[2] == p {
			sw = true
		}
		if parts[2] == p+"-fake" {
			sw = true
		}
	}
	if !sw {
		return
	}
	var s struct {
		ID string `json:"service"`
	}
	if err := json.Unmarshal(m.Data, &s); err != nil {
		return
	}
	id := s.ID
	if _, ok := users2service[id]; !ok {
		name, err := requestServiceUsername(n, id)
		if err != nil {
			log.Println(err.Error())
			return
		}
		users2service[id] = name
	}
	level = "info"
	if len(parts) == 4 {
		if parts[3] == "error" {
			level = "error"
		}
	}

	return users2service[id], level
}

func requestServiceUsername(n *nats.Conn, id string) (string, error) {
	var s struct {
		UserName string `json:"user_name"`
	}
	res, err := n.Request("service.get", []byte(`{"id":"`+id+`"}`), time.Second)
	if err != nil {
		return "system", err
	}
	if err := json.Unmarshal(res.Data, &s); err != nil {
		return "system", err
	}
	return s.UserName, nil
}
