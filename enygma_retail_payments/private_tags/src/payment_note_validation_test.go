package tags

import (
	"math/big"
	"strings"
	"testing"
)

func TestDecryptPaymentNote_PlaintextLength(t *testing.T) {
	channelSS := []byte(strings.Repeat("A", 32))
	recipientPkSpend := big.NewInt(12345)
	blockNumber := uint64(100)

	tests := []struct {
		name      string
		length    int
		wantError bool
	}{
		{"95_bytes_rejected", 95, true},
		{"96_bytes_accepted", 96, false},
		{"97_bytes_rejected", 97, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plaintext := make([]byte, tt.length)
			key := DeriveBlockKey(channelSS, blockNumber, recipientPkSpend)

			ct, err := EncryptPayload(key, plaintext)
			if err != nil {
				t.Fatalf("EncryptPayload failed: %v", err)
			}

			note, err := DecryptPaymentNote(
				channelSS, recipientPkSpend, blockNumber, ct,
			)

			if tt.wantError {
				if err == nil {
					t.Fatal("expected invalid plaintext length to be rejected")
				}
				if !strings.Contains(err.Error(), "unexpected plaintext length") {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("valid 96-byte plaintext rejected: %v", err)
			}
			if note == nil {
				t.Fatal("expected a payment note")
			}
		})
	}
}