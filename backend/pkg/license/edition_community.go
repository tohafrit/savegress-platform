//go:build community

package license

// Edition constants for Community build
const (
	Edition         = "community"
	EditionFull     = "Savegress Community Edition"
	MaxTierAllowed  = TierCommunity
	AllowAllSources = false
)

// BuiltInFeatures returns features compiled into Community build
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
