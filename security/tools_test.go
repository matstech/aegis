package security

import (
	"net/http"
	"testing"
	"tokenguard/configuration"

	"github.com/stretchr/testify/assert"
)

var signature = "6w0jio7c3F35jFBhZmk+rhKJTl6DEMLkyrXJlGMciReYqOmZB0furxlZC8++fO8dFsdMvNq7Ngmane9xiSJkzQ=="
var signatureNoBody = "3YeNZUvn0HcoNrnzWRpLzvLjr07025Qy+KGYe2KB83xPKi6RLW+B0A0+eo0aLMBlQOIZwp3o0Nd8NAxxXrGVrg=="
var authKid = "c0y44e8LL4"
var authHeaders = "header1;header2"
var payload = []byte("DuqjbeoyE9LIo77MaATfF0zl3hu2BZ31")

var headersMap = map[string]string{
	"Signature":          signature,
	"Auth-Kid":           authKid,
	"Auth-Headers":       authHeaders,
	"header1":            "header1",
	"header2":            "header2",
	"Auth-CorrelationId": "1fkEphx2qq",
}

var entities = []configuration.Entity{
	{
		Name: "c0y44e8LL4",
		Key:  "QTEiL2Jy92",
	},
}

// func VerifySignature(signature, authKid, authHeaders string, payload []byte, headers http.Header, entities []configuration.Entity) bool {
func TestVerifySignatureOk(t *testing.T) {
	headers := createHttpHeader(headersMap)

	v := VerifySignature(signature, authKid, authHeaders, payload, headers, entities)

	assert.True(t, v)
}

func TestVerifySignatureWrongSignature(t *testing.T) {
	headers := createHttpHeader(headersMap)

	v := VerifySignature("J9VwOWfLz8", authKid, authHeaders, payload, headers, entities)

	assert.False(t, v)
}

func TestVerifySignatureWrongKid(t *testing.T) {
	headers := createHttpHeader(headersMap)

	v := VerifySignature(signature, "wrongKid", authHeaders, payload, headers, entities)

	assert.False(t, v)
}

func TestVerifySignatureNoKid(t *testing.T) {
	headers := createHttpHeader(headersMap)

	v := VerifySignature(signature, "", authHeaders, payload, headers, entities)

	assert.False(t, v)
}

func TestVerifySignatureNoBody(t *testing.T) {
	headers := createHttpHeader(headersMap)

	v := VerifySignature(signatureNoBody, authKid, authHeaders, nil, headers, entities)

	assert.True(t, v)
}

// Test function tools
func createHttpHeader(hs map[string]string) http.Header {
	h := http.Header{}

	for hName, hValue := range hs {
		h.Add(hName, hValue)
	}

	return h
}
