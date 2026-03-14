package model

import "strings"

// Title — VO заголовка задачи.
// Все методы неэкспортируемые (требования №2, №19).
type Title struct {
	value string
}

// NewTitle валидирует формат и создаёт заголовок (требование №9, №18).
func NewTitle(raw string) (Title, error) {
	if strings.TrimSpace(raw) == "" {
		return Title{}, ErrTitleRequired
	}

	return Title{value: raw}, nil
}

func (t Title) String() string {
	return t.value
}
