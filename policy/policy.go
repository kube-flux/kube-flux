package policy

// Status represents Policy energy consumption status: Green, Brown, Black
type Status string

const (
	Green = "Green"
	Brown = "Brown"
	Black = "Black"
)

// Policy defines the object that maintains energy related status
type Policy struct {
	Status    Status
	CreatedAt string
	UpdatedAt string
}

// PolicyTracker defines the API interface of the PolicyReceiver(Zeus)
type PolicyTracker interface {

	// GetPolicy gets the policy
	GetPolicy() (Policy, error)

	// UpdatePolicy updates the policy with given status
	UpdatePolicy(status uint8) (Policy, error)
}
