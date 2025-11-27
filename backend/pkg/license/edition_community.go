//go:build community

package license

// Edition constants for Community build
const (
	Edition         = "community"
	EditionFull     = "Savegress Community Edition"
	MaxTierAllowed  = TierCommunity
	AllowAllSources = false
)

// BuiltInFeatures returns features compiled into Community build.
// Community edition provides a fully functional CDC system for startups and small projects.
// Basic safety features (rate limiting, circuit breaker) are always available.
func BuiltInFeatures() []Feature {
	return CommunityFeatures
}

// IsSourceCompiled returns true if source type is compiled in Community build
func IsSourceCompiled(sourceType string) bool {
	switch sourceType {
	case "postgres", "postgresql", "mysql", "mariadb":
		return true
	default:
		return false
	}
}

// IsFeatureCompiled returns true if a feature is available in Community build.
// Note: Basic safety features are always available regardless of this check.
func IsFeatureCompiled(feature Feature) bool {
	for _, f := range CommunityFeatures {
		if f == feature {
			return true
		}
	}
	return false
}
