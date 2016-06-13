/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"encoding/json"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/nats-io/nats"
	"github.com/r3labs/nats_to_logstash"
)

type lsConfig struct {
	Hostname string `json:"hostname"`
	Port     int    `json:"port"`
	Timeout  int    `json:"timeout"`
}

func main() {
	messages := []string{"*", "*.*", "*.*.*", "*.*.*.*"}
	nc, err := nats.Connect(os.Getenv("NATS_URI"))
	if err != nil {
		panic("Can't connect to NATS")
	}

	msg, err := nc.Request("config.get.logstash", nil, 1*time.Second)
	if err != nil {
		panic("Can't get logstash config")
	}

	logstash := lsConfig{}
	json.Unmarshal(msg.Data, &logstash)
	l := nats_to_logstash.New(logstash.Hostname, logstash.Port, logstash.Timeout, os.Getenv("NATS_URI"))

	err = l.Subscribe(messages, PreProcess)
	l.Connect()
	l.Writeln(`{"service":"init"}`)

	if err == nil {
		runtime.Goexit()
	}
}

// Message holds the general structure of all messages
type Message struct {
	Datacenter    Datacenter  `json:"datacenter"`
	Routers       []PwdStruct `json:"routers"`
	Networks      []PwdStruct `json:"networks"`
	Executions    []PwdStruct `json:"executions"`
	Firewalls     []PwdStruct `json:"firewalls"`
	Instances     []PwdStruct `json:"instances"`
	Loadbalancers []PwdStruct `json:"loadbalancers"`
	Datacenters   Itemable    `json:"datacenters"`
	Nats          []PwdStruct `json:"nats"`
	Password      string      `json:"datacenter_password"`
}

// Itemable holds any items for any datacenters
type Itemable struct {
	Items []Datacenter `json:"items"`
}

// Datacenter holds datacenter passwords
type Datacenter struct {
	Pwd string `json:"password"`
}

// PwdStruct holds datacenter passwords for other items
type PwdStruct struct {
	Pwd string `json:"datacenter_password"`
}

// PreProcess gets the password and replaces it before writing to a log
func PreProcess(s string) string {
	pwd := getPassword(s)
	if pwd != "" {
		s = strings.Replace(s, pwd, "***", -1)
	}

	return s
}

func getPassword(s string) string {
	m := Message{}
	json.Unmarshal([]byte(s), &m)
	if len(m.Routers) > 0 {
		if m.Routers[0].Pwd != "" {
			return m.Routers[0].Pwd
		}
	}
	if len(m.Networks) > 0 {
		if m.Networks[0].Pwd != "" {
			return m.Networks[0].Pwd
		}
	}
	if len(m.Executions) > 0 {
		if m.Executions[0].Pwd != "" {
			return m.Executions[0].Pwd
		}
	}
	if len(m.Firewalls) > 0 {
		if m.Firewalls[0].Pwd != "" {
			return m.Firewalls[0].Pwd
		}
	}
	if len(m.Instances) > 0 {
		if m.Instances[0].Pwd != "" {
			return m.Instances[0].Pwd
		}
	}
	if len(m.Nats) > 0 {
		if m.Nats[0].Pwd != "" {
			return m.Nats[0].Pwd
		}
	}
	if len(m.Loadbalancers) > 0 {
		if m.Loadbalancers[0].Pwd != "" {
			return m.Loadbalancers[0].Pwd
		}
	}
	if len(m.Datacenters.Items) > 0 {
		if m.Datacenters.Items[0].Pwd != "" {
			return m.Datacenters.Items[0].Pwd
		}
	}
	if m.Password != "" {
		return m.Password
	}

	return m.Datacenter.Pwd
}
