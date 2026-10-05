package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	appconfig "github.com/vekio/overmind/internal/config"
)

// ErrSetupCancelled means setup exited without submitting valid settings.
var ErrSetupCancelled = errors.New("setup cancelled")

// RunSetup opens the shared form with editable defaults and returns the submitted settings.
// The CLI creates the configuration only after this form has completed successfully.
func RunSetup(ctx context.Context, defaults appconfig.Settings, path string) (appconfig.Settings, error) {
	initial := newSetupModel(defaults, path)
	result, err := tea.NewProgram(initial, tea.WithContext(ctx)).Run()
	if err != nil {
		return appconfig.Settings{}, err
	}
	completed := result.(setupModel)
	if !completed.submitted {
		return appconfig.Settings{}, ErrSetupCancelled
	}
	return completed.settings, nil
}

type setupModel struct {
	form          form
	settings      appconfig.Settings
	submitted     bool
	width, height int
}

func newSetupModel(defaults appconfig.Settings, path string) setupModel {
	if defaults.Mode == "" {
		defaults.Mode = appconfig.ModeLocal
	}
	f := newForm("Setup",
		fieldSpec{
			id:          "mode",
			label:       "Mode",
			kind:        fieldText,
			placeholder: "Required, local",
			validate: func(raw string) error {
				if appconfig.Mode(strings.TrimSpace(raw)) != appconfig.ModeLocal {
					return fmt.Errorf("mode must be local")
				}
				return nil
			},
		},
		fieldSpec{
			id:          "vault",
			label:       "Vault",
			kind:        fieldText,
			placeholder: "Required, directory for notes and index",
			validate: func(raw string) error {
				_, err := appconfig.NormalizeVaultPath(raw)
				return err
			},
		},
	)
	setFormValue(&f, "mode", string(defaults.Mode))
	setFormValue(&f, "vault", defaults.VaultPath)
	f.detail = "Config > " + path
	f.Resize(80, 24)
	return setupModel{form: f, width: 80, height: 24}
}

// Init focuses the first editable configuration field.
func (m setupModel) Init() tea.Cmd { return m.form.Focus() }

// Update validates submitted settings and retains the form when validation fails.
// Cancellation exits without producing settings for the CLI to write.
func (m setupModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.form.Resize(m.width, m.height)
		return m, nil
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case formCancelled:
		return m, tea.Quit
	case formSubmitted:
		vault, err := appconfig.NormalizeVaultPath(msg.values["vault"])
		if err != nil {
			m.form.problem = err.Error()
			return m, m.form.Focus()
		}
		settings := appconfig.Settings{
			Mode:      appconfig.Mode(strings.TrimSpace(msg.values["mode"])),
			VaultPath: vault,
		}
		if err := settings.Validate(); err != nil {
			m.form.problem = err.Error()
			return m, m.form.Focus()
		}
		m.settings, m.submitted = settings, true
		return m, tea.Quit
	}
	var cmd tea.Cmd
	m.form, cmd = m.form.Update(message)
	return m, cmd
}

// View renders editable setup values and the form's contextual controls.
func (m setupModel) View() tea.View {
	footer := ansi.Wrap(formatControls(m.form.Controls()), max(1, m.width), "")
	view := tea.NewView(withFooter(m.form.View(), footer, m.height))
	view.AltScreen = true
	return view
}
