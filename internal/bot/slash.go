package bot

import (
	"fmt"
	"log"
	"time"

	"botucp/internal/config"
	"botucp/internal/utils"

	"github.com/bwmarrin/discordgo"
)

// slashCommands mendefinisikan semua slash command yang didaftarkan saat bot start
var slashCommands = []*discordgo.ApplicationCommand{
	{
		Name:        "ping",
		Description: "Cek latensi bot",
	},
	{
		Name:        "serverinfo",
		Description: "Lihat info dan status server SA-MP",
	},
	{
		Name:        "ucpinfo",
		Description: "Lihat info akun UCP kamu",
	},
	{
		Name:        "setserver",
		Description: "Ubah konfigurasi server SA-MP (admin only)",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "host",
				Description: "IP atau hostname server",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "port",
				Description: "Port server (default: 7777)",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "name",
				Description: "Nama server",
				Required:    false,
			},
		},
	},
}

// registerSlashCommands mendaftarkan semua slash command ke Discord
func (b *Bot) registerSlashCommands() {
	for _, cmd := range slashCommands {
		_, err := b.session.ApplicationCommandCreate(b.session.State.User.ID, "", cmd)
		if err != nil {
			log.Printf("❌ Gagal daftarkan command /%s: %v", cmd.Name, err)
		} else {
			log.Printf("✅ Slash command /%s terdaftar", cmd.Name)
		}
	}
}

// handleSlashCommand mendispatch slash command ke handler yang sesuai
func (b *Bot) handleSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	name := i.ApplicationCommandData().Name
	switch name {
	case "ping":
		b.handlePing(s, i)
	case "serverinfo":
		b.handleServerInfo(s, i)
	case "ucpinfo":
		b.handleUCPInfo(s, i)
	case "setserver":
		b.handleSetServer(s, i)
	}
}

// ─────────────────────────────────────────
//  /ping
// ─────────────────────────────────────────

func (b *Bot) handlePing(s *discordgo.Session, i *discordgo.InteractionCreate) {
	start := time.Now()
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral},
	})
	latency := time.Since(start).Milliseconds()
	wsLatency := s.HeartbeatLatency().Milliseconds()

	embed := utils.BuildEmbed(b.cfg.LogoURL, b.cfg.ServerName,
		"Pong!",
		"",
		utils.ColorPrimary,
		[]*utils.EmbedField{
			utils.Field("Respons", fmt.Sprintf("`%d ms`", latency), true),
			utils.Field("WebSocket", fmt.Sprintf("`%d ms`", wsLatency), true),
		},
	)

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	})
}

// ─────────────────────────────────────────
//  /serverinfo
// ─────────────────────────────────────────

func (b *Bot) handleServerInfo(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	sc := config.LoadServerConfig(b.cfg)
	info, err := querySAMP(sc.Host, sc.Port)

	var embed *discordgo.MessageEmbed
	if err != nil {
		embed = utils.BuildEmbed(b.cfg.LogoURL, b.cfg.ServerName,
			"Info Server",
			"Server sedang offline atau tidak dapat dijangkau.",
			utils.ColorError,
			[]*utils.EmbedField{
				utils.Field("Nama", sc.Name, true),
				utils.Field("Alamat", fmt.Sprintf("`%s:%s`", sc.Host, sc.Port), true),
				utils.Field("Status", "Offline", false),
			},
		)
	} else {
		embed = utils.BuildEmbed(b.cfg.LogoURL, b.cfg.ServerName,
			"Info Server",
			"",
			utils.ColorSuccess,
			[]*utils.EmbedField{
				utils.Field("Nama", sc.Name, true),
				utils.Field("Alamat", fmt.Sprintf("`%s:%s`", sc.Host, sc.Port), true),
				utils.Field("Pemain", fmt.Sprintf("%d / %d", info.Online, info.MaxPlayers), true),
				utils.Field("Status", "Online", true),
			},
		)
	}

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	})
}

// ─────────────────────────────────────────
//  /ucpinfo
// ─────────────────────────────────────────

