package discord

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type byDeathCount []playerDeathCount

func (a byDeathCount) Len() int           { return len(a) }
func (a byDeathCount) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byDeathCount) Less(i, j int) bool { return a[i].count > a[j].count }

type playerDeathCount struct {
	playerName string
	count      int
	discordID  string
}

func (b *Bot) HandleLeaderboard(s *discordgo.Session, i *discordgo.InteractionCreate) *discordgo.InteractionResponse {
	opts := i.ApplicationCommandData().Options
	today := false
	title := "💀🏆 ALL-TIME LEADERBOARD"
	for _, opt := range opts {
		if opt.Name == "timeframe" && opt.StringValue() == "today" {
			today = true
			title = "📅 TODAY'S LEADERBOARD"
		}
	}

	game := b.cfg.Game

	deaths, err := b.db.GetDeaths(game, today)
	if err != nil {
		b.AddDebug(fmt.Sprintf("leaderboard GetDeaths: %v", err))
		return &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Something went wrong!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		}
	}

	if len(deaths) == 0 {
		return &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("%s\n━━━━━━━━━━━━━━━━━━━━━━\nNo deaths recorded yet! 🎉", title),
			},
		}
	}

	playerLinks, err := b.db.GetPlayerLinksByGame(game)
	if err != nil {
		b.AddDebug(fmt.Sprintf("get player links: %v", err))
	}

	linkMap := make(map[string]string)
	for _, link := range playerLinks {
		linkMap[link.PlayerName] = link.DiscordUserID
	}

	deathCounts := make(map[string]int)
	for _, d := range deaths {
		deathCounts[d.PlayerName]++
	}

	var sorted []playerDeathCount
	for name, count := range deathCounts {
		sorted = append(sorted, playerDeathCount{
			playerName: name,
			count:      count,
			discordID:  linkMap[name],
		})
	}
	sort.Sort(byDeathCount(sorted))

	description := ""
	medals := []string{"🥇", "🥈", "🥉"}
	for i, p := range sorted {
		if i < 3 {
			description += fmt.Sprintf("%s **%s**: %d deaths\n", medals[i], p.playerName, p.count)
		} else {
			description += fmt.Sprintf("%d. **%s**: %d deaths\n", i+1, p.playerName, p.count)
		}
	}

	description += "\n"

	fallStats, err := b.db.CountDeathsByFallDamage(game, today)
	if err != nil {
		b.AddDebug(fmt.Sprintf("count deaths by fall damage: %v", err))
	}

	if fallStats != nil && fallStats.Total > 0 {
		var topPlayer string
		var topCount int
		for player, count := range fallStats.PlayerNameCounter {
			if count > topCount {
				topPlayer = player
				topCount = count
			}
		}
		description += fmt.Sprintf("📉 MOST FALL DEATHS\n**%s**: %d fall deaths\n\n", topPlayer, topCount)
	}

	suicideStats, err := b.db.CountDeathsBySuicide(game, today)
	if err != nil {
		b.AddDebug(fmt.Sprintf("count deaths by suicide: %v", err))
	}

	if suicideStats != nil && suicideStats.Total > 0 {
		var topPlayer string
		var topCount int
		for player, count := range suicideStats.PlayerNameCounter {
			if count > topCount {
				topPlayer = player
				topCount = count
			}
		}
		description += fmt.Sprintf("💔 MOST SUICIDES\n**%s**: %d suicides\n\n", topPlayer, topCount)
	}

	causeStats, err := b.db.GetMostCommonCauseOfDeath(game, today)
	if err != nil {
		b.AddDebug(fmt.Sprintf("get most common cause of death: %v", err))
	}

	if causeStats != nil && causeStats.DeathCause != "" {
		description += fmt.Sprintf("💀 MOST COMMON CAUSE\n**%s** (%d deaths)", causeStats.DeathCause, causeStats.Total)
		if len(causeStats.PlayerNameCounter) > 0 {
			description += " — "
			first := true
			for player, count := range causeStats.PlayerNameCounter {
				if !first {
					description += ", "
				}
				description += fmt.Sprintf("%s (%d)", player, count)
				first = false
			}
		}
		description += "\n\n"
	}

	killerStats, err := b.db.GetMostCommonKiller(game, today)
	if err != nil {
		b.AddDebug(fmt.Sprintf("get most common killer: %v", err))
	}

	if killerStats != nil && killerStats.KillerName != "" {
		description += fmt.Sprintf("👹 DEADLIEST ENEMY\n**%s** — %d total kills", killerStats.KillerName, killerStats.Total)
		if len(killerStats.PlayerNameCounter) > 0 {
			description += " — "
			first := true
			for player, count := range killerStats.PlayerNameCounter {
				if !first {
					description += ", "
				}
				description += fmt.Sprintf("%s (%d)", player, count)
				first = false
			}
		}
		description += "\n"
	}

	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       title,
					Description: description,
					Color:       0xFF0000,
				},
			},
		},
	}
}

