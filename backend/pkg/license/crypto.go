package license

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

// KeyPair holds Ed25519 key pair for license signing
type KeyPair struct {
	PublicKey  ed25519.PublicKey
	PrivateKey ed25519.PrivateKey
}

// GenerateKeyPair generates a new Ed25519 key pair for license signing
// The private key should be kept secure and used only by the license server
func GenerateKeyPair() (*KeyPair, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}
	return &KeyPair{
		PublicKey:  pub,
		PrivateKey: priv,
	}, nil
}

// PublicKeyBase64 returns the public key as base64 string
func (kp *KeyPair) PublicKeyBase64() string {
	return base64.StdEncoding.EncodeToString(kp.PublicKey)
}

// PrivateKeyBase64 returns the private key as base64 string
func (kp *KeyPair) PrivateKeyBase64() string {
	return base64.StdEncoding.EncodeToString(kp.PrivateKey)
}

// LoadKeyPair loads a key pair from base64 encoded strings
func LoadKeyPair(pubBase64, privBase64 string) (*KeyPair, error) {
	pub, err := base64.StdEncoding.DecodeString(pubBase64)
	if err != nil {
		return nil, fmt.Errorf("invalid public key: %w", err)
	}

	priv, err := base64.StdEncoding.DecodeString(privBase64)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	if len(pub) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key size")
	}
	if len(priv) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key size")
	}

	return &KeyPair{
		PublicKey:  ed25519.PublicKey(pub),
		PrivateKey: ed25519.PrivateKey(priv),
	}, nil
}

// LicenseGenerator creates signed licenses
type LicenseGenerator struct {
	keyPair *KeyPair
	issuer  string
}

// NewLicenseGenerator creates a new license generator
func NewLicenseGenerator(keyPair *KeyPair, issuer string) *LicenseGenerator {
	return &LicenseGenerator{
		keyPair: keyPair,
		issuer:  issuer,
	}
}

// NewLicenseGeneratorFromBase64 creates a license generator from a base64 encoded private key
func NewLicenseGeneratorFromBase64(privateKeyBase64 string) (*LicenseGenerator, error) {
	privBytes, err := base64.StdEncoding.DecodeString(privateKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("invalid private key encoding: %w", err)
	}
	if len(privBytes) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key size")
	}

	privateKey := ed25519.PrivateKey(privBytes)
	publicKey := privateKey.Public().(ed25519.PublicKey)

	return &LicenseGenerator{
		keyPair: &KeyPair{
			PrivateKey: privateKey,
			PublicKey:  publicKey,
		},
		issuer: "license.savegress.io",
	}, nil
}

// GenerateRequest contains parameters for generating a license
type GenerateRequest struct {
	CustomerID   string
	CustomerName string
	Tier         Tier
	Features     []Feature // Additional features beyond tier defaults
	Limits       *Limits   // Custom limits (nil = tier defaults)
	ValidDays    int       // How long the license is valid
	HardwareID   string    // Optional hardware binding
	Metadata     map[string]string
}

// Generate creates a new signed license
func (g *LicenseGenerator) Generate(req GenerateRequest) (LicenseKey, error) {
	// Generate unique ID
	idBytes := make([]byte, 16)
	if _, err := rand.Read(idBytes); err != nil {
		return "", fmt.Errorf("failed to generate license ID: %w", err)
	}
	licenseID := fmt.Sprintf("%x-%x-%x-%x-%x",
		idBytes[0:4], idBytes[4:6], idBytes[6:8], idBytes[8:10], idBytes[10:16])

	// Determine limits
	limits := g.defaultLimits(req.Tier)
	if req.Limits != nil {
		limits = *req.Limits
	}

	// Merge features
	features := g.defaultFeatures(req.Tier)
	for _, f := range req.Features {
		if !containsFeature(features, f) {
			features = append(features, f)
		}
	}

	// Create license
	license := License{
		ID:           licenseID,
		CustomerID:   req.CustomerID,
		CustomerName: req.CustomerName,
		Tier:         req.Tier,
		Features:     features,
		Limits:       limits,
		IssuedAt:     time.Now().UTC(),
		ExpiresAt:    time.Now().UTC().AddDate(0, 0, req.ValidDays),
		HardwareID:   req.HardwareID,
		Issuer:       g.issuer,
		Version:      1,
		Metadata:     req.Metadata,
	}

	// Serialize to JSON (without signature)
	license.Signature = ""
	jsonData, err := json.Marshal(license)
	if err != nil {
		return "", fmt.Errorf("failed to serialize license: %w", err)
	}

	// Sign
	signature := ed25519.Sign(g.keyPair.PrivateKey, jsonData)

	// Create license key: base64url(json).base64url(signature)
	key := fmt.Sprintf("%s.%s",
		base64.RawURLEncoding.EncodeToString(jsonData),
		base64.RawURLEncoding.EncodeToString(signature))

	return LicenseKey(key), nil
}

func (g *LicenseGenerator) defaultLimits(tier Tier) Limits {
	switch tier {
	case TierEnterprise:
		return EnterpriseLimits
	case TierPro, TierTrial:
		return ProLimits
	default:
		return CommunityLimits
	}
}

func (g *LicenseGenerator) defaultFeatures(tier Tier) []Feature {
	var features []Feature

	// Always include community features
	features = append(features, CommunityFeatures...)

	switch tier {
	case TierEnterprise:
		features = append(features, EnterpriseFeatures...)
		fallthrough
	case TierPro, TierTrial:
		features = append(features, ProFeatures...)
	}

	return features
}

func containsFeature(features []Feature, f Feature) bool {
	for _, feat := range features {
		if feat == f {
			return true
		}
	}
	return false
}

// VerifyLicense verifies a license signature using a public key
func VerifyLicense(key LicenseKey, publicKey ed25519.PublicKey) (*License, error) {
	parts := splitLicenseKey(string(key))
	if len(parts) != 2 {
		return nil, ErrInvalidLicense
	}

	jsonData, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("%w: invalid payload encoding", ErrInvalidLicense)
	}

	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("%w: invalid signature encoding", ErrInvalidLicense)
	}

	if !ed25519.Verify(publicKey, jsonData, signature) {
		return nil, ErrInvalidSignature
	}

	var license License
	if err := json.Unmarshal(jsonData, &license); err != nil {
		return nil, fmt.Errorf("%w: invalid JSON", ErrInvalidLicense)
	}

	return &license, nil
}

func splitLicenseKey(key string) []string {
	result := make([]string, 0, 2)
	lastDot := -1
	for i := len(key) - 1; i >= 0; i-- {
		if key[i] == '.' {
			lastDot = i
			break
		}
	}
	if lastDot == -1 {
		return []string{key}
	}
	return append(result, key[:lastDot], key[lastDot+1:])
}
