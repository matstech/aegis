// Copyright 2026-present matstech
// SPDX-License-Identifier: GPL-3.0-only

package server

import (
	"aegis/configuration"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildUpstreamHeadersRemovesAegisAuthHeaders(t *testing.T) {
	cfg := &configuration.MainConfiguration{
		Server: configuration.Server{},
	}
	require.NoError(t, cfg.Validate())
	router := NewRouter(cfg)

	headers := http.Header{}
	headers.Set(configuration.AUTH_CORRELATIONID, "corr-1")
	headers.Set(configuration.AUTH_KID, "kid-1")
	headers.Set(configuration.AUTH_HEADERS, "header1")
	headers.Set(configuration.SIGNATURE, "signature")
	headers.Set("X-Keep", "keep")

	upstreamHeaders := router.buildUpstreamHeaders(headers)

	assert.Equal(t, "corr-1", upstreamHeaders.Get(configuration.AUTH_CORRELATIONID))
	assert.Equal(t, "keep", upstreamHeaders.Get("X-Keep"))
	assert.Empty(t, upstreamHeaders.Get(configuration.AUTH_KID))
	assert.Empty(t, upstreamHeaders.Get(configuration.AUTH_HEADERS))
	assert.Empty(t, upstreamHeaders.Get(configuration.SIGNATURE))
}

func TestBuildUpstreamHeadersAppliesDropAndInjectWithoutMutatingSource(t *testing.T) {
	cfg := &configuration.MainConfiguration{
		Server: configuration.Server{
			DropHeaders: []string{"Authorization", "X-Drop-Me"},
			InjectHeaders: []configuration.InjectHeader{
				{Name: "Authorization", Value: "Bearer upstream-token"},
				{Name: "X-Aegis-Proxy", Value: "true"},
			},
		},
	}
	require.NoError(t, cfg.Validate())
	router := NewRouter(cfg)

	headers := http.Header{}
	headers.Set(configuration.AUTH_CORRELATIONID, "corr-1")
	headers.Set(configuration.AUTH_KID, "kid-1")
	headers.Set(configuration.AUTH_HEADERS, "header1")
	headers.Set(configuration.SIGNATURE, "signature")
	headers.Set("Authorization", "Bearer incoming-token")
	headers.Set("X-Drop-Me", "drop")
	headers.Set("X-Keep", "keep")

	upstreamHeaders := router.buildUpstreamHeaders(headers)

	assert.Equal(t, "corr-1", upstreamHeaders.Get(configuration.AUTH_CORRELATIONID))
	assert.Equal(t, "keep", upstreamHeaders.Get("X-Keep"))
	assert.Equal(t, "Bearer upstream-token", upstreamHeaders.Get("Authorization"))
	assert.Equal(t, "true", upstreamHeaders.Get("X-Aegis-Proxy"))
	assert.Empty(t, upstreamHeaders.Get("X-Drop-Me"))
	assert.Empty(t, upstreamHeaders.Get(configuration.AUTH_KID))
	assert.Empty(t, upstreamHeaders.Get(configuration.AUTH_HEADERS))
	assert.Empty(t, upstreamHeaders.Get(configuration.SIGNATURE))

	assert.Equal(t, "kid-1", headers.Get(configuration.AUTH_KID))
	assert.Equal(t, "header1", headers.Get(configuration.AUTH_HEADERS))
	assert.Equal(t, "signature", headers.Get(configuration.SIGNATURE))
	assert.Equal(t, "Bearer incoming-token", headers.Get("Authorization"))
	assert.Equal(t, "drop", headers.Get("X-Drop-Me"))
	assert.Empty(t, headers.Get("X-Aegis-Proxy"))
}

func TestBuildUpstreamHeadersAppliesEnvBackedInjectedHeaders(t *testing.T) {
	t.Setenv("UPSTREAM_AUTHORIZATION", "Bearer env-token")

	cfg := &configuration.MainConfiguration{
		Server: configuration.Server{
			InjectHeaders: []configuration.InjectHeader{
				{Name: "Authorization", ValueFromEnv: "UPSTREAM_AUTHORIZATION"},
			},
		},
	}
	require.NoError(t, cfg.Validate())
	router := NewRouter(cfg)

	headers := http.Header{}
	headers.Set(configuration.AUTH_CORRELATIONID, "corr-1")

	upstreamHeaders := router.buildUpstreamHeaders(headers)

	assert.Equal(t, "Bearer env-token", upstreamHeaders.Get("Authorization"))
}
