package model

import domainerrors "github.com/shuldan/errors"

// status — VO статуса задачи (требование №2 — методы неэкспортируемые).
type status struct {
	value string
}

var (
	statusDraft      = status{value: "draft"}
	statusInProgress = status{value: "in_progress"}
	statusDone       = status{value: "done"}
)

// StatusFromString восстанавливает status из строки (для Restore).
func statusFromString(s string) status {
	return status{value: s}
}

var transitions = map[string][]status{
	"draft":       {statusInProgress, statusDone},
	"in_progress": {statusDone},
}

// TransitionTo проверяет допустимость и возвращает новый статус.
func (s status) transitionTo(target status) (status, error) {
	for _, allowed := range transitions[s.value] {
		if allowed.equals(target) {
			return target, nil
		}
	}

	return s, ErrInvalidStatusTransition.WithDetails(domainerrors.D{
		"from": string(s.value),
		"to":   string(target.value),
	})
}

func (s status) String() string {
	return s.value
}

func (s status) equals(other status) bool {
	return s.value == other.value
}
