package main

import (
	"tokenguard/configuration"
	"tokenguard/server"

	"github.com/cristalhq/aconfig"
	"github.com/rs/zerolog/log"
)

func main() {
	var cfg configuration.MainConfiguration

	loader := aconfig.LoaderFor(&cfg, aconfig.Config{
		EnvPrefix:  "APP",
		FlagPrefix: "app",
		Files:      []string{"/home/mat/workspace/personal/go-token-guard/config.json"},
	})

	if err := loader.Load(); err != nil {
		panic(err)
	}

	log.Info().Msgf("Configuration loaded: %v", cfg)

	router := server.NewRouter(&cfg)

	router.Start()
}
