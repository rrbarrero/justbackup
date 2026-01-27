package domain

// AggregateRoot is the base struct for all aggregate roots.
// It handles the recording of domain events.
type AggregateRoot struct {
	events []DomainEvent
}

// RecordEvent records a new domain event.
func (a *AggregateRoot) RecordEvent(event DomainEvent) {
	a.events = append(a.events, event)
}

// PullEvents returns all recorded events and clears the list.
func (a *AggregateRoot) PullEvents() []DomainEvent {
	events := a.events
	a.events = []DomainEvent{}
	return events
}
