package s3

import (
	"os"
)

// providerEndpoints stores default endpoints for each provider.
var providerEndpoints = map[Provider]string{
	ProviderB2: "s3.us-west-004.backblazeb2.com",
}

// Config stores S3 connection parameters.
type Config struct {
	Provider  Provider // provider type
	Endpoint  string   // custom endpoint (overrides provider default)
	Region    string   // e.g., "us-west-004"
	AccessKey string   // Key ID
	SecretKey string   // Application Key
	Secure    bool     // use HTTPS, defaults to true
}

// resolveEndpoint resolves the endpoint based on provider.
func resolveEndpoint(cfg Config) string {
	if cfg.Endpoint != "" {
		return cfg.Endpoint
	}
	if ep, ok := providerEndpoints[cfg.Provider]; ok {
		return ep
	}
	return ""
}

// applyEnvOverrides fills empty Config fields from environment variables.
func applyEnvOverrides(cfg *Config) {
	if cfg.Provider == "" {
		if p := os.Getenv("S3_PROVIDER"); p != "" {
			cfg.Provider = Provider(p)
		}
	}
	if cfg.Endpoint == "" {
		cfg.Endpoint = os.Getenv("S3_ENDPOINT")
	}
	if cfg.AccessKey == "" {
		cfg.AccessKey = os.Getenv("S3_ACCESS_KEY")
	}
	if cfg.SecretKey == "" {
		cfg.SecretKey = os.Getenv("S3_SECRET_KEY")
	}
}
