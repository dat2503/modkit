import type { IFeatureFlagsService, FlagContext, FlagState } from '../../../contracts/ts/feature-flags'
import type { ICacheService } from '../../../contracts/ts/cache'

export interface FlagsmithConfig {
  serverKey: string
  apiUrl?: string
  cacheTtlSeconds?: number
}

/**
 * FlagsmithService implements IFeatureFlagsService using Flagsmith.
 */
export class FlagsmithService implements IFeatureFlagsService {
  constructor(
    private readonly config: FlagsmithConfig,
    private readonly cache?: ICacheService,
  ) {}

  async isEnabled(flagName: string, evalCtx?: FlagContext): Promise<boolean> {
    // TODO: implement using flagsmith-nodejs or @flagsmith/flag-engine
    console.warn('[flagsmith] stub: isEnabled() not implemented')
    return false
  }

  async getVariant(flagName: string, evalCtx?: FlagContext): Promise<string> {
    // TODO: implement using flagsmith getFeatureValue(flagName)
    console.warn('[flagsmith] stub: getVariant() not implemented')
    return ''
  }

  async getValue(flagName: string, evalCtx?: FlagContext): Promise<unknown> {
    // TODO: implement using flagsmith getFeatureValue(flagName)
    console.warn('[flagsmith] stub: getValue() not implemented')
    return undefined
  }

  async getAllFlags(evalCtx?: FlagContext): Promise<Record<string, FlagState>> {
    // TODO: implement by fetching all flags and building a map
    console.warn('[flagsmith] stub: getAllFlags() not implemented')
    return {}
  }
}
