//go:build pro

package license

// Edition constants for Pro build
const (
	Edition         = "pro"
	EditionFull     = "Savegress Pro Edition"
	MaxTierAllowed  = TierPro
	AllowAllSources = false
)

// BuiltInFeatures returns features compiled into Pro build.
// Pro edition is for production at scale - includes performance, reliability,
// and DevOps tooling for serious production deployments.
func BuiltInFeatures() []Feature {
	return append(CommunityFeatures, ProFeatures...)
}

// IsSourceCompiled returns true if source type is compiled in Pro build
func IsSourceCompiled(sourceType string) bool {
	switch sourceType {
	case "postgres", "postgresql", "mysql", "mariadb":
		return true
	case "mongodb", "sqlserver", "cassandra", "dynamodb":
		return true
	default:
		return false
	}
}

// IsFeatureCompiled returns true if a feature is available in Pro build.
func IsFeatureCompiled(feature Feature) bool {
	for _, f := range CommunityFeatures {
		if f == feature {
			return true
		}
	}
	for _, f := range ProFeatures {
		if f == feature {
			return true
		}
	}
	return false
}
