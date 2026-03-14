package model

// status — VO статуса счёта.
type status struct {
	value string
}

var statusCreated = status{value: "created"}

func statusFromString(s string) status {
	return status{value: s}
}

func (s status) String() string {
	return s.value
}
