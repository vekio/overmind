package ports

import "context"

// Renderer renders a named template with the supplied data.
type Renderer interface {
	Render(ctx context.Context, name string, data any) ([]byte, error)
}
