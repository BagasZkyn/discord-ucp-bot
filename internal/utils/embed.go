package utils

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

// Warna tema — palette bersih & modern
const (
	ColorPrimary = 0x5865F2 // Blurple Discord
	ColorSuccess = 0x57F287 // Hijau
	ColorError   = 0xED4245 // Merah
	ColorWarning = 0xFEE75C // Kuning
	ColorInfo    = 0x2b2d31 // Dark neutral
)

// EmbedField adalah shorthand untuk discordgo.MessageEmbedField
type EmbedField = discordgo.MessageEmbedField

// BuildEmbed membuat embed standar UCP
func BuildEmbed(logoURL, serverName, title, description string, color int, fields []*EmbedField) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Color:       color,
		Title:       title,
		Description: description,
		Footer: &discordgo.MessageEmbedFooter{
			Text:    serverName,
			IconURL: logoURL,
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}
	if len(fields) > 0 {
		embed.Fields = fields
	}
	return embed
}

// Field adalah helper untuk membuat EmbedField
func Field(name, value string, inline bool) *EmbedField {
	return &EmbedField{Name: name, Value: value, Inline: inline}
}

// EphemeralResponse membuat response ephemeral dengan embed
func EphemeralResponse(embeds ...*discordgo.MessageEmbed) *discordgo.InteractionResponseData {
	return &discordgo.InteractionResponseData{
		Embeds: embeds,
		Flags:  discordgo.MessageFlagsEphemeral,
	}
}
