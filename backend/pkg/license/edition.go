// Package license - edition detection
// This file is overridden by build tags: community, pro, enterprise

//go:build !community && !pro && !enterprise

package license

// Edition constants - default build includes all features
const (
	Edition         = "development"
	EditionFull     = "Savegress Development Edition"
	MaxTierAllowed  = TierEnterprise
	AllowAllSources = true
)

// BuiltInFeatures returns features compiled into this build
func BuiltInFeatures() []Feature {
	// Development build has all features
	return append(append(
		CommunityFeatures,
		ProFeatures...),
		EnterpriseFeatures...,
	)
}

// IsSourceCompiled returns true if source type is compiled in
func IsSourceCompiled(sourceType string) bool {
	return true // All sources in dev build
}
