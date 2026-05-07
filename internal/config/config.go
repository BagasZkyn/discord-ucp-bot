package config

import "os"

// Config menyimpan semua konfigurasi aplikasi
type Config struct {
	// Discord
	DiscordToken string
	UCPRoleID    string

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
