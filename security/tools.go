package security

import (
	"aegis/configuration"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/shivakar/xxhash"
)

func VerifySignature(signature, authKid, authHeaders string, payload []byte, headers http.Header, entities []string) bool {

	secret := getKidSecret(authKid, entities)

	if secret == "" {
		log.Warn().Msgf("No accesskey available for kid %s", authKid)
		return false
	}

	h := hmac.New(sha512.New, []byte(secret))

	tbv := headers.Get(configuration.AUTH_CORRELATIONID)

	if authHeaders != "" {
		hs := strings.Split(authHeaders, ";")

		for _, h := range hs {
			tbv += fmt.Sprintf(";%s", headers.Get(h))
		}
	}

	if len(payload) > 0 {
		h := xxhash.NewXXHash64()

		h.Write(payload)

		tbv += fmt.Sprintf(":%s", h.String())
	}

	log.Debug().Msgf("Sign to verify: %s", tbv)

	h.Write([]byte(tbv))

	computed := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return hmac.Equal([]byte(computed), []byte(signature))

}

func getKidSecret(kid string, entities []string) string {
	for _, k := range entities {
		if k == kid {
			return os.Getenv(fmt.Sprintf("ACCESSKEY_%s", strings.ToUpper(kid)))
		}
	}
	return ""
}
