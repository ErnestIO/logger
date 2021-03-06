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
	"path/filepath"
	"runtime"
	"time"

	ecc "github.com/ernestio/ernest-config-client"
	ads "github.com/ernestio/logger/adapters"
	"github.com/nats-io/go-nats"
	"github.com/r3labs/broadcast"
)

var silent bool
var secret string
var err error
var nc *nats.Conn
var bc *broadcast.Server
var messages []string
var adapters map[string]ads.Adapter
var patternsToObfuscate []string

func registerAdapter(a *ads.Adapter, m *nats.Msg, err error) {
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
	registerAdapter(&a, m, err)
}

var newLogstashAdapterListener = func(m *nats.Msg) {
	a, err := ads.NewLogstashAdapter(nc, m.Data)
	registerAdapter(&a, m, err)
}

var newRollbarAdapterListener = func(m *nats.Msg) {
	a, err := ads.NewRollbarAdapter(nc, m.Data)
	registerAdapter(&a, m, err)
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

	if adapters[adapter.Type] == nil {
		return
	}

	if adapter.Type == "basic" {
		if silent == true {
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

func setupFilesystem() {
	logfile := os.Getenv("ERNEST_LOG_FILE")
	logcfg := os.Getenv("ERNEST_LOG_CONFIG")

	logdir := filepath.Dir(logfile)

	dirs := []string{logcfg, logdir}

	for _, v := range dirs {
		err := os.MkdirAll(v, 0755)
		if err != nil && !os.IsExist(err) {
			log.Fatal(err)
		}
	}

	f, err := os.Create(logfile)
	if err != nil {
		log.Fatal(err)
	}
	err = f.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	setupFilesystem()

	messages = []string{"*", "*.*", "*.*.*", "*.*.*.*"}
	adapters = make(map[string]ads.Adapter)

	nc = ecc.NewConfig(os.Getenv("NATS_URI")).Nats()

	for {
		// wait for project store to become available
		_, err := getNeedles()
		if err == nil {
			break
		}
		log.Println("could not get secrets")
		time.Sleep(time.Second * 3)
	}

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

	if _, err = nc.Subscribe("logger.log", logListener); err != nil {
		log.Println(err.Error())
	}

	secret = os.Getenv("JWT_SECRET")

	bc = broadcast.New()
	defer bc.Close()

	s := bc.CreateStream("logs")
	s.AutoReplay = false
	s.MaxInactivity = time.Hour * 8760

	// Create new HTTP Server and add the route handler
	mux := http.NewServeMux()
	mux.HandleFunc("/logs", handler)

	// Subscribe to subjects
	_, err = nc.Subscribe(">", natsHandler)
	if err != nil {
		log.Println(err)
		return
	}

	// Start Listening
	addr := fmt.Sprintf("%s:%s", "", "22001")
	_ = http.ListenAndServe(addr, mux)

	runtime.Goexit()
}
