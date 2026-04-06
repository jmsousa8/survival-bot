package valheim

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParser_ParseLine(t *testing.T) {
	p := New()

	t.Run("player death", func(t *testing.T) {
		logLine := "Got character ZDOID from PlayerOne : 0:0"
		got := p.ParseLine(logLine)
		assert.IsType(t, &playerDiedEvent{}, got)
		ev := got.(*playerDiedEvent)
		assert.Equal(t, "PlayerOne", ev.PlayerName)
	})

	t.Run("player death with steam id", func(t *testing.T) {
		logLine := "Got character ZDOID from Viking123 : 76561198012345678:0:0"
		got := p.ParseLine(logLine)
		assert.IsType(t, &playerDiedEvent{}, got)
		ev := got.(*playerDiedEvent)
		assert.Equal(t, "Viking123", ev.PlayerName)
	})

	t.Run("not a death line (1:0 means alive)", func(t *testing.T) {
		logLine := "Got character ZDOID from PlayerOne : 76561198012345678:1:0"
		got := p.ParseLine(logLine)
		assert.Nil(t, got)
	})

	t.Run("random log line", func(t *testing.T) {
		logLine := "Some random log line"
		got := p.ParseLine(logLine)
		assert.Nil(t, got)
	})
}
