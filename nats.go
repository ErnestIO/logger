/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"encoding/json"

	"github.com/nats-io/go-nats"
)

func natsHandler(msg *nats.Msg) {
	if msg.Subject == "logger.log" {
		return
	}

	body := Obfuscate(msg.Subject, string(msg.Data))

	m := LogMessage{
		Subject: msg.Subject,
		Body:    body,
		Level:   "debug",
		User:    "system",
	}

	data, _ := json.Marshal(m)

	bc.Publish("logs", data)
}
