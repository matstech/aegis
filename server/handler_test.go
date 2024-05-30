package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"tokenguard/configuration"
	"tokenguard/constants"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var payload = []byte("DuqjbeoyE9LIo77MaATfF0zl3hu2BZ31")

var headersMap = map[string]string{
	constants.SIGNATURE:          "XciMlTpNQSefPAjCbHzHU6fF3YorGGOMyP8qMuYKCOc3Z1MD5iSb9dgUyvg6arCRd/Bz4/EfJRO00HXLZLX1Dw==",
	constants.AUTH_KID:           "c0y44e8LL4",
	constants.AUTH_HEADERS:       "header1;header2",
	"header1":                    "header1",
	"header2":                    "header2",
	constants.AUTH_CORRELATIONID: "1fkEphx2qq",
}

var entities = []configuration.Entity{
	{
		Name: "c0y44e8LL4",
		Key:  "QTEiL2Jy92",
	},
}

func TestHandlerOk(t *testing.T) {
	serverUrl, server := mockProxyServer(false)

	defer server.Close()

	ctx := buildGinContext(serverUrl)

	for hName, hValue := range headersMap {
		ctx.Request.Header.Add(hName, hValue)
	}

	ctx.Request.Body = io.NopCloser(strings.NewReader(string(payload)))

	Handler(ctx, entities, ctx.Request.URL.Host)

	assert.Equal(t, http.StatusOK, ctx.Writer.Status())

}

func TestHandlerKoSignature(t *testing.T) {
	serverUrl, server := mockProxyServer(true)

	defer server.Close()

	ctx := buildGinContext(serverUrl)

	for hName, hValue := range headersMap {
		if hName == constants.SIGNATURE {
			ctx.Request.Header.Add(hName, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
		} else {
			ctx.Request.Header.Add(hName, hValue)
		}
	}

	ctx.Request.Body = io.NopCloser(strings.NewReader(string(payload)))

	Handler(ctx, entities, ctx.Request.URL.Host)

	assert.Equal(t, http.StatusBadRequest, ctx.Writer.Status())
}

func TestHandlerKoFromServer(t *testing.T) {
	serverUrl, server := mockProxyServer(true)

	defer server.Close()

	ctx := buildGinContext(serverUrl)

	for hName, hValue := range headersMap {
		ctx.Request.Header.Add(hName, hValue)
	}

	ctx.Request.Body = io.NopCloser(strings.NewReader(string(payload)))

	Handler(ctx, entities, ctx.Request.URL.Host)

	assert.Equal(t, http.StatusFailedDependency, ctx.Writer.Status())
}

func TestHandlerNoHeaders(t *testing.T) {

	testHost, server := mockProxyServer(false)

	defer server.Close()

	ctx := buildGinContext(testHost)

	Handler(ctx, entities, ctx.Request.URL.Host)

	assert.Equal(t, http.StatusBadRequest, ctx.Writer.Status())
}

// tools
func buildGinContext(testHost string) *gin.Context {

	gin.SetMode(gin.TestMode)

	w := CreateTestResponseRecorder(testHost)

	c, _ := gin.CreateTestContext(w)

	h := strings.Split(testHost, "http://")[1]

	c.Request = &http.Request{
		Method: "POST",
		URL:    &url.URL{Host: h},
		Header: make(http.Header),
		Body:   http.NoBody,
	}

	return c
}

func mockProxyServer(error bool) (string, *httptest.Server) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//no action required
		if error {
			w.WriteHeader(http.StatusFailedDependency)
		}

		if r.Header.Get(constants.SIGNATURE) != "" ||
			r.Header.Get(constants.AUTH_KID) != "" ||
			r.Header.Get(constants.AUTH_HEADERS) != "" ||
			r.Header.Get(constants.AUTH_CORRELATIONID) == "" {
			w.WriteHeader(http.StatusNotAcceptable)
		}
	}))

	return server.URL, server
}

type TestResponseRecorder struct {
	*httptest.ResponseRecorder
	closeChannel chan bool
}

func (r *TestResponseRecorder) CloseNotify() <-chan bool {
	return r.closeChannel
}

func CreateTestResponseRecorder(testHost string) *TestResponseRecorder {

	recorder := httptest.NewRecorder()

	return &TestResponseRecorder{
		recorder,
		make(chan bool, 1),
	}
}
