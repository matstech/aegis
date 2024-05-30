package server

import (
	"io"
	"net/http"
	"net/http/httputil"
	"tokenguard/configuration"
	"tokenguard/constants"
	"tokenguard/security"

	"github.com/gin-gonic/gin"
)

func Handler(ctx *gin.Context, entities []configuration.Entity, proxyHost string) {

	authKid, authHeaders, signature := checkHeaders(ctx)

	var body []byte

	if ctx.Request.Body != nil {
		body, _ = io.ReadAll(ctx.Request.Body)
	}

	vs := security.VerifySignature(signature, authKid, authHeaders, body, ctx.Request.Header, entities)

	if !vs {
		InvalidSignature(ctx)
		return
	}

	director := func(req *http.Request) {

		ctx.Request.Header.Del(constants.AUTH_KID)
		ctx.Request.Header.Del(constants.AUTH_HEADERS)
		ctx.Request.Header.Del(constants.SIGNATURE)

		req.Header = ctx.Request.Header
		req.Host = ctx.Request.Host
		req.URL.Scheme = constants.PROTOCOL_SCHEME
		req.URL.Host = proxyHost
		req.URL.Path = ctx.Request.URL.Path

	}

	proxy := &httputil.ReverseProxy{Director: director}

	proxy.ServeHTTP(ctx.Writer, ctx.Request)

}

func checkHeaders(ctx *gin.Context) (string, string, string) {

	headers := ctx.Request.Header

	authCorrelationId := headers.Get(constants.AUTH_CORRELATIONID)

	if authCorrelationId == "" {
		ParamMissingError(ctx, constants.AUTH_CORRELATIONID)
	}

	authKid := headers.Get(constants.AUTH_KID)

	if authKid == "" {
		ParamMissingError(ctx, constants.AUTH_KID)
	}

	signature := headers.Get(constants.SIGNATURE)

	if signature == "" {
		ParamMissingError(ctx, constants.SIGNATURE)
	}

	return authKid,
		headers.Get(constants.AUTH_HEADERS),
		signature
}
