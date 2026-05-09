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
				h.embed("Akun Tidak Ditemukan",
					"Discord kamu belum terdaftar. Klik **Daftar UCP** untuk membuat akun.",
					utils.ColorWarning),
			},
		})
		return
	}

	hasRole := false
	for _, roleID := range i.Member.Roles {
		if roleID == h.cfg.UCPRoleID {
			hasRole = true
			break
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

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{
			utils.BuildEmbed(h.cfg.LogoURL, h.cfg.ServerName,
				"Status Akun",
				"Informasi akun UCP kamu.",
				utils.ColorPrimary,
				[]*utils.EmbedField{
					utils.Field("Username", fmt.Sprintf("`%s`", ucpName), true),
					utils.Field("Role", roleStatus, true),
					utils.Field("Admin", adminStr, true),
					utils.Field("Discord ID", fmt.Sprintf("`%s`", discordID), false),
				}),
		},
	})
}
