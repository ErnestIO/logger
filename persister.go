/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/nats-io/nats"
)

// Persistence : representation of the persisted file
type Persistence struct {
	Basic    []byte `json:"basic"`
	Logstash []byte `json:"logstash"`
	Rollbar  []byte `json:"rollbar"`
}

func persist(m *nats.Msg) {
	var per Persistence
	var adapter GenericAdapter
	file := ".logger"

	if path := os.Getenv("ERNEST_LOG_CONFIG"); path != "" {
		file = path + file
	}

	if _, err := os.Stat(file); os.IsNotExist(err) {
		f, err := os.Create(file)
		if err != nil {
			log.Println("Can't create persistence file '" + file + "'")
			return
		}
		err = ioutil.WriteFile(file, []byte("{}"), 0644)
		defer f.Close()
	}

	dat, err := ioutil.ReadFile(file)
	if err != nil {
		log.Println("Error reading " + file + " file")
		return
	}
	if err = json.Unmarshal(dat, &per); err != nil {
		log.Println("Persistence file is corrupted")
		return
	}

	if err := json.Unmarshal(m.Data, &adapter); err != nil {
		log.Println("Error processing logger.set message")
		log.Println(err.Error())
	}

	switch adapter.Type {
	case "basic":
		per.Basic = m.Data
	case "logstash":
		per.Logstash = m.Data
	case "rollbar":
		per.Rollbar = m.Data
	}

	body, err := json.Marshal(per)
	err = ioutil.WriteFile(file, body, 0644)
}

func load() error {
	var per Persistence
	file := ".logger"

	if path := os.Getenv("ERNEST_LOG_CONFIG"); path != "" {
		file = path + file
	}

	if _, err := os.Stat(file); os.IsNotExist(err) {
		return err
	}

	dat, err := ioutil.ReadFile(file)
	if err != nil {
		log.Println("Error reading '" + file + "' file")
		return err
	}
	if err = json.Unmarshal(dat, &per); err != nil {
		log.Println("Persistence file is corrupted")
		return err
	}

	m := nats.Msg{}
	if string(per.Basic) != "" {
		m.Data = per.Basic
		newBasicAdapterListener(&m)
	}
	if string(per.Logstash) != "" {
		m.Data = per.Logstash
		newLogstashAdapterListener(&m)
	}
	if string(per.Rollbar) != "" {
		m.Data = per.Rollbar
		newRollbarAdapterListener(&m)
	}

	return nil
}
