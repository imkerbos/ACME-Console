package acme

import (
	"crypto"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
)

const (
	// LetsEncryptProduction is the production ACME directory URL for Let's Encrypt
	LetsEncryptProduction = "https://acme-v02.api.letsencrypt.org/directory"
	// LetsEncryptStaging is the staging ACME directory URL for Let's Encrypt
	LetsEncryptStaging = "https://acme-staging-v02.api.letsencrypt.org/directory"
)

var (
	ErrNoAccount      = errors.New("no ACME account registered")
	ErrOrderNotReady  = errors.New("order is not ready for finalization")
	ErrNoCertificate  = errors.New("no certificate returned from CA")
)

// User implements the lego registration.User interface
type User struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *User) GetEmail() string {
	return u.Email
}

func (u *User) GetRegistration() *registration.Resource {
	return u.Registration
}

func (u *User) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

// Client wraps the lego ACME client with additional functionality
type Client struct {
	client   *lego.Client
	user     *User
	caURL    string
	provider *ManualDNSProvider
}

// ClientConfig holds configuration for creating an ACME client
type ClientConfig struct {
	CAURL      string // ACME directory URL
	Email      string
	PrivateKey crypto.PrivateKey
	// Registration is optional; if nil, a new account will be registered
	Registration *registration.Resource
}

// NewClient creates a new ACME client.
func NewClient(cfg ClientConfig) (*Client, error) {
	user := &User{
		Email:        cfg.Email,
		Registration: cfg.Registration,
		key:          cfg.PrivateKey,
	}

	config := lego.NewConfig(user)
	config.CADirURL = cfg.CAURL

	client, err := lego.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create lego client: %w", err)
	}

	return &Client{
		client: client,
		user:   user,
		caURL:  cfg.CAURL,
	}, nil
}

// SetDNSProvider sets the DNS provider for DNS-01 challenges.
func (c *Client) SetDNSProvider(provider *ManualDNSProvider) error {
	c.provider = provider
	return c.client.Challenge.SetDNS01Provider(provider, dns01.AddRecursiveNameservers([]string{"8.8.8.8:53", "1.1.1.1:53"}))
}

// Register registers a new ACME account with the CA.
func (c *Client) Register() (*registration.Resource, error) {
	reg, err := c.client.Registration.Register(registration.RegisterOptions{
		TermsOfServiceAgreed: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to register account: %w", err)
	}
	c.user.Registration = reg
	return reg, nil
}

// RegisterOrRetrieve tries to register a new account, or retrieves an existing one if email already registered.
func (c *Client) RegisterOrRetrieve() (*registration.Resource, error) {
	// Try to register first
	reg, err := c.Register()
	if err == nil {
		return reg, nil
	}

	// If registration failed, try to retrieve existing account
	reg, err = c.client.Registration.ResolveAccountByKey()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve existing account: %w", err)
	}
	c.user.Registration = reg
	return reg, nil
}

// ObtainCertificate requests a new certificate for the given domains.
// This creates an order, waits for challenges to be set up (via the DNS provider),
// and returns the certificates.
func (c *Client) ObtainCertificate(domains []string, privateKey crypto.PrivateKey) (*certificate.Resource, error) {
	if c.user.Registration == nil {
		return nil, ErrNoAccount
	}

	request := certificate.ObtainRequest{
		Domains:    domains,
		Bundle:     true,
		PrivateKey: privateKey,
	}

	certificates, err := c.client.Certificate.Obtain(request)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain certificate: %w", err)
	}

	if certificates == nil {
		return nil, ErrNoCertificate
	}

	return certificates, nil
}

// ObtainCertificateForOrder obtains a certificate for a pending order.
// This should be called after DNS records have been set up.
func (c *Client) ObtainCertificateForOrder(domains []string, privateKey crypto.PrivateKey) (*certificate.Resource, error) {
	return c.ObtainCertificate(domains, privateKey)
}

// GetCAURL returns the ACME directory URL being used.
func (c *Client) GetCAURL() string {
	return c.caURL
}

// GetEmail returns the email address associated with the account.
func (c *Client) GetEmail() string {
	return c.user.Email
}

// GetRegistration returns the current registration resource.
func (c *Client) GetRegistration() *registration.Resource {
	return c.user.Registration
}

// RegistrationToJSON serializes the registration resource to JSON.
func RegistrationToJSON(reg *registration.Resource) (string, error) {
	if reg == nil {
		return "", nil
	}
	data, err := json.Marshal(reg)
	if err != nil {
		return "", fmt.Errorf("failed to marshal registration: %w", err)
	}
	return string(data), nil
}

// RegistrationFromJSON deserializes a registration resource from JSON.
func RegistrationFromJSON(data string) (*registration.Resource, error) {
	if data == "" {
		return nil, nil
	}
	var reg registration.Resource
	if err := json.Unmarshal([]byte(data), &reg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal registration: %w", err)
	}
	return &reg, nil
}

// GetDirectoryURL returns the appropriate ACME directory URL based on environment.
func GetDirectoryURL(environment string) string {
	switch environment {
	case "production":
		return LetsEncryptProduction
	case "staging":
		return LetsEncryptStaging
	default:
		return LetsEncryptStaging
	}
}
