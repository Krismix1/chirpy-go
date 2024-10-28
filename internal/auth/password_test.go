package auth

import (
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashAndCheckSame(t *testing.T) {
	password1 := "hello"
	hashed, err := HashPassword(password1)
	if err != nil {
		t.Fatalf("Failed to hash: %s", err)
	}

	password2 := "hello"
	err = CheckPasswordHash(password2, hashed)
	if err != nil {
		t.Fatalf("Expected password %s to match %s: %s", password2, password1, err)
	}
}

func TestHashAndCheckDiffer(t *testing.T) {
	password1 := "hello"
	hashed, err := HashPassword(password1)
	if err != nil {
		t.Fatalf("Failed to hash: %s", err)
	}

	password2 := "hello2"
	err = CheckPasswordHash(password2, hashed)
	if err == nil {
		t.Fatalf("Expected password %s to not match %s", password2, password1)
	}

	if !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		t.Fatalf("Expected to fail due to mismatch: %s", err)
	}
}
