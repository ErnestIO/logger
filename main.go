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
	Datacenter     Datacenter  `json:"datacenter"`
	Datacenters    Itemable    `json:"datacenters"`
	Components     []PwdStruct `json:"components"`
	Password       string      `json:"datacenter_password"`
	Token          string      `json:"datacenter_token"`
	Secret         string      `json:"datacenter_secret"`
	ConfigPassword string      `json:"password"`
	ConfigToken    string      `json:"datacenter_access_token"`
	ConfigSecret   string      `json:"datacenter_access_key"`
	BasicToken     string      `json:"token"`
	BasicSecret    string      `json:"secret"`
}

// Itemable holds any items for any datacenters
type Itemable struct {
	Items []Datacenter `json:"items"`
}

// Datacenter holds datacenter passwords
type Datacenter struct {
	Pwd    string `json:"password"`
	Token  string `json:"token"`
	Secret string `json:"secret"`
}

// PwdStruct holds datacenter passwords for other items
type PwdStruct struct {
	Pwd    string `json:"datacenter_password"`
	Token  string `json:"datacenter_token"`
	Secret string `json:"datacenter_secret"`
}

type ServiceSet struct {
	Message string `json:"mapping"`
}

// PreProcess gets the password and replaces it before writing to a log
func PreProcess(s string) string {
	// Password
	if pwd := getPassword(s); pwd != "" {
		s = strings.Replace(s, pwd, "***", -1)
	} else if mappingPassword := getSeedFromMapping(s, getPassword); mappingPassword != "" {
		s = strings.Replace(s, mappingPassword, "***", -1)
	} else if l := getSeedFromList(s, getPassword); l != "" {
		s = strings.Replace(s, l, "***", -1)
	}

	// Token
	if token := getToken(s); token != "" {
		s = strings.Replace(s, token, "***", -1)
	} else if mappingToken := getSeedFromMapping(s, getToken); mappingToken != "" {
		s = strings.Replace(s, mappingToken, "***", -1)
	} else if l := getSeedFromList(s, getToken); l != "" {
		s = strings.Replace(s, l, "***", -1)
	}

	// Secret
	if secret := getSecret(s); secret != "" {
		s = strings.Replace(s, secret, "***", -1)
	} else if mappingSecret := getSeedFromMapping(s, getSecret); mappingSecret != "" {
		s = strings.Replace(s, mappingSecret, "***", -1)
	} else if l := getSeedFromList(s, getSecret); l != "" {
		s = strings.Replace(s, l, "***", -1)
	}

	return s
}

func getPassword(s string) string {
	m := Message{}
	json.Unmarshal([]byte(s), &m)
	if len(m.Components) > 0 {
		if m.Components[0].Pwd != "" {
			return m.Components[0].Pwd
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
	if m.ConfigPassword != "" {
		return m.ConfigPassword
	}

	return m.Datacenter.Pwd
}

func getToken(s string) string {
	m := Message{}
	json.Unmarshal([]byte(s), &m)
	if len(m.Components) > 0 {
		if m.Components[0].Token != "" {
			return m.Components[0].Token
		}
	}
	if len(m.Datacenters.Items) > 0 {
		if m.Datacenters.Items[0].Token != "" {
			return m.Datacenters.Items[0].Token
		}
	}
	if m.Token != "" {
		return m.Token
	}
	if m.ConfigToken != "" {
		return m.ConfigToken
	}
	if m.BasicToken != "" {
		return m.BasicToken
	}

	return m.Datacenter.Token
}

func getSecret(s string) string {
	m := Message{}
	json.Unmarshal([]byte(s), &m)
	if len(m.Components) > 0 {
		if m.Components[0].Secret != "" {
			return m.Components[0].Secret
		}
	}
	if len(m.Datacenters.Items) > 0 {
		if m.Datacenters.Items[0].Secret != "" {
			return m.Datacenters.Items[0].Secret
		}
	}
	if m.Secret != "" {
		return m.Secret
	}
	if m.ConfigSecret != "" {
		return m.ConfigSecret
	}
	if m.BasicSecret != "" {
		return m.BasicSecret
	}

	return m.Datacenter.Secret
}

type getSeed func(string) string

func getSeedFromMapping(s string, fn getSeed) string {
	m := ServiceSet{}
	json.Unmarshal([]byte(s), &m)
	message := strings.Replace(m.Message, "\\\"", "\"", -1)

	return fn(message)
}

func getSeedFromList(s string, fn getSeed) string {
	m := []Datacenter{}
	json.Unmarshal([]byte(s), &m)
	if len(m) == 0 {
		return ""
	}
	message, _ := json.Marshal(m[0])

	return fn(string(message))
}
