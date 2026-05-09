package user

import (
	"fmt"
	"log"

	"botucp/internal/utils"

	"github.com/bwmarrin/discordgo"
)

// HandleSyncRoleButton menangani tombol Sync Role
func (h *Handler) HandleSyncRoleButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral},
	})

	discordID := i.Member.User.ID

	var ucpName string
	err := h.db.QueryRow("SELECT `username` FROM `ucp` WHERE `DiscordID` = ? LIMIT 1", discordID).Scan(&ucpName)
	if err != nil {
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{
				h.embed("Akun Tidak Ditemukan",
					"Discord kamu belum terdaftar. Silakan daftar UCP terlebih dahulu.",
					utils.ColorError),
			},
		})
		return
	}

	if h.cfg.UCPRoleID == "" {
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{
				h.embed("Konfigurasi Tidak Lengkap",
					"Role UCP belum dikonfigurasi. Hubungi administrator.",
					utils.ColorWarning),
			},
		})
		return
	}

	for _, roleID := range i.Member.Roles {
		if roleID == h.cfg.UCPRoleID {
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Embeds: &[]*discordgo.MessageEmbed{
					h.embed("Role Sudah Aktif",
						fmt.Sprintf("Kamu sudah memiliki role <@&%s>.", h.cfg.UCPRoleID),
						utils.ColorInfo),
				},
			})
			return
		}
	}

	if err := s.GuildMemberRoleAdd(i.GuildID, discordID, h.cfg.UCPRoleID); err != nil {
		log.Printf("❌ Gagal berikan role: %v", err)
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{
				h.embed("Gagal Sync Role",
					"Tidak bisa memberikan role. Hubungi administrator.",
					utils.ColorError),
			},
		})
		return
	}

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{
			utils.BuildEmbed(h.cfg.LogoURL, h.cfg.ServerName,
				"Sync Role Berhasil",
				fmt.Sprintf("Role <@&%s> berhasil diberikan kembali.", h.cfg.UCPRoleID),
				utils.ColorSuccess,
				[]*utils.EmbedField{
					utils.Field("Username", fmt.Sprintf("`%s`", ucpName), true),
					utils.Field("Role", fmt.Sprintf("<@&%s>", h.cfg.UCPRoleID), true),
				}),
		},
	})
}
