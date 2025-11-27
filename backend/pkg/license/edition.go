// Package license - edition detection
// This file is overridden by build tags: community, pro, enterprise

//go:build !community && !pro && !enterprise

package license

// Edition constants - default build includes all features (for development)
const (
	Edition         = "development"
	EditionFull     = "Savegress Development Edition"
	MaxTierAllowed  = TierEnterprise
	AllowAllSources = true
)

// BuiltInFeatures returns features compiled into this build.
// Development build has all features enabled for testing.
func BuiltInFeatures() []Feature {
	result := make([]Feature, 0, len(CommunityFeatures)+len(ProFeatures)+len(EnterpriseFeatures))
	result = append(result, CommunityFeatures...)
	result = append(result, ProFeatures...)
	result = append(result, EnterpriseFeatures...)
	return result
}

// IsSourceCompiled returns true if source type is compiled in
func IsSourceCompiled(sourceType string) bool {
	return true // All sources in dev build
}

// IsFeatureCompiled returns true if a feature is available in this build.
// Development build has all features.
func IsFeatureCompiled(feature Feature) bool {
	return true // All features in dev build
}
