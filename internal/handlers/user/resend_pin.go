package user

import (
	"fmt"
	"log"
	"time"

	"botucp/internal/utils"

	"github.com/bwmarrin/discordgo"
)

// HandleResendPINButton menangani tombol Resend PIN
func (h *Handler) HandleResendPINButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral},
	})

	discordID := i.Member.User.ID

	var ucpName string
	err := h.db.QueryRow(
		"SELECT `username` FROM `ucp` WHERE `DiscordID` = ? LIMIT 1",
		discordID,
	).Scan(&ucpName)

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

	newPIN := utils.GeneratePIN()

	_, err = h.db.Exec("UPDATE `ucp` SET `verifycode` = ? WHERE `DiscordID` = ?", newPIN, discordID)
	if err != nil {
		log.Printf("❌ Gagal update PIN: %v", err)
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{
				h.embed("Gagal Update PIN",
					"Terjadi kesalahan pada database. Coba lagi nanti.",
					utils.ColorError),
			},
		})
		return
	}

	ch, err := s.UserChannelCreate(discordID)
	if err != nil {
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{
				h.embed("DM Tidak Bisa Dibuka",
					"PIN baru sudah dibuat tapi gagal dikirim. Buka DM kamu lalu coba lagi.",
					utils.ColorWarning),
			},
		})
		return
	}

	dmEmbed := &discordgo.MessageEmbed{
		Color:       utils.ColorPrimary,
		Title:       "PIN Baru",
		Description: fmt.Sprintf("PIN baru untuk akun **%s** sudah diterbitkan.", ucpName),
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Username", Value: fmt.Sprintf("`%s`", ucpName), Inline: true},
			{Name: "PIN Baru", Value: fmt.Sprintf("`%d`", newPIN), Inline: true},
			{Name: "Peringatan", Value: "Jangan bagikan PIN ini ke siapapun.", Inline: false},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    h.cfg.ServerName,
			IconURL: h.cfg.LogoURL,
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	_, err = s.ChannelMessageSendEmbed(ch.ID, dmEmbed)
	if err != nil {
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{
				h.embed("DM Tidak Bisa Dibuka",
					"PIN baru sudah dibuat tapi gagal dikirim. Buka DM kamu lalu coba lagi.",
					utils.ColorWarning),
			},
		})
		return
	}

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{
			h.embed("PIN Terkirim",
				fmt.Sprintf("PIN baru untuk **%s** sudah dikirim ke DM kamu.", ucpName),
				utils.ColorSuccess),
		},
	})
}
