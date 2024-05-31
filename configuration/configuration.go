package configuration

import "time"

type MainConfiguration struct {
	Ginmode  string `default:"debug"`
	Loglevel string
	Server   Server `required:"true"`
	Entities []Entity
}

type Server struct {
	Mode string `default:"PLAIN"` //TODO: possible values
	Mtls struct {
		Certpath, Keypath, Cacert string
	}
	Port                  int `default:"8080"`
	Proxy                 string
	Timeout               time.Duration `default:"0"`
	IdleConnectionTimeout time.Duration `default:"0"`
}

type Entity struct {
	Name string
	Key  string
}
