package bot

import (
	"database/sql"
	"fmt"
	"log"

	"botucp/internal/config"
	"botucp/internal/handlers/user"

	"github.com/bwmarrin/discordgo"
)

// Bot adalah struct utama yang menyimpan semua dependency
type Bot struct {
	session     *discordgo.Session
	cfg         *config.Config
	db          *sql.DB
	userHandler *user.Handler
}

// New membuat instance Bot baru
func New(cfg *config.Config, db *sql.DB) (*Bot, error) {
	if cfg.DiscordToken == "" {
		return nil, fmt.Errorf("DISCORD_TOKEN tidak diset")
	}

	session, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat session Discord: %w", err)
	}

	session.Identify.Intents = discordgo.IntentsGuilds |
		discordgo.IntentsGuildMembers |
		discordgo.IntentsGuildMessages |
		discordgo.IntentsMessageContent |
		discordgo.IntentsDirectMessages

	b := &Bot{
		session:     session,
		cfg:         cfg,
		db:          db,
		userHandler: user.NewHandler(cfg, db),
	}

	session.AddHandler(b.onReady)
	session.AddHandler(b.onMessageCreate)
	session.AddHandler(b.onInteractionCreate)

	return b, nil
}

// Start membuka koneksi ke Discord
func (b *Bot) Start() error {
	return b.session.Open()
}

// Stop menutup koneksi ke Discord
func (b *Bot) Stop() {
	b.session.Close()
}

// Tag mengembalikan tag bot
func (b *Bot) Tag() string {
	if b.session.State.User != nil {
		return b.session.State.User.String()
	}
	return "Unknown"
}

// ─────────────────────────────────────────
//  EVENT: Ready
// ─────────────────────────────────────────

func (b *Bot) onReady(s *discordgo.Session, r *discordgo.Ready) {
	log.Printf("✅ Bot online: %s", r.User.String())

	b.startStatusUpdater()
	b.registerSlashCommands()

	// Verifikasi UCP panel persisten
	cfg := config.LoadPanelConfig()
	if cfg.UCPPanelChannelID != "" && cfg.UCPPanelMessageID != "" {
		_, err1 := s.Channel(cfg.UCPPanelChannelID)
		_, err2 := s.ChannelMessage(cfg.UCPPanelChannelID, cfg.UCPPanelMessageID)
		if err1 != nil || err2 != nil {
			log.Println("⚠️  UCP Panel tidak ditemukan — kirim ulang dengan !panel")
			config.SavePanelConfig(config.PanelConfig{})
		} else {
			log.Printf("🟢  UCP Panel aktif: channel=%s msg=%s", cfg.UCPPanelChannelID, cfg.UCPPanelMessageID)
		}
	}
}

// ─────────────────────────────────────────
//  EVENT: Message Create
// ─────────────────────────────────────────

func (b *Bot) onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}
	if m.Content == "!panel" {
		b.userHandler.HandlePanelCommand(s, m)
	}
}

// ─────────────────────────────────────────
//  EVENT: Interaction Create
// ─────────────────────────────────────────

func (b *Bot) onInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		b.handleSlashCommand(s, i)
	case discordgo.InteractionMessageComponent:
		b.handleComponent(s, i)
	case discordgo.InteractionModalSubmit:
		b.handleModalSubmit(s, i)
	}
}

func (b *Bot) handleComponent(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.MessageComponentData().CustomID {
	case "btn_register":
		b.userHandler.HandleRegisterButton(s, i)
	case "btn_refund_role":
		b.userHandler.HandleSyncRoleButton(s, i)
	case "btn_check_status":
		b.userHandler.HandleCheckStatusButton(s, i)
	case "btn_resend_pin":
		b.userHandler.HandleResendPINButton(s, i)
	case "btn_reset_password":
		b.userHandler.HandleResetPasswordButton(s, i)
	}
}

func (b *Bot) handleModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.ModalSubmitData().CustomID {
	case "modal_register":
		b.userHandler.HandleRegisterModal(s, i)
	case "modal_reset_password":
		b.userHandler.HandleResetPasswordModal(s, i)
	}
}
