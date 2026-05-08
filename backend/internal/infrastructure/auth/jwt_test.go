package auth

import (
	"testing"
)

func TestJWTGenerateAndValidate(t *testing.T) {
	svc := NewJWTService("super-secret-1234567890")
	tok, err := svc.GenerateAdminToken(7, "alice", "admin")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	claims, err := svc.ValidateAdminToken(tok)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if claims.UserID != 7 || claims.Username != "alice" || claims.Role != "admin" {
		t.Fatalf("claims mismatch: %+v", claims)
	}
}

func TestJWTValidateBadSignature(t *testing.T) {
	a := NewJWTService("aaaaaaaaaaaaaaaa")
	tok, _ := a.GenerateAdminToken(1, "u", "admin")
	b := NewJWTService("bbbbbbbbbbbbbbbb")
	if _, err := b.ValidateAdminToken(tok); err == nil {
		t.Fatal("expected error for foreign secret")
	}
}

func TestJWTValidateGarbage(t *testing.T) {
	svc := NewJWTService("aaaaaaaaaaaaaaaa")
	if _, err := svc.ValidateAdminToken("not.a.real.token"); err == nil {
		t.Fatal("expected error for malformed token")
	}
}
