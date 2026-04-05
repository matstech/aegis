// Copyright 2026-present matstech
// SPDX-License-Identifier: GPL-3.0-only

package main

import (
	"aegis/configuration"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/shivakar/xxhash"
)

type headerFlags []string

func (h *headerFlags) String() string {
	return strings.Join(*h, ",")
}

func (h *headerFlags) Set(value string) error {
	*h = append(*h, value)
	return nil
}

func main() {
	var headers headerFlags

	kid := flag.String("kid", "", "kid used for the request")
	secret := flag.String("secret", "", "shared secret associated to the kid")
	correlationID := flag.String("correlation-id", "", "Auth-CorrelationId header value")
	authHeaders := flag.String("auth-headers", "", "semicolon-separated list of signed headers")
	bodyFile := flag.String("body-file", "", "path to request body file")
	flag.Var(&headers, "header", "request header in 'Name: Value' format")
	flag.Parse()

	if *kid == "" || *secret == "" || *correlationID == "" || *bodyFile == "" {
		fmt.Fprintln(os.Stderr, "kid, secret, correlation-id and body-file are required")
		os.Exit(1)
	}

	body, err := os.ReadFile(*bodyFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot read body file: %v\n", err)
		os.Exit(1)
	}

	httpHeaders := http.Header{}
	httpHeaders.Set(configuration.AUTH_CORRELATIONID, *correlationID)
	httpHeaders.Set(configuration.AUTH_KID, *kid)

	for _, rawHeader := range headers {
		parts := strings.SplitN(rawHeader, ":", 2)
		if len(parts) != 2 {
			fmt.Fprintf(os.Stderr, "invalid header %q, expected 'Name: Value'\n", rawHeader)
			os.Exit(1)
		}

		headerName := strings.TrimSpace(parts[0])
		headerValue := strings.TrimSpace(parts[1])
		httpHeaders.Set(headerName, headerValue)
	}

	toBeVerified := httpHeaders.Get(configuration.AUTH_CORRELATIONID)

	if *authHeaders != "" {
		for _, headerName := range strings.Split(*authHeaders, ";") {
			toBeVerified += fmt.Sprintf(";%s", httpHeaders.Get(strings.TrimSpace(headerName)))
		}
	}

	if len(body) > 0 {
		hash := xxhash.NewXXHash64()
		hash.Write(body)
		toBeVerified += fmt.Sprintf(":%s", hash.String())
	}

	signer := hmac.New(sha512.New, []byte(*secret))
	signer.Write([]byte(toBeVerified))

	fmt.Print(base64.StdEncoding.EncodeToString(signer.Sum(nil)))
}
