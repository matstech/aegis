package server

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"strings"
	"tokenguard/configuration"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/shivakar/xxhash"
)

func PostHandler(ctx *gin.Context, entities []configuration.Entity, proxyHost string) {
	log.Info().Msg("PostHandler:: start")

	headers := ctx.Request.Header

	if len(headers) <= 0 {
		panic(errors.New("no heders found"))
	}

	authKid := headers.Get("Auth-Kid")

	if authKid == "" {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("no auth kid"))
	}

	authHeaders := headers.Get("Auth-Headers")

	if authHeaders == "" {
		log.Info().Msgf("no auth headers specified")
	}

	signature := headers.Get("Signature")

	if signature == "" {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("no signature"))
		return
	}

	body, _ := io.ReadAll(ctx.Request.Body)

	vs := verifySignature(signature, authKid, authHeaders, body, headers, entities)

	if !vs {
		ctx.AbortWithError(http.StatusUnauthorized, errors.New("signature cannot be verified"))
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

func verifySignature(signature, authKid, authHeaders string, payload []byte, headers http.Header, entities []configuration.Entity) bool {

	secret := getKidSecret(authKid, entities)

	if secret == "" {
		return false
	}

	h := hmac.New(sha512.New, []byte(secret))

	tbv := ""
	if authHeaders != "" {
		hs := strings.Split(authHeaders, ";")

		for _, h := range hs {
			tbv += fmt.Sprintf("%s;", headers.Get(h))
		}
	}

	if len(payload) > 0 {
		h := xxhash.NewXXHash64()

		h.Write(payload)

		fmt.Println(h.String())

		tbv += fmt.Sprintf(":%s", h.String())
	}

	h.Write([]byte(tbv))

	computed := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return hmac.Equal([]byte(computed), []byte(signature))

}

func getKidSecret(kid string, entities []configuration.Entity) string {
	for _, k := range entities {
		if k.Name == kid {
			return k.Key
		}
	}
	return ""
}
