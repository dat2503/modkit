// Package flagsmith implements the FeatureFlagsService interface using Flagsmith.
package flagsmith

import (
	"context"
	"log"

	contracts "github.com/dat2503/modkit/contracts/go"
)

// Config holds the configuration for the Flagsmith feature flags provider.
type Config struct {
	// ServerKey is the Flagsmith server-side API key (starts with "ser.").
	ServerKey string

	// APIURL is the Flagsmith API endpoint. Defaults to Flagsmith Cloud.
	APIURL string

	// CacheTTL is how long to cache flag values in seconds. 0 = disabled.
	CacheTTL int
}

// Service implements contracts.FeatureFlagsService using Flagsmith.
type Service struct {
	cfg   Config
	cache contracts.CacheService
	// TODO: add flagsmith-go client
}

// New creates a new Flagsmith feature flags service.
// cache is optional — pass nil to disable caching.
func New(cfg Config, cache contracts.CacheService) *Service {
	if cfg.APIURL == "" {
		cfg.APIURL = "https://edge.api.flagsmith.com/api/v1/"
	}
	return &Service{cfg: cfg, cache: cache}
}

func (s *Service) IsEnabled(ctx context.Context, flagName string, evalCtx contracts.FlagContext) (bool, error) {
	// TODO: implement using github.com/Flagsmith/flagsmith-go-client
	// client.GetEnvironmentFlags() or client.GetIdentityFlags(identifier, traits)
	log.Printf("[flagsmith] stub: IsEnabled() not implemented")
	return false, nil
}

func (s *Service) GetVariant(ctx context.Context, flagName string, evalCtx contracts.FlagContext) (string, error) {
	// TODO: implement using flags.GetFeatureValue(flagName) for multivariate
	log.Printf("[flagsmith] stub: GetVariant() not implemented")
	return "", nil
}

func (s *Service) GetValue(ctx context.Context, flagName string, evalCtx contracts.FlagContext) (any, error) {
	// TODO: implement using flags.GetFeatureValue(flagName)
	log.Printf("[flagsmith] stub: GetValue() not implemented")
	return nil, nil
}

func (s *Service) GetAllFlags(ctx context.Context, evalCtx contracts.FlagContext) (map[string]contracts.FlagState, error) {
	// TODO: implement by calling GetEnvironmentFlags or GetIdentityFlags and building map
	log.Printf("[flagsmith] stub: GetAllFlags() not implemented")
	return nil, nil
}
