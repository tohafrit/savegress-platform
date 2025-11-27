//go:build enterprise

package license

// Edition constants for Enterprise build
const (
	Edition         = "enterprise"
	EditionFull     = "Savegress Enterprise Edition"
	MaxTierAllowed  = TierEnterprise
	AllowAllSources = true
)

// BuiltInFeatures returns features compiled into Enterprise build.
// Enterprise edition is for governance, compliance, and multi-team operations.
// Includes all features for large organizations and regulated industries.
func BuiltInFeatures() []Feature {
	result := make([]Feature, 0, len(CommunityFeatures)+len(ProFeatures)+len(EnterpriseFeatures))
	result = append(result, CommunityFeatures...)
	result = append(result, ProFeatures...)
	result = append(result, EnterpriseFeatures...)
	return result
}

// IsSourceCompiled returns true if source type is compiled in Enterprise build
func IsSourceCompiled(sourceType string) bool {
	return true // All sources in enterprise build
}

// IsFeatureCompiled returns true if a feature is available in Enterprise build.
// Enterprise has access to all features.
func IsFeatureCompiled(feature Feature) bool {
	return true // All features in enterprise build
}
