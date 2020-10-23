package midterm

// Status represents Policy energy consumption status: Green, Yellow, Red
type Status string

const (
	Green Status = "Green"
	Yellow Status = "Yellow"
	Red Status = "Red"
)

// Policy defines the object that maintains energy related status
type Policy struct {
	Status    Status
	UpdatedAt string
}
