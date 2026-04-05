package configuration

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type MainConfiguration struct {
	Ginmode  string   `json:"ginmode" default:"debug" usage:"release in production"`
	Loglevel string   `json:"loglevel"`
	Server   Server   `json:"server" required:"true"`
	Kids     []string `json:"kids"`
}

func (c *MainConfiguration) Validate() error {
	return c.Server.Validate()
}

type InjectHeader struct {
	Name string `json:"name"`

	Value        string `json:"value"`
	ValueFromEnv string `json:"valueFromEnv"`
}

func (h InjectHeader) Resolve() (string, string, error) {
	if strings.TrimSpace(h.Name) == "" {
		return "", "", fmt.Errorf("injectHeaders entry name must be configured")
	}

	hasValue := h.Value != ""
	hasValueFromEnv := h.ValueFromEnv != ""

	switch {
	case hasValue == hasValueFromEnv:
		return "", "", fmt.Errorf("exactly one of value or valueFromEnv must be configured for header %s", h.Name)
	case hasValue:
		return h.Name, h.Value, nil
	default:
		value, found := os.LookupEnv(h.ValueFromEnv)
		if !found {
			return "", "", fmt.Errorf("environment variable %q not found", h.ValueFromEnv)
		}

		return h.Name, value, nil
	}
}

type Server struct {
	Mode string `json:"mode" default:"PLAIN" usage:"PLAIN,TLS,MTLS"`
	Tls  struct {
		Certpath string `json:"certpath"`
		Keypath  string `json:"keypath"`
		Cacert   string `json:"cacert"`
	} `json:"tls"`
	Port                  int            `json:"port" default:"8080"`
	ProbesPort            int            `json:"probesport" default:"2112"`
	Upstream              string         `json:"upstream"`
	DropHeaders           []string       `json:"dropHeaders"`
	InjectHeaders         []InjectHeader `json:"injectHeaders"`
	Timeout               time.Duration  `json:"timeout" default:"0"`
	IdleConnectionTimeout time.Duration  `json:"idleConnectionTimeout" default:"0"`
	resolvedInjectHeaders map[string]string
}

func (s *Server) Validate() error {
	resolved, err := s.ResolveInjectHeaders()
	if err != nil {
		return err
	}

	s.resolvedInjectHeaders = resolved

	return nil
}

func (s Server) ResolveInjectHeaders() (map[string]string, error) {
	resolved := make(map[string]string, len(s.InjectHeaders))

	for _, injectHeader := range s.InjectHeaders {
		headerName, value, err := injectHeader.Resolve()
		if err != nil {
			return nil, err
		}

		normalizedHeaderName := httpHeaderKey(headerName)
		if _, found := resolved[normalizedHeaderName]; found {
			return nil, fmt.Errorf("duplicate injectHeaders entry for header %s", headerName)
		}

		resolved[normalizedHeaderName] = value
	}

	return resolved, nil
}

func (s Server) ResolvedInjectHeaders() map[string]string {
	resolved := make(map[string]string, len(s.resolvedInjectHeaders))

	for headerName, headerValue := range s.resolvedInjectHeaders {
		resolved[headerName] = headerValue
	}

	return resolved
}

func httpHeaderKey(headerName string) string {
	return strings.TrimSpace(strings.ToLower(headerName))
}
