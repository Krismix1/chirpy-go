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
		name      string
		userID    uuid.UUID
		customKey string
		expired   bool
		wantErr   bool
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
			userID:    uuid.New(),
			customKey: "",
			expired:   true,
			wantErr:   true,
		},
		{
			name:      "Wrong key",
			userID:    uuid.New(),
			customKey: "custom",
			expired:   false,
			wantErr:   true,
		},
		// {
		// 	name:     "Password doesn't match different hash",
		// 	password: password1,
		// 	hash:     hash2,
		// 	wantErr:  true,
		// },
		// {
		// 	name:     "Empty password",
		// 	password: "",
		// 	hash:     hash1,
		// 	wantErr:  true,
		// },
		// {
		// 	name:     "Invalid hash",
		// 	password: password1,
		// 	hash:     "invalidhash",
		// 	wantErr:  true,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := tokenKey
			if tt.customKey != "" {
				key = tt.customKey
			}
			token, err := MakeJWT(tt.userID, key, 1*time.Minute)
			if tt.expired {
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

func TestGetBearerToken(t *testing.T) {

	tests := []struct {
		name    string
		headers http.Header
		value   string
		wantErr bool
	}{
		{
			name: "Valid header",
			headers: http.Header{
				"Authorization": []string{"Bearer 123"},
			},
			value:   "123",
			wantErr: false,
		},
		{
			name:    "Missing header",
			headers: http.Header{},
			wantErr: true,
		},
		{
			name: "Duplicated header",
			headers: http.Header{
				"Authorization": []string{"Bearer 123", "Hello"},
			},
			value:   "123",
			wantErr: false,
		},
		{
			name: "Header without prefix",
			headers: http.Header{
				"Authorization": []string{"123"},
			},
			value:   "123",
			wantErr: false, // really?
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GetBearerToken(tt.headers)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetBearerToken() error = %v, wantErr = %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if token != tt.value {
					t.Fatalf("GetBearerToken() expected = %v, actual = %v", tt.value, token)
				}
			}
		})
	}
}
