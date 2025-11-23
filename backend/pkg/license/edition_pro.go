//go:build pro

package license

// Edition constants for Pro build
const (
	Edition         = "pro"
	EditionFull     = "Savegress Pro Edition"
	MaxTierAllowed  = TierPro
	AllowAllSources = false
)

// BuiltInFeatures returns features compiled into Pro build
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
