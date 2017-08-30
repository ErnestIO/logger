/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"encoding/json"
	"strings"
	"time"
)

var obfuscation = "[OBFUSCATED]"

// Datacenter holds datacenter passwords
type Datacenter struct {
	Credentials struct {
		Pwd             string `json:"password"`
		AccessKeyID     string `json:"aws_access_key_id"`
		SecretAccessKey string `json:"aws_secret_access_key"`
		SubscriptionID  string `json:"azure_subscription_id"`
		ClientID        string `json:"azure_client_id"`
		ClientSecret    string `json:"azure_client_secret"`
		TenantID        string `json:"azure_tenant_id"`
	} `json:"credentials"`
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
	if len(patternsToObfuscate) > 0 {
		return patternsToObfuscate, nil
	}
	var datacenters []Datacenter

	msg, err := nc.Request("datacenter.find", []byte("{}"), time.Second)
	if err != nil {
		return needles, err
	}

	err = json.Unmarshal(msg.Data, &datacenters)
	if err != nil {
		return needles, err
	}

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
	if d.Credentials.Pwd != "" {
		*needles = append(*needles, d.Credentials.Pwd)
	}
	if d.Credentials.AccessKeyID != "" {
		*needles = append(*needles, d.Credentials.AccessKeyID)
	}
	if d.Credentials.SecretAccessKey != "" {
		*needles = append(*needles, d.Credentials.SecretAccessKey)
	}
	if d.Credentials.SubscriptionID != "" {
		*needles = append(*needles, d.Credentials.SubscriptionID)
	}
	if d.Credentials.ClientID != "" {
		*needles = append(*needles, d.Credentials.ClientID)
	}
	if d.Credentials.ClientSecret != "" {
		*needles = append(*needles, d.Credentials.ClientSecret)
	}
	if d.Credentials.TenantID != "" {
		*needles = append(*needles, d.Credentials.TenantID)
	}
}
