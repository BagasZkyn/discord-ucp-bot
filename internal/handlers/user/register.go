package user

import (
	"fmt"
	"log"
	"time"

	"botucp/internal/utils"

	"github.com/bwmarrin/discordgo"
)

// HandleRegisterButton menampilkan modal registrasi
func (h *Handler) HandleRegisterButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: "modal_register",
			Title:    "Daftar UCP",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "input_ucp",
							Label:       "Username",
							Style:       discordgo.TextInputShort,
							Placeholder: "Contoh: BagasDjava, ZeroCool99",
							MinLength:   3,
							MaxLength:   20,
							Required:    true,
						},
					},
				},
			},
		},
	})
	if err != nil {
		log.Printf("❌ Gagal tampilkan modal register: %v", err)
	}
}

// HandleRegisterModal memproses submit modal registrasi
func (h *Handler) HandleRegisterModal(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral},
	})

	data := i.ModalSubmitData()
	ucpName := getModalField(data, "input_ucp")
	discordID := i.Member.User.ID

	if !utils.IsValidUCP(ucpName) {
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{
				h.embed("Username Tidak Valid",
					"Username hanya boleh huruf dan angka, 3–20 karakter, tanpa spasi atau simbol.",
					utils.ColorError),
			},
		})
		return
	}

	var existing string
	err := h.db.QueryRow("SELECT `username` FROM `ucp` WHERE `username` = ? LIMIT 1", ucpName).Scan(&existing)
	if err == nil {
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{
				h.embed("Username Sudah Dipakai",
					fmt.Sprintf("Username **%s** sudah terdaftar. Coba nama lain.", ucpName),
					utils.ColorError),
			},
		})
		return
	}

	var existingUCP string
	err = h.db.QueryRow("SELECT `username` FROM `ucp` WHERE `DiscordID` = ? LIMIT 1", discordID).Scan(&existingUCP)
	if err == nil {
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{
				h.embed("Akun Sudah Ada",
					fmt.Sprintf("Discord kamu sudah terhubung ke akun **%s**.", existingUCP),
					utils.ColorError),
			},
		})
		return
	}

	verifyCode := utils.GeneratePIN()
	registerDate := time.Now().Unix()

	_, err = h.db.Exec(
		"INSERT INTO `ucp` (`username`, `verifycode`, `DiscordID`, `password`, `salt`, `ip`, `admin`, `registerdate`) VALUES (?, ?, ?, '', '', '', 0, ?)",
		ucpName, verifyCode, discordID, registerDate,
	)
	if err != nil {
		log.Printf("❌ Gagal insert UCP: %v", err)
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{
				h.embed("Gagal Mendaftar",
					"Terjadi kesalahan pada database. Coba lagi nanti.",
					utils.ColorError),
			},
		})
		return
	}

	roleGiven := false
	if h.cfg.UCPRoleID != "" && i.Member != nil {
		err := s.GuildMemberRoleAdd(i.GuildID, discordID, h.cfg.UCPRoleID)
		roleGiven = err == nil
	}

	dmSent := h.sendPINDM(s, i.Member.User, ucpName, verifyCode)

	roleStatus := "Role gagal diberikan, hubungi admin."
	if roleGiven {
		roleStatus = fmt.Sprintf("Role <@&%s> berhasil diberikan.", h.cfg.UCPRoleID)
	}

	desc := fmt.Sprintf("Akun **%s** berhasil dibuat.\n\n%s", ucpName, roleStatus)
	if !dmSent {
		desc += "\n\nGagal kirim PIN via DM — buka DM kamu lalu gunakan **Resend PIN**."
	} else {
		desc += "\n\nPIN dikirim ke DM kamu."
	}

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{
			h.embed("Registrasi Berhasil", desc, utils.ColorSuccess),
		},
	})
}

// sendPINDM mengirim PIN via DM ke user, return true jika berhasil
func (h *Handler) sendPINDM(s *discordgo.Session, user *discordgo.User, ucpName string, pin int) bool {
	ch, err := s.UserChannelCreate(user.ID)
	if err != nil {
		return false
	}

	dmEmbed := &discordgo.MessageEmbed{
		Color:       utils.ColorPrimary,
		Title:       "PIN Akun UCP",
		Description: fmt.Sprintf("Berikut PIN untuk akun **%s**.\nGunakan saat login in-game.", ucpName),
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Username", Value: fmt.Sprintf("`%s`", ucpName), Inline: true},
			{Name: "PIN", Value: fmt.Sprintf("`%d`", pin), Inline: true},
			{Name: "Peringatan", Value: "Jangan bagikan PIN ini ke siapapun.", Inline: false},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    h.cfg.ServerName,
			IconURL: h.cfg.LogoURL,
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	_, err = s.ChannelMessageSendEmbed(ch.ID, dmEmbed)
	return err == nil
}

// getModalField mengambil nilai field dari modal submit data
func getModalField(data discordgo.ModalSubmitInteractionData, fieldID string) string {
	for _, row := range data.Components {
		if ar, ok := row.(*discordgo.ActionsRow); ok {
			for _, comp := range ar.Components {
				if ti, ok := comp.(*discordgo.TextInput); ok && ti.CustomID == fieldID {
					return ti.Value
				}
			}
		}
	}
	return ""
}
