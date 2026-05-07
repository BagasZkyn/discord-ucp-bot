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
			Title:    "📋 Terminal Registrasi",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "input_ucp",
							Label:       "Identitas UCP (Username)",
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
	// Defer reply ephemeral
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})

	data := i.ModalSubmitData()
	ucpName := getModalField(data, "input_ucp")
	discordID := i.Member.User.ID

	// Validasi format UCP
	if !utils.IsValidUCP(ucpName) {
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{
				h.embed("「 ⛔ 」Validasi Ditolak",
					"> Input identitas melanggar protokol keamanan!\n\n**Parameter yang Diizinkan:**\n> • Hanya karakter alfabet & numerik.\n> • Tanpa spasi atau simbol khusus (_).\n> • Panjang 3-20 karakter.",
					utils.ColorError),
			},
		})
		return
	}

	// Cek username sudah ada
	var existing string
	err := h.db.QueryRow("SELECT `username` FROM `ucp` WHERE `username` = ? LIMIT 1", ucpName).Scan(&existing)
	if err == nil {
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{
				h.embed("「 ❌ 」Identitas Duplikat",
					fmt.Sprintf("> Identitas **`%s`** telah diregistrasi oleh entitas lain.\n> Gunakan kombinasi nama alternatif.", ucpName),
					utils.ColorError),
			},
		})
		return
	}

	// Cek Discord ID sudah terdaftar
	var existingUCP string
	err = h.db.QueryRow("SELECT `username` FROM `ucp` WHERE `DiscordID` = ? LIMIT 1", discordID).Scan(&existingUCP)
	if err == nil {
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{
				h.embed("「 ❌ 」Akses Terbatas",
					fmt.Sprintf("> Node Discord ini telah dikaitkan dengan identitas: **`%s`**.\n> Tidak dapat mendaftar lagi.", existingUCP),
					utils.ColorError),
			},
		})
		return
	}

	// Generate verify code dan insert ke database
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
				h.embed("「 ❌ 」Critical Error",
					fmt.Sprintf("Fatal system exception:\n```%v```", err),
					utils.ColorError),
			},
		})
		return
	}

	// Berikan role UCP
	roleGiven := false
	if h.cfg.UCPRoleID != "" && i.Member != nil {
		err := s.GuildMemberRoleAdd(i.GuildID, discordID, h.cfg.UCPRoleID)
		roleGiven = err == nil
	}

	// Kirim PIN via DM
	dmSent := h.sendPINDM(s, i.Member.User, ucpName, verifyCode)

	// Buat pesan status role
	roleStatus := "⚠️ Gagal memberikan otorisasi."
	if roleGiven {
		roleStatus = fmt.Sprintf("✅ Otorisasi <@&%s> diberikan.", h.cfg.UCPRoleID)
	}

	dmStatus := ""
	if !dmSent {
		dmStatus = "\n> ⚠️ Gagal kirim DM — buka DM Anda lalu gunakan **Resend PIN**."
	}

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{
			h.embed("「 ✅ 」Registrasi Diterima",
				fmt.Sprintf("> Identitas **`%s`** sukses masuk ke dalam sistem. 🟢\n\n> 📬 Cek **Direct Message (DM)** Anda untuk melihat Token Autentikasi.\n> %s%s",
					ucpName, roleStatus, dmStatus),
				utils.ColorSuccess),
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
		Color: utils.ColorPrimary,
		Author: &discordgo.MessageEmbedAuthor{
			Name: "[ DJAVA ROLEPLAY - ENCRYPTED TRANSMISSION ]",
		},
		Title:       "「 🔐 」SECURITY CLEARANCE",
		Description: fmt.Sprintf("Sistem telah menerbitkan token akses untuk identitas **%s**.\nSilakan gunakan token di bawah ini untuk autentikasi in-game.", ucpName),
		Fields: []*discordgo.MessageEmbedField{
			{Name: "IDENTITAS", Value: fmt.Sprintf("```bash\n\"%s\"\n```", ucpName), Inline: false},
			{Name: "TOKEN AUTENTIKASI", Value: fmt.Sprintf("```diff\n+ %d +\n```", pin), Inline: false},
			{Name: "⚠️ PERINGATAN", Value: "> **JAGA KERAHASIAAN TOKEN INI!**\n> Staf tidak akan pernah meminta kode ini.", Inline: false},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    "Djava Secure Node",
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
