/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/websocket"
)

// Session : stores authentication data
type Session struct {
	Token         string  `json:"token"`
	Stream        *string `json:"stream"`
	EventID       *string `json:"event_id"`
	Username      string
	Authenticated bool
}

func unauthorized(w http.ResponseWriter) error {
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
	return errors.New("Unauthorized")
}

func authenticate(w http.ResponseWriter, c *websocket.Conn) (*Session, error) {
	var s Session

	mt, message, err := c.ReadMessage()
	if err != nil {
		return nil, badrequest(w)
	}

	err = json.Unmarshal(message, &s)
	if err != nil {
		return nil, badrequest(w)
	}

	token, err := jwt.Parse(s.Token, func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected jwt signing method=%v", t.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		_ = c.WriteMessage(mt, []byte(`{"status": "unauthorized"}`))
		return nil, unauthorized(w)
	}

	s.Authenticated = true
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok {
		s.Username = claims["username"].(string)
	}

	err = c.WriteMessage(mt, []byte(`{"status": "ok"}`))
	if err != nil {
		return nil, internalerror(w)
	}

	return &s, nil
}
