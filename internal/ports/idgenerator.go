package ports

// IDGenerator creates unique identifiers.
type IDGenerator interface {
	Generate() (string, error)
}
