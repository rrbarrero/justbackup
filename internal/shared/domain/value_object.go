package domain

// ValueObject is the interface that all value objects should implement.
// In Go, we can't enforce immutability easily, but this marker interface
// helps in identifying value objects.
type ValueObject interface {
	Equals(other ValueObject) bool
}
