package server

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"tokenguard/constants"
	"tokenguard/security"

	"github.com/gin-gonic/gin"
)

func (r *Router) Handler(ctx *gin.Context) {

	authKid, authHeaders, signature := checkHeaders(ctx)

	var body []byte

	if ctx.Request.Body != nil {
		body, _ = io.ReadAll(ctx.Request.Body)
	}

	vs := security.VerifySignature(signature, authKid, authHeaders, body, ctx.Request.Header, r.conf.Entities)

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
		req.URL.Host = r.conf.Server.Proxy
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
