package crypto

import (
	"encoding/base64"
	"testing"
)

func TestAESGCMEncryptionService_Cycle(t *testing.T) {
	key := "1234abcdefghi12341d2v2e31q3d5132"
	svc, err := NewAESGCMEncryptionService(key)
	if err != nil {
		t.Fatalf("failed to create svc: %v", err)
	}

	plaintext := []byte("{\"url\": \"https://example.com\", \"token\": \"secret\"}")
	encrypted, err := svc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("failed to encrypt: %v", err)
	}

	// This is what the repository does
	encoded := base64.StdEncoding.EncodeToString(encrypted)

	// And then back
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("failed to decode b64: %v", err)
	}

	decrypted, err := svc.Decrypt(decoded)
	if err != nil {
		t.Fatalf("failed to decrypt: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("decrypted doesn't match! got %s, want %s", string(decrypted), string(plaintext))
	}
}
