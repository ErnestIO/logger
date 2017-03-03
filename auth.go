/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)

func unauthorized(w http.ResponseWriter) {
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

func extractToken(r *http.Request) (string, error) {
	auth := r.Header.Get("Authorization")
	l := len("Bearer")
	if len(auth) > l+1 && auth[:l] == "Bearer" {
		return auth[l+1:], nil
	}
	return "", errors.New("Invalid Token")
}

func authMiddleware(w http.ResponseWriter, r *http.Request) {
	// Check Auth, Until Proper Auth Service is implemented
	authToken, err := extractToken(r)
	if err != nil {
		unauthorized(w)
		return
	}

	secret = os.Getenv("JWT_SECRET")
	if secret == "" {
		token, err := nc.Request("config.get.jwt_token", []byte(""), 1*time.Second)
		if err != nil {
			panic("Can't get jwt_config config")
		}

		secret = string(token.Data)
	}

	token, err := jwt.Parse(authToken, func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected jwt signing method=%v", t.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		unauthorized(w)
		return
	}

	if token.Valid != true {
		unauthorized(w)
		return
	}

	// Pass to sse server
	s.HTTPHandler(w, r)
}
