// Package config loads runtime configuration from environment variables.
package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port              string
	MongoURI          string
	MongoDB           string
	CORSAllowedOrigin string
	MaxBodyBytes      int64
	LogLevel          string
}

func Load() (Config, error) {
	c := Config{
		Port:              getenv("PORT", "8080"),
		MongoURI:          os.Getenv("MONGO_URI"),
		MongoDB:           os.Getenv("MONGO_DB"),
		CORSAllowedOrigin: os.Getenv("CORS_ALLOWED_ORIGIN"),
		MaxBodyBytes:      1 << 20,
		LogLevel:          getenv("LOG_LEVEL", "info"),
	}
	for name, v := range map[string]string{
		"MONGO_URI":           c.MongoURI,
		"MONGO_DB":            c.MongoDB,
		"CORS_ALLOWED_ORIGIN": c.CORSAllowedOrigin,
	} {
		if v == "" {
			return Config{}, fmt.Errorf("missing required environment variable %s", name)
		}
	}
	if v := os.Getenv("MAX_BODY_BYTES"); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil || n <= 0 {
			return Config{}, fmt.Errorf("MAX_BODY_BYTES must be a positive integer")
		}
		c.MaxBodyBytes = n
	}
	return c, nil
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