func (b *Bot) handleUCPInfo(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral},
	})

	discordID := i.Member.User.ID

	var ucpName string
	var adminLevel int
	var registerDate int64
	err := b.db.QueryRow(
		"SELECT `username`, `admin`, `registerdate` FROM `ucp` WHERE `DiscordID` = ? LIMIT 1",
		discordID,
	).Scan(&ucpName, &adminLevel, &registerDate)

	if err != nil {
		embed := utils.BuildEmbed(b.cfg.LogoURL, b.cfg.ServerName,
			"Akun Tidak Ditemukan",
			"Kamu belum terdaftar di UCP. Gunakan panel untuk mendaftar.",
			utils.ColorWarning,
			nil,
		)
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{embed},
		})
		return
	}

	hasRole := false
	if i.Member != nil {
		for _, roleID := range i.Member.Roles {
			if roleID == b.cfg.UCPRoleID {
				hasRole = true
				break
			}
		}
	}

	roleStatus := "Tidak aktif"
	if hasRole {
		roleStatus = "Aktif"
	}

	adminStr := "—"
	if adminLevel > 0 {
		adminStr = fmt.Sprintf("Level %d", adminLevel)
	}

	regDate := "—"
	if registerDate > 0 {
		regDate = fmt.Sprintf("<t:%d:D>", registerDate)
	}

	embed := utils.BuildEmbed(b.cfg.LogoURL, b.cfg.ServerName,
		"Info Akun UCP",
		"",
		utils.ColorPrimary,
		[]*utils.EmbedField{
			utils.Field("Username", fmt.Sprintf("`%s`", ucpName), true),
			utils.Field("Role", roleStatus, true),
			utils.Field("Admin", adminStr, true),
			utils.Field("Terdaftar", regDate, true),
			utils.Field("Discord ID", fmt.Sprintf("`%s`", discordID), false),
		},
	)

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	})
}

// ─────────────────────────────────────────
//  /setserver
// ─────────────────────────────────────────

func (b *Bot) handleSetServer(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Cek admin access
	if !b.isAdminInteraction(s, i) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: utils.EphemeralResponse(
				utils.BuildEmbed(b.cfg.LogoURL, b.cfg.ServerName,
					"Akses Ditolak",
					"Kamu tidak memiliki izin untuk menjalankan perintah ini.",
					utils.ColorError,
					nil,
				),
			),
		})
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral},
	})

	options := i.ApplicationCommandData().Options
	optMap := make(map[string]string)
	for _, opt := range options {
		optMap[opt.Name] = opt.StringValue()
	}

	sc := config.LoadServerConfig(b.cfg)
	sc.Host = optMap["host"]
	sc.Port = optMap["port"]
	if name, ok := optMap["name"]; ok && name != "" {
		sc.Name = name
	}

	if err := config.SaveServerConfig(sc); err != nil {
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{
				utils.BuildEmbed(b.cfg.LogoURL, b.cfg.ServerName,
					"Gagal Menyimpan",
					"Terjadi kesalahan saat menyimpan konfigurasi.",
					utils.ColorError,
					nil,
				),
			},
		})
		return
	}

	log.Printf("🔧 Server config diubah: %s:%s (%s)", sc.Host, sc.Port, sc.Name)

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{
			utils.BuildEmbed(b.cfg.LogoURL, b.cfg.ServerName,
				"Konfigurasi Server Diperbarui",
				"Bot akan menggunakan server baru mulai sekarang.",
				utils.ColorSuccess,
				[]*utils.EmbedField{
					utils.Field("Nama", sc.Name, true),
					utils.Field("Alamat", fmt.Sprintf("`%s:%s`", sc.Host, sc.Port), true),
				},
			),
		},
	})
}

// isAdminInteraction mengecek admin access dari InteractionCreate
func (b *Bot) isAdminInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	if i.Member == nil {
		return false
	}
	userID := i.Member.User.ID

	if len(b.cfg.AdminUserIDs) > 0 {
		for _, id := range b.cfg.AdminUserIDs {
			if id == userID {
				return true
			}
		}
		return false
	}

	guild, err := s.Guild(i.GuildID)
	if err != nil {
		return false
	}

	if guild.OwnerID == userID {
		return true
	}

	var totalPerms int64
	for _, memberRoleID := range i.Member.Roles {
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
