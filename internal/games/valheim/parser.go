package valheim

import (
	"regexp"

	"survival-bot/internal/events"
)

type Parser struct {
	deathRegex *regexp.Regexp
}

func New() *Parser {
	return &Parser{
		deathRegex: regexp.MustCompile(`Got character ZDOID from (.+?) : \d+:0`),
	}
}

func (p *Parser) ParseLine(line string) events.Event {
	if matches := p.deathRegex.FindStringSubmatch(line); len(matches) == 2 {
		return newPlayerDiedEvent(line, matches[1])
	}
	return nil
}
