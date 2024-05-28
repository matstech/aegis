package security

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"tokenguard/configuration"

	"github.com/shivakar/xxhash"
)

func VerifySignature(signature, authKid, authHeaders string, payload []byte, headers http.Header, entities []configuration.Entity) bool {

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
