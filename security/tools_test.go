package security

import (
	"net/http"
	"testing"
	"tokenguard/configuration"

	"github.com/stretchr/testify/assert"
)

var signature = "XciMlTpNQSefPAjCbHzHU6fF3YorGGOMyP8qMuYKCOc3Z1MD5iSb9dgUyvg6arCRd/Bz4/EfJRO00HXLZLX1Dw=="
var signatureNoBody = "r2ncXWTsILhGhaDByZUFRrUPxT3nz1pw9qeXd2TdRizH75qq6m5UFoDa31CapZIp2TyTKTs3v6TqZr+8qYdHGQ=="
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
