package model

import domainerrors "github.com/shuldan/errors"

// Status — value object статуса задачи.
type Status string

const (
	StatusDraft      Status = "draft"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

var transitions = map[Status][]Status{
	StatusDraft:      {StatusInProgress, StatusDone},
	StatusInProgress: {StatusDone},
}

// TransitionTo проверяет допустимость и возвращает новый статус.
func (s Status) TransitionTo(target Status) (Status, error) {
	for _, allowed := range transitions[s] {
		if allowed == target {
			return target, nil
		}
	}

	return s, ErrInvalidStatusTransition.WithDetails(domainerrors.D{
		"from": string(s),
		"to":   string(target),
	})
}

func (s Status) String() string { return string(s) }
