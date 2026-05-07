package user

import (
	"fmt"
	"log"

	"botucp/internal/utils"

	"github.com/bwmarrin/discordgo"
)

// HandleResetPasswordButton menampilkan modal reset password (setelah cek akun)
func (h *Handler) HandleResetPasswordButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	discordID := i.Member.User.ID

	// Cek akun dulu sebelum tampilkan modal
	var ucpName string
	err := h.db.QueryRow(
		"SELECT `username` FROM `ucp` WHERE `DiscordID` = ? LIMIT 1",
		discordID,
	).Scan(&ucpName)

	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: utils.EphemeralResponse(
				h.embed("「 ❌ 」Akun Tidak Ditemukan",
					"> Discord ID Anda tidak terdaftar di sistem.\n> Silakan lakukan **REGISTRASI UCP** terlebih dahulu.",
					utils.ColorError),
			),
		})
		return
	}

	// Tampilkan modal
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: "modal_reset_password",
			Title:    "🔑 Reset Password Akun",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "input_new_password",
							Label:       "Password Baru",
							Style:       discordgo.TextInputShort,
							Placeholder: "Masukkan password baru (min 6 karakter)",
							MinLength:   6,
							MaxLength:   32,
							Required:    true,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "input_confirm_password",
							Label:       "Konfirmasi Password Baru",
							Style:       discordgo.TextInputShort,
							Placeholder: "Ulangi password baru",
							MinLength:   6,
							MaxLength:   32,
							Required:    true,
						},
					},
				},
			},
		},
	})
	if err != nil {
		log.Printf("❌ Gagal tampilkan modal reset password: %v", err)
	}
}

// HandleResetPasswordModal memproses submit modal reset password
func (h *Handler) HandleResetPasswordModal(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral},
	})

	data := i.ModalSubmitData()
	discordID := i.Member.User.ID
	newPassword := getModalField(data, "input_new_password")
	confirmPassword := getModalField(data, "input_confirm_password")

	if newPassword != confirmPassword {
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{
				h.embed("「 ❌ 」Password Tidak Cocok",
					"> Password baru dan konfirmasi tidak sama.\n> Silakan coba lagi.",
					utils.ColorError),
			},
		})
		return
	}

	var ucpName string
	err := h.db.QueryRow(
		"SELECT `username` FROM `ucp` WHERE `DiscordID` = ? LIMIT 1",
		discordID,
	).Scan(&ucpName)

	if err != nil {
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{
				h.embed("「 ❌ 」Akun Tidak Ditemukan",
					"> Discord ID Anda tidak terdaftar di sistem.",
					utils.ColorError),
			},
		})
		return
	}

	// Hash password dengan SHA256 + salt baru
	salt := utils.GenerateSalt()
	hashedPassword := utils.HashPassword(newPassword, salt)

	// Update tabel ucp
	_, err = h.db.Exec(
		"UPDATE `ucp` SET `password` = ?, `salt` = ? WHERE `DiscordID` = ?",
		hashedPassword, salt, discordID,
	)
	if err != nil {
		log.Printf("❌ Gagal update password UCP: %v", err)
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{
				h.embed("「 ❌ 」Error Database",
					"Terjadi kegagalan komunikasi dengan server database.",
					utils.ColorError),
			},
		})
		return
	}

	// Sync ke tabel players (best effort, tidak fatal jika gagal)
	h.db.Exec(
		"UPDATE `players` SET `password` = ?, `salt` = ? WHERE `ucp` = ?",
		hashedPassword, salt, ucpName,
	)

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{
			utils.BuildEmbed(h.cfg.LogoURL, h.cfg.ServerName,
				"「 ✅ 」Password Berhasil Direset",
				fmt.Sprintf("> Password akun **`%s`** telah berhasil diperbarui.\n> Gunakan password baru Anda saat login in-game.", ucpName),
				utils.ColorSuccess,
				[]*utils.EmbedField{
					utils.Field("👤 UCP", fmt.Sprintf("`%s`", ucpName), true),
					utils.Field("🔒 Status", "`PASSWORD UPDATED`", true),
				}),
		},
	})
}
