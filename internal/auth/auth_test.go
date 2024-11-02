package auth

import (
	"net/http"
	"testing"
)

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
			wantErr: true,
		},
		{
			name: "Malformed Authorization header",
			headers: http.Header{
				"Authorization": []string{"InvalidBearer token"},
			},
			value:   "",
			wantErr: true,
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
