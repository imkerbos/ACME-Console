package acme

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
)

var (
	ErrInvalidKeyType    = errors.New("invalid key type: must be RSA or ECC")
	ErrInvalidRSAKeySize = errors.New("invalid RSA key size: must be 2048 or 4096")
	ErrInvalidECCKeySize = errors.New("invalid ECC key size: must be 256 or 384")
	ErrInvalidPEMBlock   = errors.New("invalid PEM block")
	ErrUnsupportedKey    = errors.New("unsupported key type")
)

// KeyType represents the type of cryptographic key
type KeyType string

const (
	KeyTypeRSA KeyType = "RSA"
	KeyTypeECC KeyType = "ECC"
)

// GeneratePrivateKey generates a new private key based on the specified type and size.
// For RSA: size can be 2048 or 4096
// For ECC: size can be 256 (P-256) or 384 (P-384)
func GeneratePrivateKey(keyType KeyType, size int) (crypto.PrivateKey, error) {
	switch keyType {
	case KeyTypeRSA:
		return generateRSAKey(size)
	case KeyTypeECC:
		return generateECCKey(size)
	default:
		return nil, ErrInvalidKeyType
	}
}

func generateRSAKey(size int) (*rsa.PrivateKey, error) {
	if size != 2048 && size != 4096 {
		return nil, ErrInvalidRSAKeySize
	}
	return rsa.GenerateKey(rand.Reader, size)
}

func generateECCKey(size int) (*ecdsa.PrivateKey, error) {
	var curve elliptic.Curve
	switch size {
	case 256:
		curve = elliptic.P256()
	case 384:
		curve = elliptic.P384()
	default:
		return nil, ErrInvalidECCKeySize
	}
	return ecdsa.GenerateKey(curve, rand.Reader)
}

// EncodePrivateKeyPEM encodes a private key to PEM format.
func EncodePrivateKeyPEM(key crypto.PrivateKey) ([]byte, error) {
	switch k := key.(type) {
	case *rsa.PrivateKey:
		return pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(k),
		}), nil
	case *ecdsa.PrivateKey:
		der, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal EC private key: %w", err)
		}
		return pem.EncodeToMemory(&pem.Block{
			Type:  "EC PRIVATE KEY",
			Bytes: der,
		}), nil
	default:
		return nil, ErrUnsupportedKey
	}
}

// DecodePrivateKeyPEM decodes a PEM-encoded private key.
func DecodePrivateKeyPEM(pemData []byte) (crypto.PrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, ErrInvalidPEMBlock
	}

	switch block.Type {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	case "EC PRIVATE KEY":
		return x509.ParseECPrivateKey(block.Bytes)
	case "PRIVATE KEY":
		// PKCS#8 format
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PKCS8 private key: %w", err)
		}
		return key, nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedKey, block.Type)
	}
}

// GetDefaultKeySize returns the default key size for the given key type.
func GetDefaultKeySize(keyType KeyType) int {
	switch keyType {
	case KeyTypeRSA:
		return 2048
	case KeyTypeECC:
		return 256
	default:
		return 0
	}
}

// ValidateKeySize checks if the key size is valid for the given key type.
func ValidateKeySize(keyType KeyType, size int) error {
	switch keyType {
	case KeyTypeRSA:
		if size != 2048 && size != 4096 {
			return ErrInvalidRSAKeySize
		}
	case KeyTypeECC:
		if size != 256 && size != 384 {
			return ErrInvalidECCKeySize
		}
	default:
		return ErrInvalidKeyType
	}
	return nil
}
