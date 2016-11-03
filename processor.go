/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"encoding/json"
	"strings"
)

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
	for _, pwd := range getPasswords(s) {
		s = strings.Replace(s, pwd, "***", -1)
	}

	for _, mappingPassword := range getSeedFromMapping(s, getPasswords) {
		s = strings.Replace(s, mappingPassword, "***", -1)
	}

	for _, l := range getSeedFromList(s, getPasswords) {
		s = strings.Replace(s, l, "***", -1)
	}

	// Token
	for _, token := range getTokens(s) {
		s = strings.Replace(s, token, "***", -1)
	}
	for _, mappingToken := range getSeedFromMapping(s, getTokens) {
		s = strings.Replace(s, mappingToken, "***", -1)
	}
	for _, l := range getSeedFromList(s, getTokens) {
		s = strings.Replace(s, l, "***", -1)
	}

	// Secret
	for _, secret := range getSecrets(s) {
		s = strings.Replace(s, secret, "***", -1)
	}
	for _, mappingSecret := range getSeedFromMapping(s, getSecrets) {
		s = strings.Replace(s, mappingSecret, "***", -1)
	}
	for _, l := range getSeedFromList(s, getSecrets) {
		s = strings.Replace(s, l, "***", -1)
	}

	return s
}

func getPasswords(s string) []string {
	var pwds []string

	var m Message
	err := json.Unmarshal([]byte(s), &m)
	if err == nil {
		return processPasswords(&m)
	}

	var cm []Message
	err = json.Unmarshal([]byte(s), &cm)
	if err != nil {
		return pwds
	}

	for _, m := range cm {
		pwds = append(pwds, processPasswords(&m)...)
	}

	return pwds
}

func processPasswords(m *Message) []string {
	var pwds []string
	for _, c := range m.Components {
		if c.Pwd != "" {
			pwds = append(pwds, c.Pwd)
		}
	}

	for _, d := range m.Datacenters.Items {
		if d.Pwd != "" {
			pwds = append(pwds, d.Pwd)
		}
	}

	if m.Password != "" {
		pwds = append(pwds, m.Password)
	}
	if m.ConfigPassword != "" {
		pwds = append(pwds, m.ConfigPassword)
	}
	if m.Datacenter.Pwd != "" {
		pwds = append(pwds, m.Datacenter.Pwd)
	}

	return pwds
}

func getTokens(s string) []string {
	var pwds []string

	var m Message
	err := json.Unmarshal([]byte(s), &m)
	if err == nil {
		return processTokens(&m)
	}

	var cm []Message
	err = json.Unmarshal([]byte(s), &cm)
	if err != nil {
		return pwds
	}

	for _, m := range cm {
		pwds = append(pwds, processTokens(&m)...)
	}

	return pwds
}

func processTokens(m *Message) []string {
	var pwds []string

	for _, c := range m.Components {
		if c.Token != "" {
			pwds = append(pwds, c.Token)
		}
	}
	for _, d := range m.Datacenters.Items {
		if d.Token != "" {
			pwds = append(pwds, d.Token)
		}
	}
	if m.Token != "" {
		pwds = append(pwds, m.Token)
	}
	if m.ConfigToken != "" {
		pwds = append(pwds, m.ConfigToken)
	}
	if m.BasicToken != "" {
		pwds = append(pwds, m.BasicToken)
	}

	if m.Datacenter.Token != "" {
		pwds = append(pwds, m.Datacenter.Token)
	}

	return pwds
}

func getSecrets(s string) []string {
	var pwds []string

	var m Message
	err := json.Unmarshal([]byte(s), &m)
	if err == nil {
		return processSecrets(&m)
	}

	var cm []Message
	err = json.Unmarshal([]byte(s), &cm)
	if err != nil {
		return pwds
	}

	for _, m := range cm {
		pwds = append(pwds, processSecrets(&m)...)
	}

	return pwds
}

func processSecrets(m *Message) []string {
	var pwds []string

	for _, c := range m.Components {
		if c.Secret != "" {
			pwds = append(pwds, c.Secret)
		}
	}
	for _, d := range m.Datacenters.Items {
		if d.Secret != "" {
			pwds = append(pwds, d.Secret)
		}
	}
	if m.Secret != "" {
		pwds = append(pwds, m.Secret)
	}
	if m.ConfigSecret != "" {
		pwds = append(pwds, m.ConfigSecret)
	}
	if m.BasicSecret != "" {
		pwds = append(pwds, m.BasicSecret)
	}
	if m.Datacenter.Secret != "" {
		pwds = append(pwds, m.Datacenter.Secret)
	}

	return pwds
}

type getSeed func(string) []string

func getSeedFromMapping(s string, fn getSeed) []string {
	m := ServiceSet{}
	json.Unmarshal([]byte(s), &m)
	message := strings.Replace(m.Message, "\\\"", "\"", -1)

	return fn(message)
}

func getSeedFromList(s string, fn getSeed) []string {
	m := []Datacenter{}
	json.Unmarshal([]byte(s), &m)
	if len(m) == 0 {
		return []string{}
	}
	message, _ := json.Marshal(m[0])

	return fn(string(message))
}
