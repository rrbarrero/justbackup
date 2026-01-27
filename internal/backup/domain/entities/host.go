package entities

import (
	"time"

	"github.com/google/uuid"
	shared "github.com/rrbarrero/justbackup/internal/shared/domain"
)

type HostID struct {
	value string
}

func NewHostID() HostID {
	return HostID{value: uuid.New().String()}
}

func NewHostIDFromString(id string) (HostID, error) {
	if _, err := uuid.Parse(id); err != nil {
		return HostID{}, shared.ErrInvalidID
	}
	return HostID{value: id}, nil
}

func (id HostID) String() string {
	return id.value
}

type Host struct {
	shared.AggregateRoot
	id            HostID
	name          string
	hostname      string
	user          string
	port          int
	path          string
	isWorkstation bool
	createdAt     time.Time
}

func NewHost(name, hostname, user string, port int, path string, isWorkstation bool) *Host {
	return &Host{
		id:            NewHostID(),
		name:          name,
		hostname:      hostname,
		user:          user,
		port:          port,
		path:          path,
		isWorkstation: isWorkstation,
		createdAt:     NowFunc(),
	}
}

func NewHostWithID(id HostID, name, hostname, user string, port int, path string, isWorkstation bool) *Host {
	return &Host{
		id:            id,
		name:          name,
		hostname:      hostname,
		user:          user,
		port:          port,
		path:          path,
		isWorkstation: isWorkstation,
		createdAt:     NowFunc(),
	}
}

func RestoreHost(id HostID, name, hostname, user string, port int, path string, isWorkstation bool, createdAt time.Time) *Host {
	return &Host{
		id:            id,
		name:          name,
		hostname:      hostname,
		user:          user,
		port:          port,
		path:          path,
		isWorkstation: isWorkstation,
		createdAt:     createdAt,
	}
}

func (h *Host) ID() HostID {
	return h.id
}

func (h *Host) Name() string {
	return h.name
}

func (h *Host) Hostname() string {
	return h.hostname
}

func (h *Host) User() string {
	return h.user
}

func (h *Host) Port() int {
	return h.port
}

func (h *Host) Path() string {
	return h.path
}

func (h *Host) IsWorkstation() bool {
	return h.isWorkstation
}

func (h *Host) CreatedAt() time.Time {
	return h.createdAt
}

func (h *Host) Update(name, hostname, user string, port int, path string, isWorkstation bool) {
	h.name = name
	h.hostname = hostname
	h.user = user
	h.port = port
	h.path = path
	h.isWorkstation = isWorkstation
}
