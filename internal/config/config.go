package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port             string
	DatabaseURL      string
	JWTSecret        string
	JWTExpireHours   int
	BPJSHealthCap    float64
	BPJSHealthWorker float64
	BPJSHealthEmploy float64
	JPCap            float64
	JPWorker         float64
	JPEmployer       float64
	JHTWorker        float64
	JHTEmployer      float64
	JKMEmployer      float64
	DefaultAdminUser string
	DefaultAdminPass string
	SeedDemo         bool
	DemoUserPass     string
}

// Load membaca seluruh konfigurasi aplikasi dari environment variable.
func Load() *Config {
	return &Config{
		Port:             getEnv("PORT", "8080"),
		DatabaseURL:      getEnv("DATABASE_URL", "host=localhost user=payroll password=payroll dbname=payroll port=5432 sslmode=disable TimeZone=Asia/Jakarta"),
		JWTSecret:        getEnv("JWT_SECRET", "change-me-in-production"),
		JWTExpireHours:   getEnvInt("JWT_EXPIRE_HOURS", 24),
		BPJSHealthCap:    getEnvFloat("BPJS_HEALTH_CAP", 12_000_000),
		BPJSHealthWorker: getEnvFloat("BPJS_HEALTH_WORKER", 0.01),
		BPJSHealthEmploy: getEnvFloat("BPJS_HEALTH_EMPLOYER", 0.04),
		JPCap:            getEnvFloat("JP_CAP", 10_042_300),
		JPWorker:         getEnvFloat("JP_WORKER", 0.01),
		JPEmployer:       getEnvFloat("JP_EMPLOYER", 0.02),
		JHTWorker:        getEnvFloat("JHT_WORKER", 0.02),
		JHTEmployer:      getEnvFloat("JHT_EMPLOYER", 0.037),
		JKMEmployer:      getEnvFloat("JKM_EMPLOYER", 0.003),
		DefaultAdminUser: getEnv("DEFAULT_ADMIN_USER", "admin"),
		DefaultAdminPass: getEnv("DEFAULT_ADMIN_PASS", "admin123"),
		SeedDemo:         getEnvBool("SEED_DEMO", false),
		DemoUserPass:     getEnv("DEMO_USER_PASS", "rahasia123"),
	}
}

// getEnv membaca env string dengan nilai fallback.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// getEnvInt membaca env integer dengan nilai fallback.
func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

// getEnvFloat membaca env float dengan nilai fallback.
func getEnvFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseFloat(v, 64); err == nil {
			return n
		}
	}
	return fallback
}

// getEnvBool membaca env boolean dengan nilai fallback.
func getEnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}
