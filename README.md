# discord-ucp-bot

A high-performance Discord bot for SA-MP server User Control Panel (UCP) management, written in Go. Handles player registration, role sync, PIN delivery, and password reset — all through persistent Discord panels with interactive buttons and modals.

---

## Features

- **Register UCP** — Create a new in-game identity with strict validation
- **Sync Role** — Re-link a Discord account to its UCP role
- **Check Status** — View current account info (UCP name, role status, Discord ID)
- **Resend PIN** — Generate a new PIN and deliver it via Direct Message
- **Reset Password** — Reset account password with SHA-256 + salt hashing
- **Persistent Panel** — Panel survives bot restarts; auto-refreshes on `!panel`
- **SAMP Status** — Bot presence shows live player count from the SA-MP server

---

## Requirements

- Go 1.21+
- MySQL 5.7+ or MariaDB 10.3+
- A Discord bot token with the following **Privileged Gateway Intents** enabled:
  - Server Members Intent
  - Message Content Intent

---

## Quick Start

### 1. Clone the repository

```bash
git clone https://github.com/BagasZkyn/discord-ucp-bot.git
cd discord-ucp-bot
```

### 2. Configure environment

```bash
cp .env.example .env
```

Edit `.env` with your values:

```env
DISCORD_TOKEN=your_bot_token_here
UCP_ROLE_ID=your_role_id_here

SERVER_NAME=Djava Roleplay
LOGO_URL=https://your-logo-url.png

SAMP_HOST=your_samp_server_ip
SAMP_PORT=7777

DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=samp_db
```

**How to get Role ID:** Enable Developer Mode in Discord Settings → Advanced, then right-click the role → Copy ID.

### 3. Set up the database

```sql
CREATE DATABASE IF NOT EXISTS samp_db;
USE samp_db;

CREATE TABLE `ucp` (
  `username`     VARCHAR(24)  PRIMARY KEY,
  `password`     VARCHAR(129) DEFAULT '',
  `salt`         VARCHAR(16)  DEFAULT '',
  `verifycode`   INT          DEFAULT 0,
  `DiscordID`    VARCHAR(32)  DEFAULT '',
  `admin`        INT          DEFAULT 0,
  `extrac`       INT          DEFAULT 0,
  `ip`           VARCHAR(16)  DEFAULT '',
  `registerdate` INT          DEFAULT 0
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 4. Run

**Option A — Use a pre-built binary from [Releases](../../releases)**

Download the binary for your platform, place it next to your `.env` file, and run it.

**Option B — Build from source**

```bash
go mod download
go build -o botucp ./cmd/bot
./botucp
```

---

## Usage

| Command | Permission | Description |
|---------|-----------|-------------|
| `!panel` | Manage Server | Spawn or refresh the UCP panel in the current channel |

The panel is persistent — it is stored in `panel_config.json` and automatically verified on every bot startup. If the original message is gone, run `!panel` again to create a new one.

---

## Project Structure

```
discord-ucp-bot/
├── cmd/
│   └── bot/
│       └── main.go                  # Entry point
├── internal/
│   ├── bot/
│   │   ├── bot.go                   # Discord session, event routing
│   │   └── status.go                # SAMP UDP query & presence updater
│   ├── config/
│   │   ├── config.go                # Environment variable loader
│   │   └── panel_config.go          # Persistent panel config (JSON)
│   ├── database/
│   │   └── database.go              # MySQL connection pool
│   ├── handlers/
│   │   └── user/
│   │       ├── handler.go           # !panel command + panel builder
│   │       ├── register.go          # Registration modal handler
│   │       ├── sync_role.go         # Role sync handler
│   │       ├── status.go            # Status check handler
│   │       ├── resend_pin.go        # PIN resend handler
│   │       └── reset_password.go    # Password reset handler
│   └── utils/
│       ├── embed.go                 # Embed builder helpers
│       ├── validation.go            # UCP name validation
│       └── crypto.go                # SHA-256 hashing + salt generation
├── .env.example
├── .gitignore
├── go.mod
├── go.sum
└── README.md
```

---

## Configuration Reference

| Variable | Description | Default |
|----------|-------------|---------|
| `DISCORD_TOKEN` | Discord bot token | *(required)* |
| `UCP_ROLE_ID` | Role ID granted on registration | *(required)* |
| `SERVER_NAME` | Server name shown in embeds | `Djava Roleplay` |
| `LOGO_URL` | Logo URL used in embeds | *(built-in URL)* |
| `SAMP_HOST` | SA-MP server IP | `82.25.36.26` |
| `SAMP_PORT` | SA-MP server port | `7043` |
| `DB_HOST` | MySQL host | `localhost` |
| `DB_PORT` | MySQL port | `3306` |
| `DB_USER` | MySQL username | `root` |
| `DB_PASSWORD` | MySQL password | *(empty)* |
| `DB_NAME` | MySQL database name | `samp_db` |

---

## Security

- Passwords are hashed with **SHA-256 + random 16-char hex salt** before storage
- PIN codes are delivered exclusively via **Direct Message** — never shown in public channels
- UCP names are validated with a strict regex (`^[a-zA-Z0-9]{3,20}$`)
- One Discord account can only be linked to one UCP identity
- All database queries use **parameterized statements** to prevent SQL injection

---

## Running as a Service

### Linux (systemd)

Create `/etc/systemd/system/discord-ucp-bot.service`:

```ini
[Unit]
Description=Discord UCP Bot
After=network.target mysql.service

[Service]
Type=simple
User=your_user
WorkingDirectory=/opt/discord-ucp-bot
ExecStart=/opt/discord-ucp-bot/botucp-linux-amd64
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable discord-ucp-bot
sudo systemctl start discord-ucp-bot
```

### Windows (NSSM)

```cmd
nssm install discord-ucp-bot "C:\bots\discord-ucp-bot\botucp-windows-amd64.exe"
nssm set discord-ucp-bot AppDirectory "C:\bots\discord-ucp-bot"
nssm start discord-ucp-bot
```

Download NSSM from https://nssm.cc/download

---

## Building from Source

Cross-compile for all platforms:

```bash
# Windows
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/botucp-windows-amd64.exe ./cmd/bot
GOOS=windows GOARCH=arm64 go build -ldflags="-s -w" -o dist/botucp-windows-arm64.exe ./cmd/bot

# Linux
GOOS=linux GOARCH=amd64  go build -ldflags="-s -w" -o dist/botucp-linux-amd64    ./cmd/bot
GOOS=linux GOARCH=arm64  go build -ldflags="-s -w" -o dist/botucp-linux-arm64    ./cmd/bot
GOOS=linux GOARCH=arm    go build -ldflags="-s -w" -o dist/botucp-linux-arm      ./cmd/bot

# macOS
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o dist/botucp-darwin-amd64  ./cmd/bot
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/botucp-darwin-arm64  ./cmd/bot
```

Or just push a tag — GitHub Actions will build and publish a release automatically.

---

## Releases

Pre-built binaries are available on the [Releases](../../releases) page for:

| Platform | Architecture | File |
|----------|-------------|------|
| Windows | x86-64 | `botucp-windows-amd64.exe` |
| Windows | ARM64 | `botucp-windows-arm64.exe` |
| Linux | x86-64 | `botucp-linux-amd64` |
| Linux | ARM64 | `botucp-linux-arm64` |
| Linux | ARMv7 | `botucp-linux-arm` |
| macOS | x86-64 | `botucp-darwin-amd64` |
| macOS | Apple Silicon | `botucp-darwin-arm64` |

---

## License

MIT — free to use, modify, and distribute.
