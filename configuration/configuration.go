package configuration

import (
	"fmt"
	"os"
	"time"
)

type MainConfiguration struct {
	Ginmode  string `default:"debug" usage:"release in production"`
	Loglevel string
	Server   Server `required:"true"`
	Kids     []string
}

func (c *MainConfiguration) Validate() error {
	return c.Server.Validate()
}

type HeaderValueSource struct {
	Value        string
	ValueFromEnv string
}

func (h HeaderValueSource) Resolve(headerName string) (string, error) {
	hasValue := h.Value != ""
	hasValueFromEnv := h.ValueFromEnv != ""

	switch {
	case hasValue == hasValueFromEnv:
		return "", fmt.Errorf("exactly one of value or valueFromEnv must be configured for header %s", headerName)
	case hasValue:
		return h.Value, nil
	default:
		value, found := os.LookupEnv(h.ValueFromEnv)
		if !found {
			return "", fmt.Errorf("environment variable %q not found", h.ValueFromEnv)
		}

		return value, nil
	}
}

type Server struct {
	Mode string `default:"PLAIN" usage:"PLAIN,TLS,MTLS"`
	Tls  struct {
		Certpath, Keypath, Cacert string
	}
	Port                  int `default:"8080"`
	ProbesPort            int `default:"2112"`
	Upstream              string
	DropHeaders           []string
	InjectHeaders         map[string]HeaderValueSource
	Timeout               time.Duration `default:"0"`
	IdleConnectionTimeout time.Duration `default:"0"`
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

	for headerName, headerValueSource := range s.InjectHeaders {
		value, err := headerValueSource.Resolve(headerName)
		if err != nil {
			return nil, fmt.Errorf("invalid injectHeaders entry for %q: %w", headerName, err)
		}

		resolved[headerName] = value
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
