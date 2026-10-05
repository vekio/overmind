package tui

import (
	"strings"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/vekio/overmind/internal/domain/shared"
)

type fieldKind uint8

const (
	fieldText fieldKind = iota
	fieldMultiline
	fieldTags
	fieldStatic
	fieldSelect
	fieldGroups
)

// fieldSpec describes data independently of its input widget and layout.
type fieldSpec struct {
	id, label, placeholder string
	kind                   fieldKind
	validate               func(string) error
	value                  string
	options                []string
	// rows is the preferred textarea height, capped by the available terminal space.
	rows int
}

type formField struct {
	spec     fieldSpec
	input    textinput.Model
	textarea textarea.Model
	problem  string
	selected int
	groups   groupPicker
}

// form owns widget state and local validation. Submission emits a snapshot of
// field values; application handlers remain responsible for domain validation.
type form struct {
	title   string
	fields  []formField
	focus   int
	width   int
	height  int
	saving  bool
	problem string
	detail  string
}

type formSubmitted struct{ values map[string]string }
type formCancelled struct{}

func newForm(title string, specs ...fieldSpec) form {
	f := form{title: title, width: 78, height: 24}
	for _, spec := range specs {
		field := formField{spec: spec}
		if spec.kind == fieldStatic || spec.kind == fieldSelect {
			f.fields = append(f.fields, field)
			continue
		} else if spec.kind == fieldMultiline {
			field.textarea = textarea.New()
			field.textarea.ShowLineNumbers = false
			field.textarea.MaxHeight = 0
			field.textarea.Placeholder = spec.placeholder
		} else {
			field.input = textinput.New()
			field.input.SetVirtualCursor(true)
			field.input.Prompt = spec.label + " > "
			field.input.Placeholder = spec.placeholder
		}
		f.fields = append(f.fields, field)
	}
	return f
}

// Focus blurs other inputs and activates the current editable widget.
func (f *form) Focus() tea.Cmd {
	for i := range f.fields {
		if f.fields[i].spec.kind == fieldStatic || f.fields[i].spec.kind == fieldSelect {
			continue
		}
		if f.fields[i].spec.kind == fieldMultiline {
			f.fields[i].textarea.Blur()
		} else {
			f.fields[i].input.Blur()
		}
	}
	if f.focus >= len(f.fields) {
		return nil
	}
	if f.fields[f.focus].spec.kind == fieldStatic {
		for range len(f.fields) {
			f.focus = (f.focus + 1) % len(f.fields)
			if f.fields[f.focus].spec.kind != fieldStatic {
				return f.Focus()
			}
		}
		return nil
	}
	f.Resize(f.width+2, f.height)
	if f.fields[f.focus].spec.kind == fieldMultiline {
		return f.fields[f.focus].textarea.Focus()
	}
	if f.fields[f.focus].spec.kind == fieldSelect {
		return nil
	}
	return f.fields[f.focus].input.Focus()
}

// Resize fits widgets into the available area while retaining textarea content.
func (f *form) Resize(width, height int) {
	width = max(4, min(88, width-2))
	f.width = width
	f.height = height
	multiline := 0
	for _, field := range f.fields {
		if field.spec.kind == fieldMultiline {
			multiline++
		}
	}
	controls := f.Controls()
	if f.expandedContent() && f.fields[f.focus].spec.kind != fieldMultiline {
		controls = append(controls, controlHint{"enter", "new line"})
	}
	footerRows := lipgloss.Height(ansi.Wrap(formatControls(controls), width, ""))
	reservedFields := len(f.fields)
	if f.expandedContent() {
		// Page and Journal share a height budget, including metadata and controls.
		reservedFields = max(4, reservedFields)
		footerRows = max(2, footerRows)
	}
	extraRows := 0
	if f.focus < len(f.fields) && f.fields[f.focus].spec.kind == fieldGroups {
		extraRows = max(0, lipgloss.Height(ansi.Wrap(f.fields[f.focus].groupsView("Groups", true), width, ""))-1)
	}
	areaHeight := max(1, (height-2-footerRows-3*reservedFields-multiline-extraRows)/max(1, multiline))
	for i := range f.fields {
		if f.fields[i].spec.kind == fieldStatic || f.fields[i].spec.kind == fieldSelect {
			continue
		}
		if f.fields[i].spec.kind == fieldMultiline {
			f.fields[i].textarea.SetWidth(width)
			desired := f.fields[i].spec.rows
			if desired == 0 {
				desired = 5
			}
			f.fields[i].textarea.SetHeight(min(desired, areaHeight))
		} else {
			f.fields[i].input.SetWidth(max(1, width-ansi.StringWidth(f.fields[i].input.Prompt)))
		}
	}
}

