package soulmask

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	"survival-bot/internal/events"

	"github.com/bwmarrin/discordgo"
)

type playerDiedEvent struct {
	Timestamp            time.Time
	CharacterName        string
	PlayerName           string
	KillerName           string
	KillerPlayerName     string
	DeathCausedByAbility string
	DeathCausedByEffect  string
	RawSource            string
	Message              string
}

const maxInputLength = 30

var (
	abilities = map[string]string{
		"Default__GA_Bow_Skill03_C":           "Killer used a bow",
		"Default__GA_BigSword_Attack01_C":     "Killer used a great sword",
		"Default__GA_BigSword_Attack01_2_C":   "Killer used a great sword",
		"Default__GA_BigSword_JumpAttack_C":   "Killer used a great sword",
		"Default__GA_Blade_Attack01_6_C":      "Killer used a blade",
		"Default__GA_Spear_Attack01_C":        "Killer used a spear",
		"Default__GA_Spear_Attack01_Shield_C": "Killer used a spear and shield",
		"Default__GA_Spear_Attack02_C":        "Killer used a spear",
		"Default__GA_Fist_Attack02_C":         "Killer used gauntlets",
	}
	effects = map[string]string{
		"Default__GE_Water_Empty_Damage_ZuoQi_C": "Died of thirst",
		"Default__GE_Food_Empty_Damage_C":        "Died of hunger",
		"Default__GE_FallingDamageEffect_C":      "Died of fall damage",
	}
)

func truncate(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen]
	}
	return s
}

func newPlayerDiedEvent(raw, rawCharacter, rawPlayer, rawKiller, rawPKiller, ability, effect string) *playerDiedEvent {
	deadPlayer := truncate(strings.TrimSpace(rawPlayer), maxInputLength)
	killer := truncate(strings.TrimSpace(rawKiller), maxInputLength)
	killerPlayer := truncate(strings.TrimSpace(rawPKiller), maxInputLength)

	return &playerDiedEvent{
		Timestamp:            time.Now(),
		CharacterName:        truncate(strings.TrimSpace(rawCharacter), maxInputLength),
		PlayerName:           deadPlayer,
		KillerName:           killer,
		KillerPlayerName:     killerPlayer,
		DeathCausedByAbility: ability,
		DeathCausedByEffect:  effect,
		RawSource:            raw,
		Message:              generateMessage(deadPlayer, killer, killerPlayer),
	}
}

func (e *playerDiedEvent) Type() events.EventType {
	return events.EventPlayerDied
}

func (e *playerDiedEvent) Raw() string {
	return e.RawSource
}

func (e *playerDiedEvent) GetMessage() string {
	return e.Message
}

func (e *playerDiedEvent) SetMessage(message string) {
	e.Message = message
}

func (e *playerDiedEvent) GetPlayer() string {
	return e.PlayerName
}

func (e *playerDiedEvent) GetKiller() string {
	if e.KillerPlayerName != "" {
		return e.KillerPlayerName
	}
	return e.KillerName
}

func (e *playerDiedEvent) GetDeathCause() string {
	if e.DeathCausedByEffect != "" {
		if reason, ok := effects[e.DeathCausedByEffect]; ok {
			return reason
		}
	}
	if e.DeathCausedByAbility != "" {
		if reason, ok := abilities[e.DeathCausedByAbility]; ok {
			return reason
		}
	}
	return ""
}

func (e *playerDiedEvent) ToDiscordEmbed() *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       "Player Died",
		Description: e.GetMessage(),
		Color:       0xFF0000,
	}
}

func (e *playerDiedEvent) Prompt() string {
	return fmt.Sprintf(`Generate a short, dramatic message about a death that occured in game.
%s
Requirements:
- 1-2 sentences max
- Game: Soulmask (https://soulmask.fandom.com/wiki/Soulmask)
- Include 1 relevant emoji
- ROAST THEM HARD for being so bad at the game!
- Format: just the message, no quotes or extra text, player and killer names should be wrapped in ** **
- If player killed themselves, make a self-deprecating joke
- If NPC/animal, make it about the threat

Just output the message, nothing else.`, e.deathSummary())
}

func generateMessage(playerName, killerName, killerPlayerName string) string {
	var deathMessages = []string{
		"⚔ **{player}** was slain by {killer}",
		"🩸 **{player}** didn't survive the encounter with {killer}!",
		"😵 **{player}** was eliminated by {killer}",
	}
	var suicideMessages = []string{
		"🪦 **{player}** has fallen!",
		"💀 **{player}** bit the dust...",
		"🪦 **{player}** is now a ghost...",
	}

	if strings.ToLower(playerName) == strings.ToLower(killerPlayerName) {
		return strings.Replace(suicideMessages[rand.IntN(2)], "{player}", playerName, 1)
	}
	return strings.Replace(strings.Replace(deathMessages[rand.IntN(2)], "{player}", playerName, 1), "{killer}", killerName, 1)
}

func (e *playerDiedEvent) deathSummary() string {
	deathSummary := fmt.Sprintf(`Player "%s" died. `, e.PlayerName)
	if e.PlayerName == e.KillerPlayerName {
		deathSummary += "Player killed themselves. "
		if reason, ok := effects[e.DeathCausedByEffect]; ok {
			deathSummary += reason
		}
	} else if e.KillerPlayerName == "" {
		deathSummary += fmt.Sprintf(`Was killed by "%s". `, e.KillerName)
		if reason, ok := abilities[e.DeathCausedByAbility]; ok {
			deathSummary += reason
		}
	} else {
		deathSummary += fmt.Sprintf(`Was killed by "%s". `, e.KillerPlayerName)
	}
	return deathSummary + "\n"
}
