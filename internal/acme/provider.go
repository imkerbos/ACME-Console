package acme

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
)

// ErrManualDNSPending is returned when challenges are stored and waiting for manual DNS setup
var ErrManualDNSPending = errors.New("manual DNS setup required")

// ChallengeInfo stores information about a pending DNS challenge
type ChallengeInfo struct {
	Domain   string
	TXTHost  string
	TXTValue string
	Token    string
	KeyAuth  string
}

// ManualDNSProvider implements lego's challenge.Provider interface for manual DNS verification.
// It stores challenge information for user retrieval instead of automatically managing DNS records.
type ManualDNSProvider struct {
	mu              sync.RWMutex
	challenges      map[string]ChallengeInfo // keyed by domain
	onPresent       func(info ChallengeInfo) error
	onCleanUp       func(domain string) error
	stopAfterPresent bool // If true, return error after storing challenges to stop lego from waiting
	verifyMode      bool // If true, verify that challenges match expected values instead of storing new ones
	expectedChallenges map[string]ChallengeInfo // Expected challenges in verify mode
}

// Ensure ManualDNSProvider implements challenge.Provider
var _ challenge.Provider = (*ManualDNSProvider)(nil)

// NewManualDNSProvider creates a new ManualDNSProvider.
// onPresent is called when a new challenge is presented (e.g., to store in database).
// onCleanUp is called when a challenge is cleaned up.
func NewManualDNSProvider(onPresent func(ChallengeInfo) error, onCleanUp func(string) error) *ManualDNSProvider {
	return &ManualDNSProvider{
		challenges:       make(map[string]ChallengeInfo),
		onPresent:        onPresent,
		onCleanUp:        onCleanUp,
		stopAfterPresent: false,
	}
}

// NewManualDNSProviderWithStop creates a ManualDNSProvider that stops after storing challenges.
// This is useful for the initial order creation where we just want to get challenges.
func NewManualDNSProviderWithStop(onPresent func(ChallengeInfo) error) *ManualDNSProvider {
	return &ManualDNSProvider{
		challenges:       make(map[string]ChallengeInfo),
		onPresent:        onPresent,
		stopAfterPresent: true,
	}
}

// NewManualDNSProviderWithChallenges creates a ManualDNSProvider with pre-loaded challenges.
// This is useful for FinalizeOrder where we want to verify that CA returns the same challenges.
// If the CA returns different challenges, Present will return an error.
func NewManualDNSProviderWithChallenges(challenges map[string]ChallengeInfo) *ManualDNSProvider {
	return &ManualDNSProvider{
		challenges:         make(map[string]ChallengeInfo),
		verifyMode:         true,
		expectedChallenges: challenges,
		stopAfterPresent:   false,
	}
}

// Present is called by lego when a new challenge needs to be set up.
// Instead of managing DNS records, we store the challenge information for user retrieval.
func (p *ManualDNSProvider) Present(domain, token, keyAuth string) error {
	// Calculate the TXT record value (SHA-256 hash of key authorization, base64url encoded)
	txtValue := computeTXTValue(keyAuth)

	// Get the FQDN for the TXT record
	fqdn := dns01.GetChallengeInfo(domain, keyAuth).FQDN

	// Remove trailing dot if present
	txtHost := strings.TrimSuffix(fqdn, ".")

	info := ChallengeInfo{
		Domain:   domain,
		TXTHost:  txtHost,
		TXTValue: txtValue,
		Token:    token,
		KeyAuth:  keyAuth,
	}

	// If in verify mode, check if the challenge matches expected values
	if p.verifyMode {
		expected, exists := p.expectedChallenges[domain]
		if !exists {
			return fmt.Errorf("unexpected domain in challenge: %s (not in expected challenges)", domain)
		}
		if expected.TXTValue != txtValue {
			return fmt.Errorf("challenge mismatch for domain %s: expected TXT value %s, got %s (CA generated new challenges - this should not happen)",
				domain, expected.TXTValue, txtValue)
		}
		// Challenge matches, store it and continue
		p.mu.Lock()
		p.challenges[domain] = info
		p.mu.Unlock()
		return nil
	}

	// Normal mode: store the challenge
	p.mu.Lock()
	p.challenges[domain] = info
	p.mu.Unlock()

	if p.onPresent != nil {
		if err := p.onPresent(info); err != nil {
			return fmt.Errorf("onPresent callback failed: %w", err)
		}
	}

	// If stopAfterPresent is true, return an error to stop lego from waiting for DNS
	if p.stopAfterPresent {
		return ErrManualDNSPending
	}

	return nil
}

// CleanUp is called by lego after verification is complete.
func (p *ManualDNSProvider) CleanUp(domain, token, keyAuth string) error {
	p.mu.Lock()
	delete(p.challenges, domain)
	p.mu.Unlock()

	if p.onCleanUp != nil {
		if err := p.onCleanUp(domain); err != nil {
			return fmt.Errorf("onCleanUp callback failed: %w", err)
		}
	}

	return nil
}

// GetChallenge returns the challenge information for a specific domain.
func (p *ManualDNSProvider) GetChallenge(domain string) (ChallengeInfo, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	info, ok := p.challenges[domain]
	return info, ok
}

// GetAllChallenges returns all pending challenges.
func (p *ManualDNSProvider) GetAllChallenges() []ChallengeInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()
	result := make([]ChallengeInfo, 0, len(p.challenges))
	for _, info := range p.challenges {
		result = append(result, info)
	}
	return result
}

// computeTXTValue calculates the DNS TXT record value from the key authorization.
// This is the SHA-256 hash of the key authorization, base64url encoded without padding.
func computeTXTValue(keyAuth string) string {
	hash := sha256.Sum256([]byte(keyAuth))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

// GetTXTRecordName returns the full TXT record name for a domain.
// For wildcard domains, it returns the record for the base domain.
func GetTXTRecordName(domain string) string {
	// Remove wildcard prefix if present
	baseDomain := strings.TrimPrefix(domain, "*.")
	return "_acme-challenge." + baseDomain
}
