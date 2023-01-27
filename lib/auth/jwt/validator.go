package jwt

import (
	"context"
	"crypto"
	"fmt"

	"github.com/hashicorp/cap/jwt"
	"github.com/hashicorp/nomad/nomad/structs"
)

// Validate performs bearer token validation and returns a list of claims
func Validate(ctx context.Context, token string, methodConf *structs.ACLAuthMethodConfig) (map[string]interface{}, error) {
	out := map[string]interface{}{}

	// JWT validation can happen in 3 ways:
	// - via embedded public keys, locally
	// - via JWKS
	// - or via OIDC provider
	if len(methodConf.JWTValidationPubKeys) != 0 {
		claims, err := validateStaticKeys(ctx, methodConf.JWTValidationPubKeys, token)
		if err != nil {
			return out, err
		}
		return claims, nil

	}
	return out, nil
}

func validateStaticKeys(ctx context.Context, keys []string, token string) (map[string]interface{}, error) {
	out := map[string]interface{}{}
	parsedKeys := []crypto.PublicKey{}
	for _, v := range keys {
		key, err := jwt.ParsePublicKeyPEM([]byte(v))
		parsedKeys = append(parsedKeys, key)
		if err != nil {
			return out, fmt.Errorf("unable to parse public key for JWT auth: %v", err)
		}
	}

	keySet, err := jwt.NewStaticKeySet(parsedKeys)
	if err != nil {
		return out, err
	}

	claims, err := keySet.VerifySignature(ctx, token)
	if err != nil {
		return out, fmt.Errorf("unable to verify signature of JWT bearer token: %v", err)
	}

	return claims, nil
}
