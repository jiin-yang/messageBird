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
	Id          string
	PhoneNumber string
	Content     string
	Status
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

type CreateMessage struct {
	PhoneNumber string
	Content     string
	Status
}

type CreatedMessageDbResponse struct {
	CreatedAt   *time.Time
	Id          string
	PhoneNumber string
	Content     string
	Status
}
