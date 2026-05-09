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

// hasAdminAccess mengecek apakah user boleh pakai !panel.
// Prioritas:
//  1. Jika ADMIN_USER_IDS diset di .env → hanya user ID tersebut yang boleh
//  2. Jika ADMIN_USER_IDS kosong → fallback ke cek permission Administrator/ManageServer di guild
func (h *Handler) hasAdminAccess(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	if len(h.cfg.AdminUserIDs) > 0 {
		for _, id := range h.cfg.AdminUserIDs {
			if id == m.Author.ID {
				return true
			}
		}
		return false
	}

	guild, err := s.Guild(m.GuildID)
	if err != nil {
		log.Printf("⚠️  Gagal fetch guild untuk cek permission: %v", err)
		return false
	}

	if guild.OwnerID == m.Author.ID {
		return true
	}

	member, err := s.GuildMember(m.GuildID, m.Author.ID)
	if err != nil {
		log.Printf("⚠️  Gagal fetch member untuk cek permission: %v", err)
		return false
	}

	var totalPerms int64
	for _, memberRoleID := range member.Roles {
		for _, guildRole := range guild.Roles {
			if guildRole.ID == memberRoleID {
				totalPerms |= int64(guildRole.Permissions)
				break
			}
		}
	}

	return totalPerms&int64(discordgo.PermissionAdministrator) != 0 ||
		totalPerms&int64(discordgo.PermissionManageServer) != 0
}

// ─────────────────────────────────────────
//  !panel command
// ─────────────────────────────────────────

func (h *Handler) HandlePanelCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	if !h.hasAdminAccess(s, m) {
		s.ChannelMessageSendEmbed(m.ChannelID, h.embed(
			"Akses Ditolak",
			"Kamu tidak memiliki izin untuk menjalankan perintah ini.",
			utils.ColorError,
		))
		return
	}

	cfg := config.LoadPanelConfig()
	panelData := buildPanel(h.cfg)

	var panelMsgID string

	if cfg.UCPPanelChannelID != "" && cfg.UCPPanelMessageID != "" {
		_, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
			Channel:    cfg.UCPPanelChannelID,
			ID:         cfg.UCPPanelMessageID,
			Embeds:     panelData.Embeds,
			Components: panelData.Components,
		})
		if err == nil {
			panelMsgID = cfg.UCPPanelMessageID
			log.Printf("🟢  UCP Panel di-refresh — %s", panelMsgID)
		}
	}

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
			Name:    cfg.ServerName + " — User Control Panel",
			IconURL: cfg.LogoURL,
		},
		Title:       "Selamat Datang",
		Description: "Kelola akun in-game kamu melalui panel di bawah ini.\nSemua perubahan berlaku secara langsung.",
		Fields: []*discordgo.MessageEmbedField{
			{
				Name: "Layanan",
				Value: "📝 **Daftar UCP** — Buat akun baru\n" +
					"🔄 **Sync Role** — Pulihkan role Discord\n" +
					"📊 **Cek Status** — Lihat info akunmu\n" +
					"📨 **Resend PIN** — Kirim ulang PIN via DM\n" +
					"🔑 **Reset Password** — Ganti password akun",
				Inline: false,
			},
			{
				Name: "Ketentuan",
				Value: "Username hanya boleh huruf dan angka, 3–20 karakter, tanpa spasi atau simbol.",
				Inline: false,
			},
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: cfg.LogoURL},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    cfg.ServerName,
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
