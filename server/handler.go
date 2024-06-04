package server

import (
	"aegis/configuration"
	"aegis/security"
	"bytes"
	"io"
	"net"
	"net/http"
	"net/http/httputil"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func (r *Router) Handler(ctx *gin.Context) {

	authKid, authHeaders, signature := checkHeaders(ctx)

	var body []byte

	if ctx.Request.Body != nil {
		body, _ = io.ReadAll(ctx.Request.Body)
	}

	vs := security.VerifySignature(signature, authKid, authHeaders, body, ctx.Request.Header, r.conf.Kids)

	if !vs {
		InvalidSignature(ctx)
		return
	}

	director := func(req *http.Request) {

		ctx.Request.Header.Del(configuration.AUTH_KID)
		ctx.Request.Header.Del(configuration.AUTH_HEADERS)
		ctx.Request.Header.Del(configuration.SIGNATURE)

		req.Header = ctx.Request.Header
		req.Host = ctx.Request.Host
		req.URL.Scheme = configuration.PROTOCOL_SCHEME
		req.URL.Host = r.conf.Server.Upstream
		req.URL.Path = ctx.Request.URL.Path

		if len(body) > 0 {
			req.Body = io.NopCloser(bytes.NewBuffer(body))
		}

	}

	proxy := &httputil.ReverseProxy{Director: director,
		Transport: &http.Transport{

			DialContext: (&net.Dialer{
				Timeout:   r.conf.Server.Timeout,
				KeepAlive: -1,
				DualStack: true,
			}).DialContext,
			IdleConnTimeout: r.conf.Server.IdleConnectionTimeout,
		}}

	proxy.ServeHTTP(ctx.Writer, ctx.Request)

}

func checkHeaders(ctx *gin.Context) (string, string, string) {

	headers := ctx.Request.Header

	log.Debug().Msgf("Request headers: %s", headers)

	authCorrelationId := headers.Get(configuration.AUTH_CORRELATIONID)

	if authCorrelationId == "" {
		ParamMissingError(ctx, configuration.AUTH_CORRELATIONID)
	}

	authKid := headers.Get(configuration.AUTH_KID)

	if authKid == "" {
		ParamMissingError(ctx, configuration.AUTH_KID)
	}

	signature := headers.Get(configuration.SIGNATURE)

	if signature == "" {
		ParamMissingError(ctx, configuration.SIGNATURE)
	}

	return authKid,
		headers.Get(configuration.AUTH_HEADERS),
		signature
}