// Value returns the current field value, independent of the underlying widget.
func (field formField) Value() string {
	if field.spec.kind == fieldGroups {
		return strings.Join(field.groups.chosen, ", ")
	}
	if field.spec.kind == fieldSelect {
		if len(field.spec.options) == 0 {
			return ""
		}
		return field.spec.options[field.selected]
	}
	if field.spec.kind == fieldStatic {
		return field.spec.value
	}
	if field.spec.kind == fieldMultiline {
		return field.textarea.Value()
	}
	return field.input.Value()
}

// Update handles form navigation and delegates editing to the focused widget.
// While a save is pending, user input cannot modify or resubmit the form.
func (f form) Update(message tea.Msg) (form, tea.Cmd) {
	if f.saving {
		return f, nil
	}
	if key, ok := message.(tea.KeyPressMsg); ok {
		switch key.String() {
		case "esc":
			return f, func() tea.Msg { return formCancelled{} }
		case "tab", "shift+tab":
			if len(f.fields) == 0 {
				return f, nil
			}
			step := 1
			if key.String() == "shift+tab" {
				step = -1
			}
			f.focus = (f.focus + step + len(f.fields)) % len(f.fields)
			for range len(f.fields) {
				if f.fields[f.focus].spec.kind != fieldStatic {
					break
				}
				f.focus = (f.focus + step + len(f.fields)) % len(f.fields)
			}
			return f, f.Focus()
		case "ctrl+s":
			return f.submit()
		}
	}
	if f.focus >= len(f.fields) || f.fields[f.focus].spec.kind == fieldStatic {
		return f, nil
	}
	var cmd tea.Cmd
	field := &f.fields[f.focus]
	if field.spec.kind == fieldSelect {
		if key, ok := message.(tea.KeyPressMsg); ok && len(field.spec.options) > 0 {
			switch key.String() {
			case "left", "up":
				field.selected = (field.selected - 1 + len(field.spec.options)) % len(field.spec.options)
			case "right", "down":
				field.selected = (field.selected + 1) % len(field.spec.options)
			default:
				return f, nil
			}
			field.problem = ""
			f.problem = ""
		}
		return f, nil
	}
	before := field.Value()
	if field.spec.kind == fieldGroups {
		cmd = field.updateGroups(message)
		f.Resize(f.width+2, f.height)
	} else if field.spec.kind == fieldMultiline {
		field.textarea, cmd = field.textarea.Update(message)
	} else {
		field.input, cmd = field.input.Update(message)
	}
	if field.Value() != before {
		field.problem = ""
		f.problem = ""
	}
	return f, cmd
}

// submit validates every field and focuses the first problem. Invalid input
// leaves the widgets intact so the same draft can be corrected and retried.
func (f form) submit() (form, tea.Cmd) {
	values := make(map[string]string, len(f.fields))
	valid := true
	for i := range f.fields {
		field := &f.fields[i]
		field.problem = ""
		values[field.spec.id] = field.Value()
		validate := field.spec.validate
		switch field.spec.kind {
		case fieldTags:
			validate = validateFormTags
		case fieldGroups:
			validate = validateFormGroups
		}
		if field.spec.kind == fieldGroups && strings.TrimSpace(field.input.Value()) != "" {
			field.problem = "Press Enter to select or add the group, or clear the search"
			if valid {
				f.focus = i
			}
			valid = false
			continue
		}
		if validate != nil {
			if err := validate(field.Value()); err != nil {
				field.problem = err.Error()
				if valid {
					f.focus = i
				}
				valid = false
			}
		}
	}
	if !valid {
		return f, f.Focus()
	}
	return f, func() tea.Msg { return formSubmitted{values: values} }
}

