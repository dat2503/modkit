package contracts

import "context"

// FeatureFlagsService manages feature flags for phased rollouts, A/B testing, and kill switches.
// Gate features behind flags that can be toggled without a deploy.
type FeatureFlagsService interface {
	// IsEnabled returns true if the given flag is enabled for the current context.
	// context can include user ID, environment, or custom properties for targeting rules.
	IsEnabled(ctx context.Context, flagName string, evalCtx FlagContext) (bool, error)

	// GetVariant returns the variant value for a multivariate flag.
	// Use for A/B tests where different users see different versions.
	// Returns the default variant if the flag is off or the user is not in any variant.
	GetVariant(ctx context.Context, flagName string, evalCtx FlagContext) (string, error)

	// GetValue returns the remote config value for the given flag.
	// Use for flags that carry a value (string, number, JSON) rather than just on/off.
	GetValue(ctx context.Context, flagName string, evalCtx FlagContext) (any, error)

	// GetAllFlags returns all flags and their evaluated state for the given context.
	// Use to batch-fetch all flags at the start of a request to avoid multiple round-trips.
	GetAllFlags(ctx context.Context, evalCtx FlagContext) (map[string]FlagState, error)
}

// FlagContext provides targeting context for flag evaluation.
// The more context you provide, the more precise targeting rules can be.
type FlagContext struct {
	// UserID is the authenticated user's ID. Used for user-targeted rollouts.
	UserID string

	// Email is the user's email. Used for email-domain targeting.
	Email string

	// Environment is the runtime environment (e.g. "production", "staging").
	// If empty, the service's configured environment is used.
	Environment string

	// Traits holds additional custom properties for targeting (e.g. plan, org, region).
	Traits map[string]any
}

// FlagState is the evaluated state of a single flag.
type FlagState struct {
	// Enabled is whether the flag is on for this context.
	Enabled bool

	// Variant is the variant value (for multivariate flags). Empty if not applicable.
	Variant string

	// Value is the remote config value (for value flags). Nil if not applicable.
	Value any
}
