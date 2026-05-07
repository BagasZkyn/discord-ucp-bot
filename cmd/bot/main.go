package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"botucp/internal/bot"
	"botucp/internal/config"
	"botucp/internal/database"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  .env file tidak ditemukan, menggunakan environment variables sistem")
	}

	// Load config
	cfg := config.Load()

	// Connect database
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("❌ Gagal koneksi database: %v", err)
	}
	defer db.Close()
	log.Println("✅ Database terhubung")

	// Initialize & start bot
	b, err := bot.New(cfg, db)
	if err != nil {
		log.Fatalf("❌ Gagal inisialisasi bot: %v", err)
	}

	if err := b.Start(); err != nil {
		log.Fatalf("❌ Gagal menjalankan bot: %v", err)
	}
	defer b.Stop()

	log.Printf("✅ Bot %s online dan siap!", b.Tag())

	// Graceful shutdown
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	log.Println("🛑 Bot dimatikan...")
}
