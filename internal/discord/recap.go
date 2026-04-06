package discord

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"survival-bot/internal/db"

	"github.com/bwmarrin/discordgo"
)

func (b *Bot) SendDailyRecap(ctx context.Context) error {
	game := b.cfg.Game

	deaths, err := b.db.GetDeaths(game, true)
	if err != nil {
		return fmt.Errorf("failed to get deaths: %w", err)
	}

	if len(deaths) == 0 {
		embed := &discordgo.MessageEmbed{
			Title:       "📊 Today's Death Ranking",
			Description: "No deaths today! 🎉",
			Color:       0x00008B,
		}
		return b.sendEmbed(embed)
	}

	embed := &discordgo.MessageEmbed{
		Title:       "📊 Today's Death Ranking",
		Description: b.buildLeaderboard(deaths),
		Color:       0x00008B,
	}
	if err = b.sendEmbed(embed); err != nil {
		b.AddDebug(fmt.Sprintf("daily recap leaderboard, sendEmbed: %v", err))
		return fmt.Errorf("failed to send embed of leaderboard: %w", err)
	}

	if b.llm != nil {
		summary, err := b.generateRecapSummary(ctx, deaths)
		if err != nil {
			b.AddDebug(fmt.Sprintf("daily recap generateRecapSummary: %v", err))
			return fmt.Errorf("failed generate recap summary: %w", err)
		}

		embed = &discordgo.MessageEmbed{
			Title:       "Daily Recap",
			Description: summary,
			Color:       0x00008B,
		}
		if err = b.sendEmbed(embed); err != nil {
			b.AddDebug(fmt.Sprintf("daily recap summary, sendEmbed: %v", err))
			return fmt.Errorf("failed to send embed of recap: %w", err)
		}
	}

	return nil
}

func (b *Bot) buildLeaderboard(deaths []db.Death) string {
	deathCounts := make(map[string]int)
	for _, d := range deaths {
		deathCounts[d.PlayerName]++
	}

	type playerDeathCount struct {
		name  string
		count int
	}

	var sorted []playerDeathCount
	for name, count := range deathCounts {
		sorted = append(sorted, playerDeathCount{name: name, count: count})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].count > sorted[j].count
	})

	description := ""
	medals := []string{"🥇", "🥈", "🥉"}
	for i, p := range sorted {
		if i < 3 {
			description += fmt.Sprintf("%s **%s**: %d deaths\n", medals[i], p.name, p.count)
		} else {
			description += fmt.Sprintf("%d. **%s**: %d deaths\n", i+1, p.name, p.count)
		}
	}

	return description
}

func (b *Bot) generateRecapSummary(ctx context.Context, deaths []db.Death) (string, error) {
	prompt := fmt.Sprintf(`Generate a savage recap about todays game events.

	Today's in-game deaths:
	{{death_list}}

	Requirements:
	- Game : %s
	- 4-5 sentences max
	- Be humorous and summarize the day's drama. You dont need to mention all players
	- You can roast the 2-3 players with the most deaths
	- Be savage and unforgiving with the roasting
	- Don't hold back, make it hurt (in a funny way)
	- End with a friendly "see you tomorrow" message
	- Message will be sent to a discord server (use line breaks to improve visibility)
	- Format: just the recap message, no quotes or extra text, player names should be wrapped in ** **

	Just output the recap message, nothing else.`, b.Game())

	var deathList string
	for _, d := range deaths {
		deathList += fmt.Sprintf("- %s\n", d.String())
	}

	prompt = strings.Replace(prompt, "{{death_list}}", deathList, 1)
	return b.llm.Ask(ctx, prompt)
}

func (b *Bot) sendEmbed(embed *discordgo.MessageEmbed) error {
	_, err := b.session.ChannelMessageSendEmbed(b.cfg.DiscordChannelID, embed)
	return err
}
