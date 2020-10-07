package policy

import "time"

// Status represents Policy energy consumption status: Green, Brown, Black
type Status int

const (
	Green = iota
	Brown
	Black
)

// Policy defines the object that maintains energy related status
type Policy struct {
	Status    Status
	CreatedAt time.Time
	UpdatedAt time.Time
}

// PolicyTracker defines the API interface of the PolicyReceiver(Zeus)
type PolicyTracker interface {
	// CreatePolicy creates a new policy with default Green status
	CreatePolicy() (Policy, error)

	// UpdatePolicy updates the policy with given status
	UpdatePolicy(status uint8) (Policy, error)

	// GetPolicy gets the policy
	GetPolicy() (Policy, error)

	// DeletePolicy deletes the policy
	DeletePolicy() error
}
