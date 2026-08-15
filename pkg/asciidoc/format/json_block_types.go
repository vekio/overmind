package format

type jsonSection struct {
	Kind          string            `json:"kind"`
	Source        jsonSpan          `json:"source"`
	HeadingSource jsonSpan          `json:"headingSource"`
	ContentSource jsonSpan          `json:"contentSource"`
	TitleSource   jsonSpan          `json:"titleSource"`
	Level         int               `json:"level"`
	Title         string            `json:"title"`
	Metadata      jsonBlockMetadata `json:"metadata"`
	Blocks        []any             `json:"blocks"`
}

type jsonParagraph struct {
	Kind    string       `json:"kind"`
	Source  jsonSpan     `json:"source"`
	Text    string       `json:"text"`
	Inlines []jsonInline `json:"inlines"`
}

type jsonBreak struct {
	Kind   string   `json:"kind"`
	Source jsonSpan `json:"source"`
}

type jsonDelimitedBlock struct {
	Kind          string            `json:"kind"`
	Source        jsonSpan          `json:"source"`
	OpeningSource jsonSpan          `json:"openingSource"`
	ContentSource jsonSpan          `json:"contentSource"`
	ClosingSource *jsonSpan         `json:"closingSource,omitempty"`
	BlockKind     string            `json:"blockKind"`
	ContentModel  string            `json:"contentModel"`
	Marker        string            `json:"marker"`
	Metadata      jsonBlockMetadata `json:"metadata"`
	Content       string            `json:"content"`
	Closed        bool              `json:"closed"`
}

type jsonAdmonition struct {
	Kind           string            `json:"kind"`
	Source         jsonSpan          `json:"source"`
	ContentSource  jsonSpan          `json:"contentSource"`
	LabelSource    jsonSpan          `json:"labelSource"`
	AdmonitionKind string            `json:"admonitionKind"`
	Label          string            `json:"label"`
	Text           string            `json:"text"`
	Inlines        []jsonInline      `json:"inlines"`
	Metadata       jsonBlockMetadata `json:"metadata"`
}

type jsonBlockMacro struct {
	Kind             string            `json:"kind"`
	Source           jsonSpan          `json:"source"`
	NameSource       jsonSpan          `json:"nameSource"`
	TargetSource     jsonSpan          `json:"targetSource"`
	AttributesSource jsonSpan          `json:"attributesSource"`
	Name             string            `json:"name"`
	Target           string            `json:"target"`
	Attributes       string            `json:"attributes"`
	Metadata         jsonBlockMetadata `json:"metadata"`
}

type jsonAttributeEntry struct {
	Kind        string    `json:"kind"`
	Source      jsonSpan  `json:"source"`
	NameSource  jsonSpan  `json:"nameSource"`
	ValueSource *jsonSpan `json:"valueSource,omitempty"`
	Operation   string    `json:"operation"`
	Name        string    `json:"name"`
	Value       string    `json:"value"`
	Header      bool      `json:"header"`
}
