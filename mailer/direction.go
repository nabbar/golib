package mailer

import (
	"strings"

	"github.com/matcornic/hermes/v2"
)

type TextDirection uint8

const (
	LeftToRight TextDirection = iota
	RightToLeft
)

func (d TextDirection) getDirection() hermes.TextDirection {
	switch d {
	case LeftToRight:
		return hermes.TDLeftToRight
	case RightToLeft:
		return hermes.TDRightToLeft
	}

	return LeftToRight.getDirection()
}

func (d TextDirection) String() string {
	switch d {
	case LeftToRight:
		return "Left->Right"
	case RightToLeft:
		return "Right->Left"
	}

	return LeftToRight.String()
}

func ParseTextDirection(direction string) TextDirection {
	d := strings.ToLower(direction)

	l := strings.Index(d, "left")
	r := strings.Index(d, "right")

	if l > 0 && r > 0 && l > r {
		return RightToLeft
	} else {
		return LeftToRight
	}
}
