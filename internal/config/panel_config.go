package config

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)

const panelConfigPath = "panel_config.json"

// PanelConfig menyimpan ID pesan UCP panel yang persisten
type PanelConfig struct {
	UCPPanelChannelID string `json:"ucpPanelChannelId,omitempty"`
	UCPPanelMessageID string `json:"ucpPanelMessageId,omitempty"`
}

var panelMutex sync.RWMutex

// LoadPanelConfig membaca panel_config.json
func LoadPanelConfig() PanelConfig {
	panelMutex.RLock()
	defer panelMutex.RUnlock()

	data, err := os.ReadFile(panelConfigPath)
	if err != nil {
		return PanelConfig{}
	}

	var cfg PanelConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return PanelConfig{}
	}
	return cfg
}

// SavePanelConfig menyimpan perubahan ke panel_config.json
func SavePanelConfig(cfg PanelConfig) {
	panelMutex.Lock()
	defer panelMutex.Unlock()

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		log.Printf("❌ Gagal marshal panel config: %v", err)
		return
	}

	if err := os.WriteFile(panelConfigPath, data, 0644); err != nil {
		log.Printf("❌ Gagal menyimpan panel config: %v", err)
	}
}
