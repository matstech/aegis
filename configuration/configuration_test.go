// Copyright 2026-present matstech
// SPDX-License-Identifier: GPL-3.0-only

package configuration

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInjectHeaderResolveInline(t *testing.T) {
	t.Setenv("UPSTREAM_TOKEN", "from-env")

	headerName, value, err := (InjectHeader{Name: "X-Test", Value: "inline"}).Resolve()

	require.NoError(t, err)
	assert.Equal(t, "X-Test", headerName)
	assert.Equal(t, "inline", value)
}

func TestInjectHeaderResolveFromEnv(t *testing.T) {
	t.Setenv("UPSTREAM_TOKEN", "from-env")

	headerName, value, err := (InjectHeader{Name: "Authorization", ValueFromEnv: "UPSTREAM_TOKEN"}).Resolve()

	require.NoError(t, err)
	assert.Equal(t, "Authorization", headerName)
	assert.Equal(t, "from-env", value)
}

func TestInjectHeaderResolveRejectsMissingName(t *testing.T) {
	_, _, err := (InjectHeader{Value: "inline"}).Resolve()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "name must be configured")
}

func TestInjectHeaderResolveRejectsMissingDefinition(t *testing.T) {
	_, _, err := (InjectHeader{Name: "X-Test"}).Resolve()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "exactly one of value or valueFromEnv")
	assert.Contains(t, err.Error(), "X-Test")
}

func TestInjectHeaderResolveRejectsConflictingDefinition(t *testing.T) {
	_, _, err := (InjectHeader{Name: "X-Test", Value: "inline", ValueFromEnv: "UPSTREAM_TOKEN"}).Resolve()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "exactly one of value or valueFromEnv")
	assert.Contains(t, err.Error(), "X-Test")
}

func TestInjectHeaderResolveRejectsMissingEnv(t *testing.T) {
	_, _, err := (InjectHeader{Name: "Authorization", ValueFromEnv: "UPSTREAM_TOKEN"}).Resolve()

	require.Error(t, err)
	assert.Contains(t, err.Error(), `environment variable "UPSTREAM_TOKEN" not found`)
}

func TestServerResolveInjectHeaders(t *testing.T) {
	t.Setenv("UPSTREAM_AUTHORIZATION", "Bearer secret")

	resolved, err := (Server{
		InjectHeaders: []InjectHeader{
			{Name: "X-Aegis-Proxy", Value: "true"},
			{Name: "Authorization", ValueFromEnv: "UPSTREAM_AUTHORIZATION"},
		},
	}).ResolveInjectHeaders()

	require.NoError(t, err)
	assert.Equal(t, map[string]string{
		"authorization": "Bearer secret",
		"x-aegis-proxy": "true",
	}, resolved)
}

func TestConfigurationLoadAcceptsExplicitInjectHeadersShape(t *testing.T) {
	t.Setenv("UPSTREAM_AUTHORIZATION", "Bearer secret")

	cfg, err := unmarshalConfigurationJSON([]byte(`{
		"ginmode": "debug",
		"loglevel": "debug",
		"server": {
			"mode": "PLAIN",
			"port": 8080,
			"upstream": "httpbin.org",
			"injectHeaders": [
				{
					"name": "X-Aegis-Proxy",
					"value": "true"
				},
				{
					"name": "Authorization",
					"valueFromEnv": "UPSTREAM_AUTHORIZATION"
				}
			]
		},
		"kids": ["test"]
	}`))

	require.NoError(t, err)
	require.NoError(t, cfg.Validate())
	assert.Equal(t, map[string]string{
		"authorization": "Bearer secret",
		"x-aegis-proxy": "true",
	}, cfg.Server.ResolvedInjectHeaders())
}

func TestConfigurationLoadRejectsLegacyInjectHeadersStringShorthand(t *testing.T) {
	_, err := unmarshalConfigurationJSON([]byte(`{
		"ginmode": "debug",
		"loglevel": "debug",
		"server": {
			"mode": "PLAIN",
			"port": 8080,
			"upstream": "httpbin.org",
			"injectHeaders": ["X-Aegis-Proxy"]
		},
		"kids": ["test"]
	}`))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot unmarshal string")
}

func TestServerResolveInjectHeadersRejectsDuplicateNames(t *testing.T) {
	_, err := (Server{
		InjectHeaders: []InjectHeader{
			{Name: "Authorization", Value: "one"},
			{Name: "authorization", Value: "two"},
		},
	}).ResolveInjectHeaders()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate injectHeaders entry")
}

func unmarshalConfigurationJSON(raw []byte) (*MainConfiguration, error) {
	var cfg MainConfiguration
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
