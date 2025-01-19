package message

import "time"

type Status uint8

const (
	New Status = iota + 1
	Process
	Sent
	Fail
	Dead
)

func (s Status) String() string {
	switch s {
	case New:
		return "New"
	case Process:
		return "Process"
	case Sent:
		return "Sent"
	case Fail:
		return "Fail"
	case Dead:
		return "Dead"
	default:
		return "Unknown"
	}
}

type Message struct {
	PhoneNumber string
	Content     string
	Status
}

type CreatedMessage struct {
	CreatedAt   *time.Time
	Id          string
	PhoneNumber string
	Content     string
	Status
}
