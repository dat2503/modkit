/**
 * IFeatureFlagsService manages feature flags for phased rollouts, A/B testing, and kill switches.
 * Gate features behind flags that can be toggled without a deploy.
 */
export interface IFeatureFlagsService {
  /**
   * Returns true if the given flag is enabled for the current context.
   */
  isEnabled(flagName: string, evalCtx?: FlagContext): Promise<boolean>;

  /**
   * Returns the variant value for a multivariate flag.
   * Use for A/B tests where different users see different versions.
   */
  getVariant(flagName: string, evalCtx?: FlagContext): Promise<string>;

  /**
   * Returns the remote config value for the given flag.
   * Use for flags that carry a value rather than just on/off.
   */
  getValue(flagName: string, evalCtx?: FlagContext): Promise<unknown>;

  /**
   * Returns all flags and their evaluated state for the given context.
   * Batch-fetch all flags at the start of a request to avoid multiple round-trips.
   */
  getAllFlags(evalCtx?: FlagContext): Promise<Record<string, FlagState>>;
}

/** Provides targeting context for flag evaluation. */
export interface FlagContext {
  /** Authenticated user's ID. Used for user-targeted rollouts. */
  userId?: string;

  /** User's email. Used for email-domain targeting. */
  email?: string;

  /** Runtime environment (e.g. "production", "staging"). */
  environment?: string;

  /** Additional custom properties for targeting (e.g. plan, org, region). */
  traits?: Record<string, unknown>;
}

/** The evaluated state of a single flag. */
export interface FlagState {
  /** Whether the flag is on for this context. */
  enabled: boolean;

  /** Variant value for multivariate flags. Empty string if not applicable. */
  variant?: string;

  /** Remote config value for value flags. Undefined if not applicable. */
  value?: unknown;
}
