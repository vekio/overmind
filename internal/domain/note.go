package domain

// Note is a supported note that exposes its common metadata.
type Note interface {
	Metadata() Metadata
}
