package configuration

type MainConfiguration struct {
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
