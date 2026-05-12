package acme

import (
	"context"
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"golang.org/x/crypto/acme"
)

const (
	// LetsEncryptProductionV2 is the production ACME directory URL
	LetsEncryptProductionV2 = "https://acme-v02.api.letsencrypt.org/directory"
	// LetsEncryptStagingV2 is the staging ACME directory URL (no rate limits, fake certs)
	LetsEncryptStagingV2 = "https://acme-staging-v02.api.letsencrypt.org/directory"
)

// GetCAURL returns the ACME directory URL for the given environment.
func GetCAURL(env string) string {
	switch env {
	case "staging":
		return LetsEncryptStagingV2
	default:
		return LetsEncryptProductionV2
	}
}

// ClientV2 wraps the official Go ACME client
type ClientV2 struct {
	client *acme.Client
	email  string
}

// NewClientV2 creates a new ACME client using the official Go library
func NewClientV2(accountKey crypto.Signer, email string) *ClientV2 {
	client := &acme.Client{
		Key:          accountKey,
		DirectoryURL: LetsEncryptProductionV2,
	}

	return &ClientV2{
		client: client,
		email:  email,
	}
}

// NewClientV2WithCA creates a new ACME client with a specific CA directory URL.
func NewClientV2WithCA(accountKey crypto.Signer, email string, caURL string) *ClientV2 {
	if caURL == "" {
		caURL = LetsEncryptProductionV2
	}

	client := &acme.Client{
		Key:          accountKey,
		DirectoryURL: caURL,
	}

	return &ClientV2{
		client: client,
		email:  email,
	}
}

// Register registers a new ACME account
func (c *ClientV2) Register(ctx context.Context) error {
	account := &acme.Account{
		Contact: []string{"mailto:" + c.email},
	}

	_, err := c.client.Register(ctx, account, acme.AcceptTOS)
	if err != nil {
		// If account already exists, that's fine
		if err == acme.ErrAccountAlreadyExists {
			return nil
		}
		return fmt.Errorf("failed to register account: %w", err)
	}

	return nil
}

// CreateOrder creates a new certificate order
func (c *ClientV2) CreateOrder(ctx context.Context, domains []string) (*acme.Order, error) {
	// Convert domains to AuthzID
	var ids []acme.AuthzID
	for _, domain := range domains {
		ids = append(ids, acme.AuthzID{
			Type:  "dns",
			Value: domain,
		})
	}

	order, err := c.client.AuthorizeOrder(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	return order, nil
}

// GetAuthorization fetches an authorization and its challenges
func (c *ClientV2) GetAuthorization(ctx context.Context, authzURL string) (*acme.Authorization, error) {
	authz, err := c.client.GetAuthorization(ctx, authzURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get authorization: %w", err)
	}

	return authz, nil
}

// AcceptChallenge tells the CA to verify a challenge
func (c *ClientV2) AcceptChallenge(ctx context.Context, challenge *acme.Challenge) (*acme.Challenge, error) {
	chal, err := c.client.Accept(ctx, challenge)
	if err != nil {
		return nil, fmt.Errorf("failed to accept challenge: %w", err)
	}

	return chal, nil
}

// WaitOrder waits for an order to reach a final state
func (c *ClientV2) WaitOrder(ctx context.Context, orderURL string) (*acme.Order, error) {
	order, err := c.client.WaitOrder(ctx, orderURL)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for order: %w", err)
	}

	return order, nil
}

// CreateOrderCert finalizes the order and retrieves the certificate
func (c *ClientV2) CreateOrderCert(ctx context.Context, orderURL string, csr []byte) ([][]byte, error) {
	certs, _, err := c.client.CreateOrderCert(ctx, orderURL, csr, true)
	if err != nil {
		return nil, fmt.Errorf("failed to create order cert: %w", err)
	}

	return certs, nil
}

// GetOrder retrieves the current state of an order
func (c *ClientV2) GetOrder(ctx context.Context, orderURL string) (*acme.Order, error) {
	order, err := c.client.GetOrder(ctx, orderURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return order, nil
}

// DNS01ChallengeRecord computes the DNS TXT record value for a DNS-01 challenge.
// This method is on ClientV2 because it needs the account key to compute the thumbprint.
func (c *ClientV2) DNS01ChallengeRecord(token string) (string, error) {
	return c.client.DNS01ChallengeRecord(token)
}

// HTTP01ChallengeResponse computes the key authorization response for an HTTP-01 challenge.
func (c *ClientV2) HTTP01ChallengeResponse(token string) (string, error) {
	return c.client.HTTP01ChallengeResponse(token)
}

// HTTP01ChallengePath returns the URL path for an HTTP-01 challenge.
func (c *ClientV2) HTTP01ChallengePath(token string) string {
	return c.client.HTTP01ChallengePath(token)
}

// RevokeCert revokes a previously issued certificate.
// The certPEM should be a PEM-encoded certificate.
func (c *ClientV2) RevokeCert(ctx context.Context, certPEM []byte) error {
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return fmt.Errorf("failed to decode certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %w", err)
	}

	if err := c.client.RevokeCert(ctx, nil, cert.Raw, acme.CRLReasonUnspecified); err != nil {
		return fmt.Errorf("failed to revoke certificate: %w", err)
	}

	return nil
}

// GetDirectoryURL returns the CA directory URL used by this client.
func (c *ClientV2) GetDirectoryURL() string {
	return c.client.DirectoryURL
}
