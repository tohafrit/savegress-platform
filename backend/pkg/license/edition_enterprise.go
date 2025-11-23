//go:build enterprise

package license

// Edition constants for Enterprise build
const (
	Edition         = "enterprise"
	EditionFull     = "Savegress Enterprise Edition"
	MaxTierAllowed  = TierEnterprise
	AllowAllSources = true
)

// BuiltInFeatures returns features compiled into Enterprise build
func BuiltInFeatures() []Feature {
	return append(append(
		CommunityFeatures,
		ProFeatures...),
		EnterpriseFeatures...,
	)
}

// IsSourceCompiled returns true if source type is compiled in Enterprise build
func IsSourceCompiled(sourceType string) bool {
	return true // All sources in enterprise build
}
