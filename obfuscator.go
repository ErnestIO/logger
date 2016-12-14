/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"encoding/json"
	"log"
	"strings"
	"time"
)

var obfuscation = "[OBFUSCATED]"

// Datacenter holds datacenter passwords
type Datacenter struct {
	Pwd             string `json:"password"`
	AccessKeyID     string `json:"aws_access_key_id"`
	SecretAccessKey string `json:"aws_secret_access_key"`
}

// Obfuscate : obfuscates sensible data on the given stack
func Obfuscate(stack string) string {
	stack = PreProcess(stack)
	if needles, err := getNeedles(); err != nil {
		stack = "[ An error occurred trying to obfuscate this message ]"
	} else {
		for _, needle := range needles {
			if needle != "" {
				stack = strings.Replace(stack, needle, obfuscation, -1)
			}
		}
	}
	return stack
}

func getNeedles() (needles []string, err error) {
	log.Println(len(patternsToObfuscate))
	if len(patternsToObfuscate) > 0 {
		return patternsToObfuscate, nil
	}
	var datacenters []Datacenter

	msg, err := nc.Request("datacenter.find", []byte("{}"), time.Second)
	if err != nil {
		return needles, err
	}
	_ = json.Unmarshal(msg.Data, &datacenters)
	if len(datacenters) == 0 {
		needles = append(needles, "")
	} else {
		for _, d := range datacenters {
			addDatacenterPatterns(d, &needles)
		}
	}
	patternsToObfuscate = needles

	return needles, nil
}

func addDatacenterPatterns(d Datacenter, needles *[]string) {
	if d.Pwd != "" {
		*needles = append(*needles, d.Pwd)
	}
	if d.AccessKeyID != "" {
		*needles = append(*needles, d.AccessKeyID)
	}
	if d.SecretAccessKey != "" {
		*needles = append(*needles, d.SecretAccessKey)
	}
}
