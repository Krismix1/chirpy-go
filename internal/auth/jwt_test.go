package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWT(t *testing.T) {
	tokenKey := "random-key"

	tests := []struct {
		name        string
		userID      uuid.UUID
		customKey   string
		customToken string
		expired     bool
		wantErr     bool
	}{
		{
			name:      "Correct token",
			userID:    uuid.New(),
			customKey: "",
			expired:   false,
			wantErr:   false,
		},
		{
			name:      "Expired token",
			userID:    uuid.Nil,
			customKey: "",
			expired:   true,
			wantErr:   true,
		},
		{
			name:      "Wrong key",
			userID:    uuid.Nil,
			customKey: "custom",
			expired:   false,
			wantErr:   true,
		},
		{
			name:        "Invalid token",
			userID:      uuid.Nil,
			customToken: "invalid.token.here",
			customKey:   "",
			expired:     false,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := tokenKey
			if tt.customKey != "" {
				key = tt.customKey
			}
			token, err := MakeJWT(tt.userID, key, 1*time.Minute)
			if tt.customToken != "" {
				token = tt.customToken
			} else if tt.expired {
				token, err = MakeJWT(tt.userID, key, 1*time.Millisecond)
				time.Sleep(2 * time.Millisecond)
			}

			if err != nil {
				t.Fatalf("MakeJWT() error = %v", err)
			}
			userID, err := ValidateJWT(token, tokenKey)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateJWT() error = %v, wantErr = %v, token = %v", err, tt.wantErr, token)
			}

			if !tt.wantErr {
				if userID != tt.userID {
					t.Fatalf("ValidateJWT() %v != %v", tt.userID, userID)
				}
			}
		})
	}
}
