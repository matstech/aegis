package security

import (
	"aegis/configuration"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var signature = "XciMlTpNQSefPAjCbHzHU6fF3YorGGOMyP8qMuYKCOc3Z1MD5iSb9dgUyvg6arCRd/Bz4/EfJRO00HXLZLX1Dw=="
var signatureNoBody = "r2ncXWTsILhGhaDByZUFRrUPxT3nz1pw9qeXd2TdRizH75qq6m5UFoDa31CapZIp2TyTKTs3v6TqZr+8qYdHGQ=="
var authKid = "c0y44e8LL4"
var authHeaders = "header1;header2"
var payload = []byte("DuqjbeoyE9LIo77MaATfF0zl3hu2BZ31")

var headersMap = map[string]string{
	configuration.SIGNATURE:          signature,
	configuration.AUTH_KID:           authKid,
	configuration.AUTH_HEADERS:       authHeaders,
	"header1":                        "header1",
	"header2":                        "header2",
	configuration.AUTH_CORRELATIONID: "1fkEphx2qq",
}

var entities = []string{authKid}

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
	setAccessKeyEnv()
	h := http.Header{}

	for hName, hValue := range hs {
		h.Add(hName, hValue)
	}

	return h
}

func setAccessKeyEnv() {
	os.Setenv("ACCESSKEY_"+strings.ToUpper(authKid), "QTEiL2Jy92")
}