func tagsField() fieldSpec {
	return fieldSpec{id: "tags", label: "Tags", kind: fieldTags, placeholder: "Optional, separated by commas. Example: ideas, work"}
}

// Empty input means no values; nonempty comma-separated entries follow domain rules.
func formListValues(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	values := strings.Split(raw, ",")
	for i := range values {
		values[i] = strings.TrimSpace(values[i])
	}
	return values
}

func validateFormTags(raw string) error {
	var tags []shared.Tag
	for _, value := range formListValues(raw) {
		tag, err := shared.NewTag(value)
		if err != nil {
			return err
		}
		tags = append(tags, tag)
	}
	_, err := shared.NewTags(tags...)
	return err
}

// View renders fields, separators and validation problems within the current layout.
func (f form) View() string {
	var lines strings.Builder
	lines.WriteString("Overmind / " + f.title + "\n\n")
	for i, field := range f.fields {
		label := field.spec.label
		if f.focus == i && field.spec.kind != fieldStatic {
			label = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F780E2")).Render(label)
		}
		if field.spec.kind == fieldStatic {
			lines.WriteString(label + " > " + field.Value())
		} else if field.spec.kind == fieldSelect {
			choices := make([]string, len(field.spec.options))
			for j, option := range field.spec.options {
				choices[j] = option
				if j == field.selected {
					choices[j] = "[" + option + "]"
				}
			}
			lines.WriteString(label + " > " + strings.Join(choices, "  "))
		} else if field.spec.kind == fieldGroups {
			lines.WriteString(field.groupsView(label, f.focus == i))
		} else if field.spec.kind == fieldMultiline {
			lines.WriteString(label + "\n")
			lines.WriteString(field.textarea.View())
		} else {
			field.input.Prompt = label + " > "
			lines.WriteString(field.input.View())
		}
		if field.problem != "" {
			lines.WriteString("\nError: " + field.problem)
		}
		lines.WriteString("\n\n")
	}
	if f.saving {
		lines.WriteString("\n\nSaving…")
	}
	if f.problem != "" {
		lines.WriteString("\n\nError: " + f.problem)
	}
	if f.detail != "" {
		lines.WriteString("\n" + controlDescriptionStyle.Render(f.detail))
	}
	return ansi.Wrap(lines.String(), f.width, "")
}

// Controls describes actions available for the focused field and current save state.
func (f form) Controls() []controlHint {
	if f.saving {
		return []controlHint{{"ctrl+c", "quit"}}
	}
	controls := []controlHint{{"tab/shift+tab", "field"}, {"ctrl+s", "save"}, {"esc", "cancel"}}
	if f.focus < len(f.fields) && f.fields[f.focus].spec.kind == fieldMultiline {
		controls = append(controls, controlHint{"enter", "new line"})
	} else if f.focus < len(f.fields) && f.fields[f.focus].spec.kind == fieldSelect {
		controls = append(controls, controlHint{"←/→", "choose"})
	}
	if f.focus < len(f.fields) && f.fields[f.focus].spec.kind == fieldGroups {
		controls = append(controls, controlHint{"↑/↓", "group"}, controlHint{"enter", "select/add"})
	}
	return controls
}

func (f form) expandedContent() bool {
	for _, field := range f.fields {
		if field.spec.kind == fieldMultiline && field.spec.rows > 5 {
			return true
		}
	}
	return false
}

func contentField(placeholder string) fieldSpec {
	return fieldSpec{
		id:          "content",
		label:       "Content",
		kind:        fieldMultiline,
		placeholder: placeholder,
		rows:        10,
	}
}

func largeContentField(placeholder string) fieldSpec {
	field := contentField(placeholder)
	field.rows = 20
	return field
}
