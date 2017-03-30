/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	ecc "github.com/ernestio/ernest-config-client"
	ads "github.com/ernestio/logger/adapters"
	"github.com/nats-io/nats"
	"github.com/r3labs/sse"
)

var s *sse.Server
var silent bool
var secret string
var err error
var nc *nats.Conn
var messages []string
var adapters map[string]ads.Adapter
var patternsToObfuscate []string

func register(a *ads.Adapter, m *nats.Msg, err error) {
	if err != nil {
		log.Println(err.Error())
		if err := nc.Publish(m.Reply, []byte(`{"_error":"`+err.Error()+`"}`)); err != nil {
			log.Println(err.Error())
		}
	} else {
		persist(m)
		if err = (*a).Manage(messages, Obfuscate); err != nil {
			log.Println(err.Error())
		}
		adapters[(*a).Name()] = *a
		body, _ := json.Marshal(adapters[(*a).Name()])
		if err := nc.Publish(m.Reply, body); err != nil {
			log.Println(err.Error())
		}
	}
}

var newBasicAdapterListener = func(m *nats.Msg) {
	a, err := ads.NewBasicAdapter(nc, m.Data)
	register(&a, m, err)
}

var newLogstashAdapterListener = func(m *nats.Msg) {
	a, err := ads.NewLogstashAdapter(nc, m.Data)
	register(&a, m, err)
}

var newRollbarAdapterListener = func(m *nats.Msg) {
	a, err := ads.NewRollbarAdapter(nc, m.Data)
	register(&a, m, err)
}

var newSseAdapterListener = func(m *nats.Msg) {
	a, err := ads.NewSseAdapter(nc, m.Data, s)
	register(&a, m, err)
}

// GenericAdapter : Minimal implementation of an adapter
type GenericAdapter struct {
	Type string `json:"type"`
}

var newAdapterListener = func(m *nats.Msg) {
	var adapter GenericAdapter
	if err := json.Unmarshal(m.Data, &adapter); err != nil {
		log.Println("Error processing logger creation")
		log.Println(err.Error())
		return
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
	case "sse":
		silent = true
		deleteAdapterListener(m)
		silent = false
		newSseAdapterListener(m)
	}
}

var deleteAdapterListener = func(m *nats.Msg) {
	var adapter GenericAdapter
	if err := json.Unmarshal(m.Data, &adapter); err != nil {
		log.Println("Error processing logger deletion")
		log.Println(err.Error())
		if err := nc.Publish(m.Reply, []byte(`{"error":"`+err.Error()+`"}`)); err != nil {
			log.Println(err.Error())
			return
		}
	}

	if adapter.Type == "basic" {
		if silent == true {
			fmt.Println("ADAPTER: ")
			fmt.Println(string(m.Data))
			fmt.Println(adapter)

			adapters[adapter.Type].Stop()
			adapters[adapter.Type] = nil
		} else {
			log.Println("Basic adapter is not optional")
			if err := nc.Publish(m.Reply, []byte(`{"error":"Basic logger is not optional"}`)); err != nil {
				log.Println(err.Error())
			}
		}
		return
	}

	if adapter.Type == "logstash" || adapter.Type == "rollbar" || adapter.Type == "sse" {
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
	active := make([]ads.Adapter, 0)
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

var addPatterns = func(m *nats.Msg) {
	var d Datacenter
	if err := json.Unmarshal(m.Data, &d); err != nil {
		log.Println(err.Error())
		return
	}
	addDatacenterPatterns(d, &patternsToObfuscate)
}

// DefaultAdapter : creates the default adapter (basic in this case)
func DefaultAdapter() {
	if err := load(); err != nil {
		if path := os.Getenv("ERNEST_LOG_FILE"); path != "" {
			m := nats.Msg{}
			m.Data = []byte(`{"type":"basic","logfile":"` + path + `"}`)
			newBasicAdapterListener(&m)
		}
	}
}

func httpServer() {
	s = sse.New()
	s.AutoStream = true
	s.EncodeBase64 = true
	defer s.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/logs", authMiddleware)

	var cfg struct {
		Host string `json:"host"`
		Port string `json:"port"`
	}
	msg, err := nc.Request("config.get.logger", []byte(""), 1*time.Second)
	if err != nil {
		panic("Can't get logger config")
	}
	if err := json.Unmarshal(msg.Data, &cfg); err != nil {
		panic("Can't process logger config")
	}

	host := cfg.Host
	port := cfg.Port

	addr := fmt.Sprintf("%s:%s", host, port)
	_ = http.ListenAndServe(addr, mux)
}

func main() {
	messages = []string{"*", "*.*", "*.*.*", "*.*.*.*"}
	adapters = make(map[string]ads.Adapter)

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

	if _, err = nc.Subscribe("datacenter.set", addPatterns); err != nil {
		log.Println(err.Error())
	}

	httpServer()

	runtime.Goexit()
}
