package config

import "testing"

func TestLoad_FromEnv(t *testing.T) {
	t.Setenv("PORT", "9000")
	t.Setenv("MONGO_URI", "mongodb://example:27017")
	t.Setenv("MONGO_DB", "db")
	t.Setenv("CORS_ALLOWED_ORIGIN", "http://localhost:3000")
	t.Setenv("MAX_BODY_BYTES", "2048")
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.Port != "9000" || c.MongoDB != "db" || c.MaxBodyBytes != 2048 || c.LogLevel != "info" {
		t.Fatalf("unexpected config %+v", c)
	}
}

func TestLoad_MissingRequired(t *testing.T) {
	t.Setenv("MONGO_URI", "")
	t.Setenv("MONGO_DB", "db")
	t.Setenv("CORS_ALLOWED_ORIGIN", "http://x")
	if _, err := Load(); err == nil {
		t.Fatal("expected error for missing MONGO_URI")
	}
}

func TestLoad_BadMaxBody(t *testing.T) {
	t.Setenv("MONGO_URI", "m")
	t.Setenv("MONGO_DB", "db")
	t.Setenv("CORS_ALLOWED_ORIGIN", "http://x")
	t.Setenv("MAX_BODY_BYTES", "abc")
	if _, err := Load(); err == nil {
		t.Fatal("expected error")
	}
}