func (b *Bot) HandleRoast(s *discordgo.Session, i *discordgo.InteractionCreate) *discordgo.InteractionResponse {
	if b.llm == nil {
		b.AddDebug("roast AI client not configured")
		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "AI client not configured",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		}); err != nil {
			fmt.Printf("Error responding to interaction: %v\n", err)
		}
		return nil
	}

	user := i.ApplicationCommandData().Options[0].UserValue(s)
	discordUserID := user.ID

	link, err := b.db.GetPlayerLinkByDiscord(discordUserID, b.cfg.Game)
	if err != nil {
		b.AddDebug(fmt.Sprintf("roast GetPlayerLinkByDiscord: %v", err))
		if err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Something went wrong!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		}); err != nil {
			fmt.Printf("Error responding to interaction: %v\n", err)
		}
		return nil
	}

	if link == nil {
		b.AddDebug(fmt.Sprintf("unknown discord user: %s, %s", user.GlobalName, user.ID))
		if err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Player not yet registered",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		}); err != nil {
			fmt.Printf("Error responding to interaction: %v\n", err)
		}
		return nil
	}

	stats, err := b.db.GetPlayerStats(b.cfg.Game, link.PlayerName)
	if err != nil {
		b.AddDebug(fmt.Sprintf("roast GetPlayerStats: %v", err))
		if err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Something went wrong!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		}); err != nil {
			fmt.Printf("Error responding to interaction: %v\n", err)
		}
		return nil
	}

	prompt := fmt.Sprintf(`Roast this player based on their stats in game.
Game: %s
Player: %s
Total Deaths: %d
of which %d were self-inflicted
Ranking on alltime deaths leaderboard: %d
Most deaths were inflicted by: %s

Requirements:
- 3-4 sentences max
- Be savage and funny. Its a roast so make it hurt while trying to be funny
- Comment on their poor gaming skills
- Only mention the death leaderboard if top 3
- Format: just the roast message, no quotes or extra text
- Player name should be wrapped in ** **`, b.cfg.Game, link.PlayerName, stats.TotalDeaths, stats.Suicides, stats.Rank, stats.DeadliestEnemy)
	roast, err := b.llm.Ask(context.Background(), prompt)
	if err != nil {
		b.AddDebug(fmt.Sprintf("roast llm.Ask: %v", err))
		if err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Something went wrong!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		}); err != nil {
			fmt.Printf("Error responding to interaction: %v\n", err)
		}
		return nil
	}

	roast = strings.Replace(roast, fmt.Sprintf("**%s**", link.PlayerName), user.Mention(), -1)
	if err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: roast,
		},
	}); err != nil {
		fmt.Printf("Error responding to interaction: %v\n", err)
	}

	return nil
}

func (b *Bot) HandleStats(s *discordgo.Session, i *discordgo.InteractionCreate) *discordgo.InteractionResponse {
	user := i.ApplicationCommandData().Options[0].UserValue(s)

	link, err := b.db.GetPlayerLinkByDiscord(user.ID, b.cfg.Game)
	if err != nil {
		b.AddDebug(fmt.Sprintf("stats GetPlayerLinkByDiscord: %v", err))
		return &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Something went wrong!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		}
	}

	if link == nil {
		b.AddDebug(fmt.Sprintf("unknown discord user: %s, %s", user.GlobalName, user.ID))
		return &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Player not yet registered",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		}
	}

	stats, err := b.db.GetPlayerStats(b.cfg.Game, link.PlayerName)
	if err != nil {
		b.AddDebug(fmt.Sprintf("stats GetPlayerStats: %v", err))
		return &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Something went wrong!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		}
	}

	description := ""

	if stats.TotalDeaths == 0 {
		description = "No deaths recorded yet!"
	} else {
		description = fmt.Sprintf("💀 **Total Deaths:** %d\n", stats.TotalDeaths)

		lastDeath, err := b.db.GetLastDeath(b.cfg.Game, link.PlayerName)
		if err != nil {
			b.AddDebug(fmt.Sprintf("stats GetLastDeath: %v", err))
		}
		if lastDeath != nil {
			description += fmt.Sprintf("last on %s: %s\n", lastDeath.Timestamp.Format("2006-01-02 15:04"), lastDeath.String())
		}

		description += "\n"

		description += fmt.Sprintf("📈 **Rank:** #%d on leaderboard\n", stats.Rank)

		if stats.FallDeaths > 0 {
			description += fmt.Sprintf("📉 **Fall Deaths:** %d\n", stats.FallDeaths)
		}

		if stats.Suicides > 0 {
			description += fmt.Sprintf("💔 **Suicides:** %d\n", stats.Suicides)
		}

		if stats.MostCommonCause != "" {
			description += fmt.Sprintf("💀 **Most Common Cause:** %s\n", stats.MostCommonCause)
		}

		if stats.DeadliestEnemy != "" {
			description += fmt.Sprintf("👹 **Deadliest Enemy:** %s\n", stats.DeadliestEnemy)
		}
	}

	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       fmt.Sprintf("📊 Stats for **%s**", link.PlayerName),
					Description: description,
					Color:       0x00FF00,
				},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	}
}
