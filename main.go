/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"encoding/json"
	"log"
	"os"
	"runtime"

	ecc "github.com/ernestio/ernest-config-client"
	"github.com/nats-io/nats"
)

// Adapter : interface for Logger adapters
type Adapter interface {
	Stop()
}

var silent bool
var err error
var nc *nats.Conn
var messages []string
var adapters map[string]Adapter

var newBasicAdapterListener = func(m *nats.Msg) {
	if a, err := NewBasicAdapter(nc, m.Data); err != nil {
		log.Println(err.Error())
		if err := nc.Publish(m.Reply, []byte(`{"error":"`+err.Error()+`"}`)); err != nil {
			log.Println(err.Error())
		}
	} else {
		if err = a.Manage(messages, PreProcess); err != nil {
			log.Println(err.Error())
		}
		adapters["basic"] = &a
		body, _ := json.Marshal(adapters["basic"])
		if err := nc.Publish(m.Reply, body); err != nil {
			log.Println(err.Error())
		}
	}
}

var newLogstashAdapterListener = func(m *nats.Msg) {
	if a, err := NewLogstashAdapter(nc, m.Data); err != nil {
		log.Println(err.Error())
		if err := nc.Publish(m.Reply, []byte(`{"error":"`+err.Error()+`"}`)); err != nil {
			log.Println(err.Error())
		}
	} else {
		if err = a.Manage(messages, PreProcess); err != nil {
			log.Println(err.Error())
		}
		adapters["logstash"] = &a
		body, _ := json.Marshal(adapters["logstash"])
		if err := nc.Publish(m.Reply, body); err != nil {
			log.Println(err.Error())
		}
	}
}

var newRollbarAdapterListener = func(m *nats.Msg) {
	if a, err := NewRollbarAdapter(nc, m.Data); err != nil {
		log.Println(err.Error())
		if err := nc.Publish(m.Reply, []byte(`{"error":"`+err.Error()+`"}`)); err != nil {
			log.Println(err.Error())
		}
	} else {
		if err = a.Manage(messages, PreProcess); err != nil {
			log.Println(err.Error())
		}
		adapters["rollbar"] = &a
		body, _ := json.Marshal(adapters["rollbar"])
		if err := nc.Publish(m.Reply, body); err != nil {
			log.Println(err.Error())
		}
	}
}

// GenericAdapter : Minimal implementation of an adapter
type GenericAdapter struct {
	Type string `json:"type"`
}

var newAdapterListener = func(m *nats.Msg) {
	var adapter GenericAdapter
	if err := json.Unmarshal(m.Data, &adapter); err != nil {
		log.Println("Error processing logger.set message")
		log.Println(err.Error())
	}

	switch adapter.Type {
	case "basic":
		silent = true
		deleteAdapterListener(m)
		silent = false
		newBasicAdapterListener(m)
	case "logstash":
		silent = true
		deleteAdapterListener(m)
		silent = false
		newLogstashAdapterListener(m)
	case "rollbar":
		silent = true
		deleteAdapterListener(m)
		silent = false
		newRollbarAdapterListener(m)
	}
}

var deleteAdapterListener = func(m *nats.Msg) {
	var adapter GenericAdapter
	if err := json.Unmarshal(m.Data, &adapter); err != nil {
		log.Println("Error processing logger.set message")
		log.Println(err.Error())
		if err := nc.Publish(m.Reply, []byte(`{"error":"`+err.Error()+`"}`)); err != nil {
			log.Println(err.Error())
		}
	}

	if adapter.Type == "basic" || adapter.Type == "logstash" || adapter.Type == "rollbar" {
		if _, ok := adapters[adapter.Type]; ok && adapters[adapter.Type] != nil {
			adapters[adapter.Type].Stop()
			adapters[adapter.Type] = nil
			if silent == false {
				body, _ := json.Marshal(adapters[adapter.Type])
				if err := nc.Publish(m.Reply, body); err != nil {
					log.Println(err.Error())
				}
			}
			return
		}
	}
	if silent == false {
		if err := nc.Publish(m.Reply, []byte(`{"error":"Invalid logger type"}`)); err != nil {
			log.Println(err.Error())
		}

	}
}

var findAdapterListener = func(m *nats.Msg) {
	var body []byte
	var err error
	active := make([]Adapter, 0)
	for _, a := range adapters {
		if a != nil {
			active = append(active, a)
		}
	}

	if body, err = json.Marshal(active); err != nil {
		if err := nc.Publish(m.Reply, []byte(`{"error":"Unexpected error ocurred"}`)); err != nil {
			log.Println("An error occurred responding")
		}
	}

	if err := nc.Publish(m.Reply, body); err != nil {
		log.Println("An error occurred responding")
	}
}

// DefaultAdapter : creates the default adapter (basic in this case)
func DefaultAdapter() {
	if path := os.Getenv("ERNEST_LOG_FILE"); path != "" {
		m := nats.Msg{}
		m.Data = []byte(`{"type":"basic","logfile":"` + path + `"}`)
		newBasicAdapterListener(&m)
	}
}

func main() {
	messages = []string{"*", "*.*", "*.*.*", "*.*.*.*"}
	adapters = make(map[string]Adapter)

	nc = ecc.NewConfig(os.Getenv("NATS_URI")).Nats()

	DefaultAdapter()

	if _, err = nc.Subscribe("logger.del", deleteAdapterListener); err != nil {
		log.Println(err.Error())
	}

	if _, err = nc.Subscribe("logger.set", newAdapterListener); err != nil {
		log.Println(err.Error())
	}

	if _, err = nc.Subscribe("logger.find", findAdapterListener); err != nil {
		log.Println(err.Error())
	}

	runtime.Goexit()
}
