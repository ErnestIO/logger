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
	Datacenter      Datacenter  `json:"datacenter"`
	Datacenters     Itemable    `json:"datacenters"`
	Components      []PwdStruct `json:"components"`
	Password        string      `json:"datacenter_password"`
	AccessKeyID     string      `json:"aws_access_key_id"`
	SecretAccessKey string      `json:"aws_secret_access_key"`
	ConfigPassword  string      `json:"password"`
	ConfigToken     string      `json:"datacenter_access_token"`
	ConfigSecret    string      `json:"datacenter_access_key"`
	BasicToken      string      `json:"token"`
	BasicSecret     string      `json:"secret"`
}

// Itemable holds any items for any datacenters
type Itemable struct {
	Items []Datacenter `json:"items"`
}

// PwdStruct holds datacenter passwords for other items
type PwdStruct struct {
	Pwd    string `json:"datacenter_password"`
	Token  string `json:"aws_access_key_id"`
	Secret string `json:"aws_secret_access_key"`
}

// ServiceSet : ...
type ServiceSet struct {
	Message string `json:"mapping"`
}

// PreProcess gets the password and replaces it before writing to a log
func PreProcess(s string) string {
	// Password
	for _, pwd := range getPasswords(s) {
		s = strings.Replace(s, pwd, obfuscation, -1)
	}

	for _, mappingPassword := range getSeedFromMapping(s, getPasswords) {
		s = strings.Replace(s, mappingPassword, obfuscation, -1)
	}

	for _, l := range getSeedFromList(s, getPasswords) {
		s = strings.Replace(s, l, obfuscation, -1)
	}

	// Token
	for _, token := range getTokens(s) {
		s = strings.Replace(s, token, obfuscation, -1)
	}
	for _, mappingToken := range getSeedFromMapping(s, getTokens) {
		s = strings.Replace(s, mappingToken, obfuscation, -1)
	}
	for _, l := range getSeedFromList(s, getTokens) {
		s = strings.Replace(s, l, obfuscation, -1)
	}

	// Secret
	for _, secret := range getSecrets(s) {
		s = strings.Replace(s, secret, obfuscation, -1)
	}
	for _, mappingSecret := range getSeedFromMapping(s, getSecrets) {
		s = strings.Replace(s, mappingSecret, obfuscation, -1)
	}
	for _, l := range getSeedFromList(s, getSecrets) {
		s = strings.Replace(s, l, obfuscation, -1)
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
		if d.AccessKeyID != "" {
			pwds = append(pwds, d.AccessKeyID)
		}
	}
	if m.AccessKeyID != "" {
		pwds = append(pwds, m.AccessKeyID)
	}
	if m.ConfigToken != "" {
		pwds = append(pwds, m.ConfigToken)
	}
	if m.BasicToken != "" {
		pwds = append(pwds, m.BasicToken)
	}

	if m.Datacenter.AccessKeyID != "" {
		pwds = append(pwds, m.Datacenter.AccessKeyID)
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
		if d.SecretAccessKey != "" {
			pwds = append(pwds, d.SecretAccessKey)
		}
	}
	if m.SecretAccessKey != "" {
		pwds = append(pwds, m.SecretAccessKey)
	}
	if m.ConfigSecret != "" {
		pwds = append(pwds, m.ConfigSecret)
	}
	if m.BasicSecret != "" {
		pwds = append(pwds, m.BasicSecret)
	}
	if m.Datacenter.SecretAccessKey != "" {
		pwds = append(pwds, m.Datacenter.SecretAccessKey)
	}

	return pwds
}

type getSeed func(string) []string

func getSeedFromMapping(s string, fn getSeed) []string {
	m := ServiceSet{}
	_ = json.Unmarshal([]byte(s), &m)
	message := strings.Replace(m.Message, "\\\"", "\"", -1)

	return fn(message)
}

func getSeedFromList(s string, fn getSeed) []string {
	m := []Datacenter{}
	_ = json.Unmarshal([]byte(s), &m)
	if len(m) == 0 {
		return []string{}
	}
	message, _ := json.Marshal(m[0])

	return fn(string(message))
}
