package configuration

type MainConfiguration struct {
	Ginmode  string `default:"debug"`
	Loglevel string
	Server   Server `required:"true"`
	Entities []Entity
}

type Server struct {
	Port  int `default:"8080"`
	Proxy string
}

type Entity struct {
	Name string
	Key  string
}
