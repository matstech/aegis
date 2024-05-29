package server

import (
	"errors"
	"io"
	"net/http"
	"net/http/httputil"
	"tokenguard/configuration"
	"tokenguard/security"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func PostHandler(ctx *gin.Context, entities []configuration.Entity, proxyHost string) {
	log.Info().Msg("PostHandler:: start")

	authKid, authHeaders, signature := checkHeaders(ctx)

	body, _ := io.ReadAll(ctx.Request.Body)

	vs := security.VerifySignature(signature, authKid, authHeaders, body, ctx.Request.Header, entities)

	if !vs {
		ctx.AbortWithError(http.StatusUnauthorized, errors.New("signature not valid"))
		return
	}

	director := func(req *http.Request) {

		ctx.Request.Header.Del(authKid)
		ctx.Request.Header.Del(authHeaders)
		ctx.Request.Header.Del(signature)

		req.Header = ctx.Request.Header
		req.Host = ctx.Request.Host
		req.URL.Scheme = "http"
		req.URL.Host = proxyHost
		req.URL.Path = ctx.Request.URL.Path

	}
	proxy := &httputil.ReverseProxy{Director: director}
	proxy.ServeHTTP(ctx.Writer, ctx.Request)

}

func checkHeaders(ctx *gin.Context) (string, string, string) {

	headers := ctx.Request.Header

	if len(headers) <= 0 {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("no headers found"))
	}

	authCorrelationId := headers.Get("Auth-CorrelationId")

	if authCorrelationId == "" {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("missing Auth-CorrelationId"))
	}

	authKid := headers.Get("Auth-Kid")

	if authKid == "" {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("missing Auth-Kid"))
	}

	signature := headers.Get("Signature")

	if signature == "" {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("missing Signature"))

	}

	return authKid, headers.Get("Auth-Headers"), signature
}
