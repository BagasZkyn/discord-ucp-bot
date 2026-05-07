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
				h.embed("「 ❌ 」Akun Tidak Ditemukan",
					"> Discord ID Anda tidak terdaftar di sistem.\n> Silakan lakukan **REGISTRASI UCP** terlebih dahulu.",
					utils.ColorError),
			},
		})
		return
	}

	// Generate PIN baru
	newPIN := utils.GeneratePIN()

	_, err = h.db.Exec("UPDATE `ucp` SET `verifycode` = ? WHERE `DiscordID` = ?", newPIN, discordID)
	if err != nil {
		log.Printf("❌ Gagal update PIN: %v", err)
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{
				h.embed("「 ❌ 」Error Database",
					"Terjadi kegagalan komunikasi dengan server database.",
					utils.ColorError),
			},
		})
		return
	}

	// Kirim PIN via DM
	ch, err := s.UserChannelCreate(discordID)
	if err != nil {
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{
				h.embed("「 ⚠️ 」DM Tertutup",
					"> PIN baru telah digenerate namun gagal dikirim via DM.\n> Buka DM Anda lalu coba lagi.",
					utils.ColorWarning),
			},
		})
		return
	}

	dmEmbed := &discordgo.MessageEmbed{
		Color: utils.ColorPrimary,
		Author: &discordgo.MessageEmbedAuthor{
			Name: "[ DJAVA ROLEPLAY - ENCRYPTED TRANSMISSION ]",
		},
		Title:       "「 🔐 」PIN BARU DITERBITKAN",
		Description: fmt.Sprintf("PIN baru telah digenerate untuk identitas **%s**.\nGunakan PIN ini untuk autentikasi in-game.", ucpName),
		Fields: []*discordgo.MessageEmbedField{
			{Name: "IDENTITAS", Value: fmt.Sprintf("```bash\n\"%s\"\n```", ucpName), Inline: false},
			{Name: "PIN BARU", Value: fmt.Sprintf("```diff\n+ %d +\n```", newPIN), Inline: false},
			{Name: "⚠️ PERINGATAN", Value: "> **JAGA KERAHASIAAN PIN INI!**\n> Staf tidak akan pernah meminta kode ini.", Inline: false},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    "Djava Secure Node",
			IconURL: h.cfg.LogoURL,
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	_, err = s.ChannelMessageSendEmbed(ch.ID, dmEmbed)
	if err != nil {
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{
				h.embed("「 ⚠️ 」DM Tertutup",
					"> PIN baru telah digenerate namun gagal dikirim via DM.\n> Buka DM Anda lalu coba lagi.",
					utils.ColorWarning),
			},
		})
		return
	}

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{
			h.embed("「 ✅ 」PIN Berhasil Dikirim",
				fmt.Sprintf("> PIN baru untuk **`%s`** telah dikirim ke DM Anda. 📬", ucpName),
				utils.ColorSuccess),
		},
	})
}
