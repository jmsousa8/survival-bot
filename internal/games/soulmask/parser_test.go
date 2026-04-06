package soulmask

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParser_ParseLine(t *testing.T) {
	p := New()

	t.Run("player killed by npc monster", func(t *testing.T) {
		logLine := "[2026.02.14-15.02.21:016][605]LogWS: Warning: 死亡日志 : You Are Dead, Name = Vagabond ( Skk1p ) [主人:(SteamId:123456789 Uid:xxxxxxxxxxxxxxxxxxxxxxxxxxx) ], Killer = Bush Dog () [ ], GA = Default__GA_SouQuan_Bite_C, GE = Default__DamageType_Default_C"
		got := p.ParseLine(logLine)
		assert.IsType(t, &playerDiedEvent{}, got)
		ev := got.(*playerDiedEvent)
		assert.Equal(t, "Vagabond", ev.CharacterName)
		assert.Equal(t, "Skk1p", ev.PlayerName)
		assert.Equal(t, "Bush Dog", ev.KillerName)
		assert.Equal(t, "", ev.KillerPlayerName)
	})

	t.Run("player killed by npc tribesman", func(t *testing.T) {
		logLine := "[2026.02.15-21.41.36:362][426]LogWS: Warning: 死亡日志 : You Are Dead, Name = BigYoinks ( BigYoinks ) [主人:(SteamId:123456789 Uid:xxxxxxxxxxxxxxxxxxxxxxxxxxx) ], Killer = Novice Hunter <Flint Tribe> () [ ], GA = Default__GA_Spear_Attack01_C, GE = Default__DamageType_Jab_C"
		got := p.ParseLine(logLine)
		assert.IsType(t, &playerDiedEvent{}, got)
		ev := got.(*playerDiedEvent)
		assert.Equal(t, "BigYoinks", ev.CharacterName)
		assert.Equal(t, "BigYoinks", ev.PlayerName)
		assert.Equal(t, "Novice Hunter <Flint Tribe>", ev.KillerName)
		assert.Equal(t, "", ev.KillerPlayerName)
	})

	t.Run("player killed themselves", func(t *testing.T) {
		logLine := "[2026.02.15-20.45.33:077][246]LogWS: Warning: 死亡日志 : You Are Dead, Name = BigYoinkers ( BigYoinkers ) [主人:(SteamId:123456789 Uid:xxxxxxxxxxxxxxxxxxxxxxxxxxx) ], Killer = BigYoinkers (( BigYoinkers )) [主人:(SteamId:123456789 Uid:xxxxxxxxxxxxxxxxxxxxxxxxxxx) ], GA = , GE = "
		got := p.ParseLine(logLine)
		assert.IsType(t, &playerDiedEvent{}, got)
		ev := got.(*playerDiedEvent)
		assert.Equal(t, "BigYoinkers", ev.CharacterName)
		assert.Equal(t, "BigYoinkers", ev.PlayerName)
		assert.Equal(t, "BigYoinkers", ev.KillerName)
		assert.Equal(t, "BigYoinkers", ev.KillerPlayerName)
	})
}
