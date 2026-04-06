package discord

import (
	"fmt"
	"strings"

	"survival-bot/internal/db"

	"github.com/bwmarrin/discordgo"
)

func (b *Bot) handleDM(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.GuildID != "" {
		return
	}

	if b.cfg.OwnerID == "" {
		return
	}

	if m.Author.ID != b.cfg.OwnerID {
		return
	}

	content := strings.TrimSpace(m.Content)

	if strings.HasPrefix(content, "link ") {
		b.handleLinkCommand(s, m.ChannelID, content[5:])
		return
	}

	b.handleDebug(s, m)
}

func (b *Bot) handleLinkCommand(s *discordgo.Session, channelID, args string) {
	parts := strings.SplitN(args, " ", 2)
	if len(parts) != 2 {
		_, _ = s.ChannelMessageSend(channelID, "Usage: link <player_name> <discord_id>")
		return
	}

	playerName := parts[0]
	discordInput := parts[1]

	link := &db.PlayerLink{
		Game:          b.cfg.Game,
		PlayerName:    playerName,
		DiscordUserID: discordInput,
	}

	if err := b.db.InsertPlayerLink(link); err != nil {
		_, _ = s.ChannelMessageSend(channelID, fmt.Sprintf("Error: %v", err))
		return
	}

	_, _ = s.ChannelMessageSend(channelID, fmt.Sprintf("✅ Linked **%s** to <@%s>", playerName, discordInput))
}

func (b *Bot) handleDebug(s *discordgo.Session, m *discordgo.MessageCreate) {
	messages := b.getAndClearDebug()
	if len(messages) == 0 {
		_, _ = s.ChannelMessageSend(m.ChannelID, "No debug messages.")
		return
	}

	response := "📋 Debug Messages:\n" + strings.Join(messages, "\n")
	_, _ = s.ChannelMessageSend(m.ChannelID, response)
}

func (b *Bot) getAndClearDebug() []string {
	b.mu.Lock()
	defer b.mu.Unlock()
	messages := make([]string, len(b.debugMessages))
	copy(messages, b.debugMessages)
	b.debugMessages = nil
	return messages
}
