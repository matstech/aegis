// Copyright 2026-present matstech
// SPDX-License-Identifier: GPL-3.0-only

package server

import (
	"aegis/configuration"
	"aegis/security"
	"bytes"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"

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

		upstreamScheme := configuration.PROTOCOL_SCHEME
		upstreamHost := r.conf.Server.Upstream
		if parsedUpstream, err := url.Parse(r.conf.Server.Upstream); err == nil &&
			parsedUpstream.Scheme != "" &&
			parsedUpstream.Host != "" {
			upstreamScheme = parsedUpstream.Scheme
			upstreamHost = parsedUpstream.Host
		}

		req.Header = r.buildUpstreamHeaders(ctx.Request.Header)
		req.Host = upstreamHost
		req.URL.Scheme = upstreamScheme
		req.URL.Host = upstreamHost
		req.URL.Path = ctx.Request.URL.Path
		req.URL.RawPath = ctx.Request.URL.RawPath
		req.URL.RawQuery = ctx.Request.URL.RawQuery

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

func (r *Router) buildUpstreamHeaders(headers http.Header) http.Header {
	upstreamHeaders := headers.Clone()

	upstreamHeaders.Del(configuration.AUTH_KID)
	upstreamHeaders.Del(configuration.AUTH_HEADERS)
	upstreamHeaders.Del(configuration.SIGNATURE)

	for _, headerName := range r.conf.Server.DropHeaders {
		upstreamHeaders.Del(headerName)
	}

	for headerName, headerValue := range r.conf.Server.ResolvedInjectHeaders() {
		upstreamHeaders.Set(headerName, headerValue)
	}

	return upstreamHeaders
}

func checkHeaders(ctx *gin.Context) (string, string, string) {

	headers := ctx.Request.Header

	log.Debug().Msgf("Request headers: %s", headers)

	authCorrelationId := headers.Get(configuration.AUTH_CORRELATIONID)

	if authCorrelationId == "" {
		ParamMissingError(ctx, configuration.AUTH_CORRELATIONID)
		return "", "", ""
	}

	authKid := headers.Get(configuration.AUTH_KID)

	if authKid == "" {
		ParamMissingError(ctx, configuration.AUTH_KID)
		return "", "", ""
	}

	signature := headers.Get(configuration.SIGNATURE)

	if signature == "" {
		ParamMissingError(ctx, configuration.SIGNATURE)
		return "", "", ""
	}

	return authKid,
		headers.Get(configuration.AUTH_HEADERS),
		signature
}
