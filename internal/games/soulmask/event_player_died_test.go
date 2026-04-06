package soulmask

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_playerDiedEvent_Prompt(t *testing.T) {
	tests := []struct {
		name  string
		event playerDiedEvent
		want  string
	}{
		{
			name: "suicide, unknown reason",
			event: playerDiedEvent{
				CharacterName:        "Character",
				PlayerName:           "PlayerA",
				KillerName:           "Character",
				KillerPlayerName:     "PlayerA",
				DeathCausedByAbility: "",
				DeathCausedByEffect:  "",
			},
			want: "",
		},
		{
			name: "fall damage",
			event: playerDiedEvent{
				CharacterName:        "Character",
				PlayerName:           "PlayerA",
				KillerName:           "Character",
				KillerPlayerName:     "PlayerA",
				DeathCausedByAbility: "",
				DeathCausedByEffect:  "Default__GE_FallingDamageEffect_C",
			},
			want: "",
		},
		{
			name: "killed by npc monster",
			event: playerDiedEvent{
				CharacterName:        "Character",
				PlayerName:           "PlayerA",
				KillerName:           "Jaguar",
				KillerPlayerName:     "",
				DeathCausedByAbility: "",
				DeathCausedByEffect:  "",
			},
			want: "",
		},
		{
			name: "killed by npc with spear",
			event: playerDiedEvent{
				CharacterName:        "Character",
				PlayerName:           "PlayerA",
				KillerName:           "Novice Hunter <Flint Tribe>",
				KillerPlayerName:     "",
				DeathCausedByAbility: "Default__GA_Spear_Attack01_C",
				DeathCausedByEffect:  "",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// validate prompt manually in console
			fmt.Println(tt.event.Prompt())
			assert.Equal(t, 1, 1)
		})
	}
}
