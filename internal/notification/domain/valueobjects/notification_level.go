package valueobjects

type NotificationLevel string

const (
	Info    NotificationLevel = "info"
	Warning NotificationLevel = "warning"
	Error   NotificationLevel = "error"
)

func (l NotificationLevel) String() string {
	return string(l)
}
