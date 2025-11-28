package main

import (
	"fmt"
	"log"

	"github.com/savegress/platform/backend/pkg/license"
)

func main() {
	keyPair, err := license.GenerateKeyPair()
	if err != nil {
		log.Fatalf("Failed to generate key pair: %v", err)
	}

	fmt.Println("=== License Key Pair Generated ===")
	fmt.Println()
	fmt.Println("Add these to your environment variables:")
	fmt.Println()
	fmt.Printf("LICENSE_PRIVATE_KEY=%s\n", keyPair.PrivateKeyBase64())
	fmt.Println()
	fmt.Printf("LICENSE_PUBLIC_KEY=%s\n", keyPair.PublicKeyBase64())
	fmt.Println()
	fmt.Println("IMPORTANT:")
	fmt.Println("- Keep the PRIVATE key secret (only on the license server)")
	fmt.Println("- The PUBLIC key can be embedded in CDC engines for offline validation")
}
