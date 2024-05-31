package configuration

import "time"

type MainConfiguration struct {
	Ginmode  string `default:"debug" usage:"release in production"`
	Loglevel string
	Server   Server `required:"true"`
	Kids     []string
}

type Server struct {
	Mode string `default:"PLAIN" usage:"PLAIN,TLS,MTLS"`
	Tls  struct {
		Certpath, Keypath, Cacert string
	}
	Port                  int `default:"8080"`
	Upstream              string
	Timeout               time.Duration `default:"0"`
	IdleConnectionTimeout time.Duration `default:"0"`
}
