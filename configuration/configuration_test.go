package configuration

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeaderValueSourceResolveInline(t *testing.T) {
	t.Setenv("UPSTREAM_TOKEN", "from-env")

	value, err := (HeaderValueSource{Value: "inline"}).Resolve("X-Test")

	require.NoError(t, err)
	assert.Equal(t, "inline", value)
}

func TestHeaderValueSourceResolveFromEnv(t *testing.T) {
	t.Setenv("UPSTREAM_TOKEN", "from-env")

	value, err := (HeaderValueSource{ValueFromEnv: "UPSTREAM_TOKEN"}).Resolve("Authorization")

	require.NoError(t, err)
	assert.Equal(t, "from-env", value)
}

func TestHeaderValueSourceResolveRejectsMissingDefinition(t *testing.T) {
	_, err := (HeaderValueSource{}).Resolve("X-Test")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "exactly one of value or valueFromEnv")
	assert.Contains(t, err.Error(), "X-Test")
}

func TestHeaderValueSourceResolveRejectsConflictingDefinition(t *testing.T) {
	_, err := (HeaderValueSource{Value: "inline", ValueFromEnv: "UPSTREAM_TOKEN"}).Resolve("X-Test")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "exactly one of value or valueFromEnv")
	assert.Contains(t, err.Error(), "X-Test")
}

func TestHeaderValueSourceResolveRejectsMissingEnv(t *testing.T) {
	_, err := (HeaderValueSource{ValueFromEnv: "UPSTREAM_TOKEN"}).Resolve("Authorization")

	require.Error(t, err)
	assert.Contains(t, err.Error(), `environment variable "UPSTREAM_TOKEN" not found`)
}

func TestServerResolveInjectHeaders(t *testing.T) {
	t.Setenv("UPSTREAM_AUTHORIZATION", "Bearer secret")

	resolved, err := (Server{
		InjectHeaders: map[string]HeaderValueSource{
			"X-Aegis-Proxy": {Value: "true"},
			"Authorization": {ValueFromEnv: "UPSTREAM_AUTHORIZATION"},
		},
	}).ResolveInjectHeaders()

	require.NoError(t, err)
	assert.Equal(t, map[string]string{
		"X-Aegis-Proxy": "true",
		"Authorization": "Bearer secret",
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
			"injectHeaders": {
				"X-Aegis-Proxy": {
					"value": "true"
				},
				"Authorization": {
					"valueFromEnv": "UPSTREAM_AUTHORIZATION"
				}
			}
		},
		"kids": ["test"]
	}`))

	require.NoError(t, err)
	require.NoError(t, cfg.Validate())
	assert.Equal(t, map[string]string{
		"X-Aegis-Proxy": "true",
		"Authorization": "Bearer secret",
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
			"injectHeaders": {
				"X-Aegis-Proxy": "true"
			}
		},
		"kids": ["test"]
	}`))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot unmarshal string")
}

func unmarshalConfigurationJSON(raw []byte) (*MainConfiguration, error) {
	var cfg MainConfiguration
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
