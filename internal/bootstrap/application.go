package bootstrap

import "git.casta.me/alberto/overmind/internal/app"

// Application is the composed local application and its owned resources.
type Application struct {
	app.Application
	close func() error
}

// Close releases resources owned by the application.
func (application Application) Close() error {
	if application.close == nil {
		return nil
	}
	return application.close()
}
