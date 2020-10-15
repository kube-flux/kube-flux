package midterm

// Status represents Policy energy consumption status: Green, Brown, Black
type Status string

const (
	Green Status = "Green"
	Brown Status = "Brown"
	Black Status = "Black"
)

// Policy defines the object that maintains energy related status
type Policy struct {
	Status    Status
	UpdatedAt string
}
