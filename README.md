# Survival Bot

Discord bot application for survival games ran on dedicated private server. \
Is able to monitor game server log files and sends notifications about game events to Discord. \
Parse and persists game events and enables Discord commands to explore stats.

## Features

- Player death notifications

## Getting Started

### Configuration

Create a `.env` file in the project root:

```bash
# Copy the example config
cp .env.example .env
```

Edit `.env` with your values:

| Variable             | Required | Description                                              |
|----------------------|----------|----------------------------------------------------------|
| `GAME`               | Yes      | Game being played (`soulmask`, `valheim`)                |
| `LOG_FILE_PATH`      | Yes      | Path to game server log file                             |
| `DISCORD_BOT_TOKEN`  | Yes      | Discord Bot Token                                        |
| `DISCORD_GUILD_ID`   | Yes      | Discord Server ID where bot is running                   |
| `DISCORD_CHANNEL_ID` | Yes      | Discord Channel ID where bot replies to commands         |
| `DISCORD_OWNER_ID`   | No       | Discord User ID that has access to admin/debug features  |
| `DB_PATH`            | No       | Where sqlite will be saved. Defaults to current path     |
| `GEMINI_API_KEY`     | No       | Google Gemini API key for AI-generated messages          |

### Create executable

```bash
# Build for current platform
make build

# Build Linux only
make build-linux

# Build Windows only
make build-windows
```

Binary outputs:
- `bin/survival-bot-linux` (Linux)
- `bin/survival-bot.exe` (Windows)

---
