package secrets

import (
	"math"
	"regexp"
	"strings"
)

func IsSensitiveKey(key string) bool {
	lowerKey := strings.ToLower(key)

	// Strong signals: substring match is fine, these are unambiguous
	strongSignals := []string{
		"password", "passwd", "secret", "private_key",
		"client_secret", "access_token", "refresh_token",
		"id_token", "auth_token", "api_token", "bearer",
		"credential", "certificate", "encryption_key",
		"signing_key", "hmac", "oauth",
	}
	for _, word := range strongSignals {
		if strings.Contains(lowerKey, word) {
			return true
		}
	}

	parts := strings.Split(lowerKey, "_")
	partSet := make(map[string]bool, len(parts))
	for _, p := range parts {
		partSet[p] = true
	}

	weakSignals := map[string]bool{
		"key": true, "secret": true, "token": true,
		"auth": true, "jwt": true, "cert": true,
		"pem": true, "pkcs": true, "sig": true,
		"pwd": true, "pass": true, "pin": true,
		"seed": true, "salt": true, "nonce": true,
		"session": true, "cookie": true,
	}

	for _, part := range parts {
		if weakSignals[part] {
			return true
		}
	}

	return false
}

func IsSensitiveValue(value string) bool {
	if len(value) == 0 {
		return false
	}

	// if the entropy is high,then yeah its probably a secret (API keys, tokens, hashes)
	if len(value) >= 16 && shannonEntropy(value) > 4.5 {
		return true
	}

	// these are regex for Known secret formats
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`^[A-Za-z0-9+/]{40,}={0,2}$`),           // base64 blobs
		regexp.MustCompile(`^[0-9a-fA-F]{32,}$`),                   // hex hashes/tokens
		regexp.MustCompile(`-----BEGIN [A-Z ]+-----`),              // PEM keys
		regexp.MustCompile(`^(ghp|gho|ghu|ghs|ghr)_[A-Za-z0-9]+$`), // GitHub tokens
		regexp.MustCompile(`^sk-[A-Za-z0-9]{32,}$`),                // OpenAI keys
		regexp.MustCompile(`^[A-Z0-9]{20}:[A-Za-z0-9_-]{40}$`),     // AWS-style keys
		regexp.MustCompile(`postgres://[^:]+:[^@]+@`),              // DSNs with credentials
		regexp.MustCompile(`^sk_(live|test|prod)_[A-Za-z0-9]+$`),   // Stripe-style keys
		regexp.MustCompile(`mysql://[^:]+:[^@]+@`),
		regexp.MustCompile(`mongodb(\+srv)?://[^:]+:[^@]+@`),
	}
	for _, re := range patterns {
		if re.MatchString(value) {
			return true
		}
	}

	return false
}

func shannonEntropy(s string) float64 {
	freq := make(map[rune]float64)
	for _, c := range s {
		freq[c]++
	}
	entropy := 0.0
	l := float64(len(s))
	for _, count := range freq {
		p := count / l
		entropy -= p * math.Log2(p)
	}
	return entropy
}
func IsRedacted(key, value string) bool {
	return IsSensitiveKey(key) || IsSensitiveValue(value)
}
