package db

import (
	"net/url"
	"testing"

	"prodlich/internal/model"
)

// Regression: postgresURL must percent-encode user/password/db so that
// reserved characters do not corrupt the connection string.
//
// The previous fmt.Sprintf-based version would emit
//
//	postgres://user:p@ss@host:5432/db?sslmode=disable
//
// which net/url parses as user="user", password="p" and host="ss@host" —
// silently routing the migration to the wrong host with empty creds.
func TestPostgresURL_EscapesReservedCharsInUserAndPassword(t *testing.T) {
	cfg := &model.DB{
		Host:     "db.internal",
		Port:     "5432",
		User:     "user@svc",
		Password: "p@:ss/word?",
		DBName:   "gateway",
		SSLMode:  "require",
	}

	got := postgresURL(cfg)

	u, err := url.Parse(got)
	if err != nil {
		t.Fatalf("postgresURL produced unparseable URL %q: %v", got, err)
	}

	if u.Scheme != "postgres" {
		t.Errorf("scheme=%q, want postgres", u.Scheme)
	}
	if u.Host != "db.internal:5432" {
		t.Errorf("host=%q, want db.internal:5432 (a malformed escape would push parts into the host)", u.Host)
	}
	if user := u.User.Username(); user != "user@svc" {
		t.Errorf("username=%q, want user@svc (decoded)", user)
	}
	pwd, ok := u.User.Password()
	if !ok {
		t.Fatalf("password missing from URL %q", got)
	}
	if pwd != "p@:ss/word?" {
		t.Errorf("password=%q, want p@:ss/word? (decoded)", pwd)
	}
	if u.Path != "/gateway" {
		t.Errorf("path=%q, want /gateway", u.Path)
	}
	if got := u.Query().Get("sslmode"); got != "require" {
		t.Errorf("sslmode=%q, want require", got)
	}
}

func TestPostgresURL_NoPortOmitsColon(t *testing.T) {
	cfg := &model.DB{
		Host:     "localhost",
		Port:     "",
		User:     "u",
		Password: "p",
		DBName:   "db",
		SSLMode:  "disable",
	}
	got := postgresURL(cfg)
	u, err := url.Parse(got)
	if err != nil {
		t.Fatalf("unparseable URL %q: %v", got, err)
	}
	if u.Host != "localhost" {
		t.Errorf("host=%q, want localhost (no trailing colon when port is empty)", u.Host)
	}
}

func TestPostgresURL_OmitsSSLModeWhenEmpty(t *testing.T) {
	cfg := &model.DB{
		Host:     "h",
		Port:     "1",
		User:     "u",
		Password: "p",
		DBName:   "d",
		SSLMode:  "",
	}
	got := postgresURL(cfg)
	u, err := url.Parse(got)
	if err != nil {
		t.Fatalf("unparseable URL %q: %v", got, err)
	}
	if u.Query().Has("sslmode") {
		t.Errorf("sslmode should not be set when empty, got %q", got)
	}
}
