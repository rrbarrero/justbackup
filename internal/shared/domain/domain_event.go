package domain

import "time"

// DomainEvent is the interface that all domain events must implement.
type DomainEvent interface {
	Name() string
	OccurredOn() time.Time
}
