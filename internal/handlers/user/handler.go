package user

import (
	"database/sql"
	"log"
	"time"

	"botucp/internal/config"
	"botucp/internal/utils"

	"github.com/bwmarrin/discordgo"
)

// Handler mengelola semua interaksi user panel
type Handler struct {
	cfg *config.Config
	db  *sql.DB
}

// NewHandler membuat instance Handler baru
func NewHandler(cfg *config.Config, db *sql.DB) *Handler {
	return &Handler{cfg: cfg, db: db}
}

// embed adalah shorthand untuk membuat embed UCP
func (h *Handler) embed(title, description string, color int, fields ...*utils.EmbedField) *discordgo.MessageEmbed {
	return utils.BuildEmbed(h.cfg.LogoURL, h.cfg.ServerName, title, description, color, fields)
}

// hasAdminAccess mengecek apakah user ID ada di daftar ADMIN_USER_IDS
func (h *Handler) hasAdminAccess(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	for _, id := range h.cfg.AdminUserIDs {
		if id == m.Author.ID {
			return true
		}
	}
	return false
}

// ─────────────────────────────────────────
//  !panel command
// ─────────────────────────────────────────

// HandlePanelCommand menangani perintah !panel
func (h *Handler) HandlePanelCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Cek permission: harus punya AdminRoleID atau PermissionManageServer
	if !h.hasAdminAccess(s, m) {
		s.ChannelMessageSendEmbed(m.ChannelID, h.embed(
			"「 ⛔ 」Akses Ditolak",
			"Otorisasi tidak mencukupi untuk mengeksekusi perintah ini.",
			utils.ColorError,
		))
		return
	}

	cfg := config.LoadPanelConfig()
	panelData := buildPanel(h.cfg)

	var panelMsgID string

	// Coba edit pesan lama
	if cfg.UCPPanelChannelID != "" && cfg.UCPPanelMessageID != "" {
		_, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
			Channel:    cfg.UCPPanelChannelID,
			ID:         cfg.UCPPanelMessageID,
			Embeds:     panelData.Embeds,
			Components: panelData.Components,
		})
		if err == nil {
			panelMsgID = cfg.UCPPanelMessageID
			log.Printf("🟢  UCP Panel di-refresh (edit) — %s", panelMsgID)
		}
	}

	// Kirim pesan baru jika belum ada atau gagal edit
	if panelMsgID == "" {
		msg, err := s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
			Embeds:     panelData.Embeds,
			Components: panelData.Components,
		})
		if err != nil {
			log.Printf("❌ Gagal kirim UCP Panel: %v", err)
			return
		}
		panelMsgID = msg.ID
		config.SavePanelConfig(config.PanelConfig{
			UCPPanelChannelID: m.ChannelID,
			UCPPanelMessageID: msg.ID,
		})
		log.Printf("🟢  UCP Panel baru dibuat — %s", msg.ID)
	}

	s.ChannelMessageDelete(m.ChannelID, m.ID)
}

// ─────────────────────────────────────────
//  Build Panel
// ─────────────────────────────────────────

type panelPayload struct {
	Embeds     []*discordgo.MessageEmbed
	Components []discordgo.MessageComponent
}

func buildPanel(cfg *config.Config) panelPayload {
	embed := &discordgo.MessageEmbed{
		Color: utils.ColorPrimary,
		Author: &discordgo.MessageEmbedAuthor{
			Name:    "SERVER MANAGEMENT SYSTEM",
			IconURL: cfg.LogoURL,
		},
		Title: "💠 USER CONTROL PANEL",
		Description: "Selamat datang di **" + cfg.ServerName + "** Control Panel.\n" +
			"Panel ini berfungsi sebagai pusat kontrol akun in-game Anda secara real-time. Gunakan menu interaktif di bawah ini untuk mengelola data Anda dengan aman.\n\n" +
			"```ansi\n\u001b[1;32m✓ SYSTEM ONLINE & SECURED\u001b[0m\n```",
		Fields: []*discordgo.MessageEmbedField{
			{
				Name: "✨ \u200b \u200bLayanan Tersedia",
				Value: "**[ 📝 ] Daftar UCP**\n└ Buat identitas baru untuk memulai petualanganmu.\n\n" +
					"**[ 🔄 ] Sync Role**\n└ Hubungkan otorisasi Discord ke akun in-game.\n\n" +
					"**[ 📊 ] Cek Status**\n└ Lihat informasi detail mengenai akunmu saat ini.\n\n" +
					"**[ 📨 ] Resend PIN**\n└ Dapatkan kembali PIN rahasia via Direct Message.\n\n" +
					"**[ 🔑 ] Reset Password**\n└ Atur ulang kata sandimu jika kamu melupakannya.",
				Inline: false,
			},
			{
				Name: "🛡️ \u200b \u200bProtokol Keamanan",
				Value: ">>> **1.** Nama identitas hanya boleh terdiri dari huruf & angka.\n" +
					"**2.** Panjang karakter minimal **3** hingga maksimal **20** digit.\n" +
					"**3.** Satu akun Discord hanya dapat memiliki **satu** identitas UCP.\n" +
					"**4.** Tidak diperbolehkan menggunakan spasi atau simbol khusus (`_`).",
				Inline: false,
			},
		},
		Image: &discordgo.MessageEmbedImage{URL: cfg.LogoURL},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    cfg.ServerName + " © 2026 • Protected by Security System",
			IconURL: cfg.LogoURL,
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	row1 := discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				CustomID: "btn_register",
				Label:    "Daftar UCP",
				Emoji:    discordgo.ComponentEmoji{Name: "📝"},
				Style:    discordgo.SuccessButton,
			},
			discordgo.Button{
				CustomID: "btn_refund_role",
				Label:    "Sync Role",
				Emoji:    discordgo.ComponentEmoji{Name: "🔄"},
				Style:    discordgo.SecondaryButton,
			},
			discordgo.Button{
				CustomID: "btn_check_status",
				Label:    "Cek Status",
				Emoji:    discordgo.ComponentEmoji{Name: "📊"},
				Style:    discordgo.SecondaryButton,
			},
		},
	}

	row2 := discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				CustomID: "btn_resend_pin",
				Label:    "Resend PIN",
				Emoji:    discordgo.ComponentEmoji{Name: "📨"},
				Style:    discordgo.PrimaryButton,
			},
			discordgo.Button{
				CustomID: "btn_reset_password",
				Label:    "Reset Password",
				Emoji:    discordgo.ComponentEmoji{Name: "🔑"},
				Style:    discordgo.DangerButton,
			},
		},
	}

	return panelPayload{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: []discordgo.MessageComponent{row1, row2},
	}
}
