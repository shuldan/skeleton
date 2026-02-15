package model

import "strings"

// Title — value object заголовка задачи.
type Title struct {
	value string
}

// NewTitle валидирует и создаёт заголовок.
func NewTitle(raw string) (Title, error) {
	if strings.TrimSpace(raw) == "" {
		return Title{}, ErrTitleRequired
	}

	return Title{value: raw}, nil
}

func (t Title) String() string { return t.value }
