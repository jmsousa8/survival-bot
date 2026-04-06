package soulmask

import (
	"regexp"

	"survival-bot/internal/events"
)

type Parser struct {
	playerDeathRegex *regexp.Regexp
}

func New() *Parser {
	return &Parser{
		playerDeathRegex: regexp.MustCompile(`Name = (.+?) \((.+?)\).*Killer = (.+?) (?:\(\((.+?)\)\)|\(\)).*(?:GA = (.*?), GE = (.*)|GA = , GE = )`),
	}
}

func (p *Parser) ParseLine(line string) events.Event {
	if matches := p.playerDeathRegex.FindStringSubmatch(line); len(matches) == 7 {
		return newPlayerDiedEvent(line, matches[1], matches[2], matches[3], matches[4], matches[5], matches[6])
	}
	return nil
}
