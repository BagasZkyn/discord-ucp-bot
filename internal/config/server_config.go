package config

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)

const serverConfigPath = "server_config.json"

// ServerConfig menyimpan konfigurasi SAMP server yang bisa diedit via Discord
type ServerConfig struct {
	Host string `json:"host"`
	Port string `json:"port"`
	Name string `json:"name"`
}

var serverMutex sync.RWMutex

// LoadServerConfig membaca server_config.json, fallback ke env jika tidak ada
func LoadServerConfig(cfg *Config) ServerConfig {
	serverMutex.RLock()
	defer serverMutex.RUnlock()

	data, err := os.ReadFile(serverConfigPath)
	if err != nil {
		return ServerConfig{
			Host: cfg.SAMPHost,
			Port: cfg.SAMPPort,
			Name: cfg.ServerName,
		}
	}

	var sc ServerConfig
	if err := json.Unmarshal(data, &sc); err != nil {
		return ServerConfig{
			Host: cfg.SAMPHost,
			Port: cfg.SAMPPort,
			Name: cfg.ServerName,
		}
	}

	// Fallback ke env jika field kosong
	if sc.Host == "" {
		sc.Host = cfg.SAMPHost
	}
	if sc.Port == "" {
		sc.Port = cfg.SAMPPort
	}
	if sc.Name == "" {
		sc.Name = cfg.ServerName
	}

	return sc
}

// SaveServerConfig menyimpan perubahan ke server_config.json
func SaveServerConfig(sc ServerConfig) error {
	serverMutex.Lock()
	defer serverMutex.Unlock()

	data, err := json.MarshalIndent(sc, "", "  ")
	if err != nil {
		log.Printf("❌ Gagal marshal server config: %v", err)
		return err
	}

	if err := os.WriteFile(serverConfigPath, data, 0644); err != nil {
		log.Printf("❌ Gagal menyimpan server config: %v", err)
		return err
	}
	return nil
}
