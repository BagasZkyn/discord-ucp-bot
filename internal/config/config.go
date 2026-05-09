package config

import (
	"os"
	"strings"
)

// Config menyimpan semua konfigurasi aplikasi
type Config struct {
	// Discord
	DiscordToken string
	UCPRoleID    string
	AdminUserIDs []string

	// Server
	ServerName string
	LogoURL    string

	// SAMP
	SAMPHost string
	SAMPPort string

	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
}

// Load membaca konfigurasi dari environment variables
func Load() *Config {
	return &Config{
		DiscordToken: getEnv("DISCORD_TOKEN", ""),
		UCPRoleID:    getEnv("UCP_ROLE_ID", ""),
		AdminUserIDs: splitEnv("ADMIN_USER_IDS"),

		ServerName: getEnv("SERVER_NAME", "Djava Roleplay"),
		LogoURL:    getEnv("LOGO_URL", "https://cdn.discordapp.com/attachments/1153557595928408163/1497964536093868174/Untitled_design.png"),

		SAMPHost: getEnv("SAMP_HOST", "82.25.36.26"),
		SAMPPort: getEnv("SAMP_PORT", "7043"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "root"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "samp_db"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

// splitEnv membaca env variable berisi daftar ID dipisah koma
func splitEnv(key string) []string {
	val := os.Getenv(key)
	if val == "" {
		return nil
	}
	parts := strings.Split(val, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
