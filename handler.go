/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	//
	"errors"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/r3labs/broadcast"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		upgradefail(w)
		return
	}

	var authorized bool
	var ch chan *broadcast.Event
	var sub *broadcast.Subscriber

	defer func() {
		_ = c.Close()

		if ch != nil && sub != nil {
			sub.Disconnect(ch)
		}
	}()

	for {
		if !authorized {
			areq, err := authenticate(w, c)
			if err != nil {
				return
			}

			sub, ch, err = register(w, areq)
			if err != nil {
				return
			}

			authorized = true
		} else {
			msg, ok := <-ch
			if !ok {
				return
			}

			err := c.WriteMessage(websocket.TextMessage, msg.Data)
			if err != nil {
				_ = internalerror(w)
				return
			}
		}
	}
}

func register(w http.ResponseWriter, s *Session) (*broadcast.Subscriber, chan *broadcast.Event, error) {
	sub := bc.GetSubscriber(s.Username)
	if sub == nil {
		sub = broadcast.NewSubscriber(s.Username)
		bc.Register("logs", sub)
	}

	return sub, sub.Connect(), nil
}

func upgradefail(w http.ResponseWriter) {
	http.Error(w, "Unable to upgrade to websocket connection", http.StatusBadRequest)
}

func badrequest(w http.ResponseWriter) error {
	http.Error(w, "Could not process sent data", http.StatusBadRequest)
	return errors.New("Could not process sent data")
}

func badstream(w http.ResponseWriter) error {
	http.Error(w, "Please specify a valid stream", http.StatusBadRequest)
	return errors.New("Please specify a valid stream")
}

func internalerror(w http.ResponseWriter) error {
	http.Error(w, "Internal server error", http.StatusInternalServerError)
	return errors.New("Internal server error")
}
