// Copyright 2026-present matstech
// SPDX-License-Identifier: GPL-3.0-only

package main

import (
	"aegis/configuration"
	"aegis/server"
	"os"

	"github.com/cristalhq/aconfig"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	var cfg configuration.MainConfiguration

	configPath := os.Getenv("CONFIG_PATH")

	if configPath == "" {
		log.Fatal().Msgf("Cannot load configuration: no CONFIG_PATH env set")
		os.Exit(1)
	}

	loader := aconfig.LoaderFor(&cfg,
		aconfig.Config{
			EnvPrefix:          "",
			SkipFlags:          true,
			AllowUnknownFields: true,

			Files: []string{configPath + "config.json"},
		})

	if err := loader.Load(); err != nil {
		log.Fatal().Msgf("Error loading configuration: %s", err.Error())
		os.Exit(1)
	}

	logLevel, logErr := zerolog.ParseLevel(cfg.Loglevel)

	if logErr != nil {
		logLevel = zerolog.ErrorLevel
	}

	zerolog.SetGlobalLevel(logLevel)

	log.Info().Msgf("Configuration loaded: %v", cfg)

	if err := cfg.Validate(); err != nil {
		log.Fatal().Msgf("Invalid configuration: %s", err.Error())
		os.Exit(1)
	}

	router := server.NewRouter(&cfg)

	router.Start()
}
