package user

import (
	"fmt"

	"botucp/internal/utils"

	"github.com/bwmarrin/discordgo"
)

// HandleCheckStatusButton menangani tombol Cek Status
func (h *Handler) HandleCheckStatusButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral},
	})

	discordID := i.Member.User.ID

	var ucpName string
	var adminLevel int
	err := h.db.QueryRow(
		"SELECT `username`, `admin` FROM `ucp` WHERE `DiscordID` = ? LIMIT 1",
		discordID,
	).Scan(&ucpName, &adminLevel)

	if err != nil {
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{
				h.embed("「 🔍 」Status: Unknown",
					"> Identitas tidak terdaftar di sistem.\n> Silakan klik **REGISTRASI UCP** untuk memulai.",
					utils.ColorWarning),
			},
		})
		return
	}

	// Cek apakah punya role
	hasRole := false
	for _, roleID := range i.Member.Roles {
		if roleID == h.cfg.UCPRoleID {
			hasRole = true
			break
		}
	}

	authStatus := "❌ **Dicabut**"
	if hasRole {
		authStatus = "✅ **Terverifikasi**"
	}

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{
			utils.BuildEmbed(h.cfg.LogoURL, h.cfg.ServerName,
				"「 🔍 」Diagnostic Data",
				"> Berikut adalah rekam data digital Anda saat ini.",
				utils.ColorPrimary,
				[]*utils.EmbedField{
					utils.Field("🏷️ UCP", fmt.Sprintf("`%s`", ucpName), true),
					utils.Field("🎭 Otorisasi", authStatus, true),
					utils.Field("🆔 Node ID", fmt.Sprintf("`%s`", discordID), false),
				}),
		},
	})
}
