package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config menampung semua konfigurasi untuk aplikasi.
type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	ServerPort string
	JWTSecret  string
}

// LoadConfig membaca konfigurasi dari environment variables.
// Fungsi ini akan memuat dari file .env jika ada.
func LoadConfig() (*Config, error) {
	// Memuat file .env dari root directory.
	// Tidak masalah jika file tidak ada, os.Getenv akan digunakan sebagai fallback.
	_ = godotenv.Load()

	cfg := &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "user"),
		DBPassword: getEnv("DB_PASSWORD", "password"),
		DBName:     getEnv("DB_NAME", "blog_db"),
		ServerPort: getEnv("SERVER_PORT", "8080"),
		JWTSecret:  getEnv("JWT_SECRET", "default-secret"),
	}

	return cfg, nil
}

// DBConnectionString membangun string koneksi database dari config.
func (c *Config) DBConnectionString() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName)
}

// getEnv membaca environment variable atau mengembalikan nilai default.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
