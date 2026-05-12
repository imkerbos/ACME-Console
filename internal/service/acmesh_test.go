package service

import (
	"strings"
	"testing"

	"github.com/imkerbos/ACME-Console/internal/model"
)

func TestAcmeShService_GenerateChallenges(t *testing.T) {
	svc := NewAcmeShService()

	tests := []struct {
		name    string
		certID  uint
		domains []string
		wantLen int
	}{
		{
			name:    "single domain",
			certID:  1,
			domains: []string{"example.com"},
			wantLen: 1,
		},
		{
			name:    "multiple domains",
			certID:  2,
			domains: []string{"example.com", "www.example.com"},
			wantLen: 2,
		},
		{
			name:    "wildcard domain",
			certID:  3,
			domains: []string{"*.example.com"},
			wantLen: 1,
		},
		{
			name:    "mixed domains",
			certID:  4,
			domains: []string{"example.com", "*.example.com"},
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			challenges, err := svc.GenerateChallenges(tt.certID, tt.domains)
			if err != nil {
				t.Fatalf("GenerateChallenges() error = %v", err)
			}

			if len(challenges) != tt.wantLen {
				t.Errorf("GenerateChallenges() got %d challenges, want %d", len(challenges), tt.wantLen)
			}

			for i, ch := range challenges {
				if ch.CertificateID != tt.certID {
					t.Errorf("Challenge[%d].CertificateID = %d, want %d", i, ch.CertificateID, tt.certID)
				}
				if ch.Domain != tt.domains[i] {
					t.Errorf("Challenge[%d].Domain = %s, want %s", i, ch.Domain, tt.domains[i])
				}
				if !strings.HasPrefix(ch.TXTHost, "_acme-challenge.") {
					t.Errorf("Challenge[%d].TXTHost = %s, should start with _acme-challenge.", i, ch.TXTHost)
				}
				if ch.TXTValue == "" {
					t.Errorf("Challenge[%d].TXTValue should not be empty", i)
				}
				if ch.Status != model.ChallengeStatusPending {
					t.Errorf("Challenge[%d].Status = %s, want %s", i, ch.Status, model.ChallengeStatusPending)
				}
			}
		})
	}
}

func TestAcmeShService_GenerateChallenges_WildcardTXTHost(t *testing.T) {
	svc := NewAcmeShService()

	challenges, err := svc.GenerateChallenges(1, []string{"*.example.com"})
	if err != nil {
		t.Fatalf("GenerateChallenges() error = %v", err)
	}

	if len(challenges) != 1 {
		t.Fatalf("Expected 1 challenge, got %d", len(challenges))
	}

	// Wildcard should have TXT host without the *. prefix
	expected := "_acme-challenge.example.com"
	if challenges[0].TXTHost != expected {
		t.Errorf("TXTHost = %s, want %s", challenges[0].TXTHost, expected)
	}
}

func TestAcmeShService_ExportTXTTemplate(t *testing.T) {
	svc := NewAcmeShService()

	challenges := []model.Challenge{
		{
			Domain:   "example.com",
			TXTHost:  "_acme-challenge.example.com",
			TXTValue: "test-token-1",
		},
		{
			Domain:   "*.example.com",
			TXTHost:  "_acme-challenge.example.com",
			TXTValue: "test-token-2",
		},
	}

	template := svc.ExportTXTTemplate(challenges)

	if !strings.Contains(template, "DNS TXT Records") {
		t.Error("Template should contain header")
	}
	if !strings.Contains(template, "_acme-challenge.example.com") {
		t.Error("Template should contain TXT host")
	}
	if !strings.Contains(template, "test-token-1") {
		t.Error("Template should contain first token")
	}
	if !strings.Contains(template, "test-token-2") {
		t.Error("Template should contain second token")
	}
}

func TestAcmeShService_VerifyChallenges(t *testing.T) {
	svc := NewAcmeShService()

	cert := &model.Certificate{
		ID:     1,
		Status: model.CertificateStatusPending,
	}

	verified, err := svc.VerifyChallenges(cert)
	if err != nil {
		t.Fatalf("VerifyChallenges() error = %v", err)
	}

	// MVP implementation always returns true
	if !verified {
		t.Error("VerifyChallenges() should return true in MVP")
	}
}
