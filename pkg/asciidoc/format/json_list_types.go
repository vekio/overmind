package format

type jsonList struct {
	Kind      string            `json:"kind"`
	Source    jsonSpan          `json:"source"`
	ListKind  string            `json:"listKind,omitempty"`
	ListLevel int               `json:"listLevel"`
	Metadata  jsonBlockMetadata `json:"metadata"`
	Items     []jsonListItem    `json:"items"`
}

type jsonListItem struct {
	Kind               string        `json:"kind"`
	Source             jsonSpan      `json:"source"`
	MarkerSource       jsonSpan      `json:"markerSource"`
	PrincipalSource    *jsonSpan     `json:"principalSource,omitempty"`
	TermSource         *jsonSpan     `json:"termSource,omitempty"`
	DescriptionSource  *jsonSpan     `json:"descriptionSource,omitempty"`
	Marker             string        `json:"marker"`
	Principal          *string       `json:"principal,omitempty"`
	Term               *string       `json:"term,omitempty"`
	Description        *string       `json:"description,omitempty"`
	Inlines            *[]jsonInline `json:"inlines,omitempty"`
	TermInlines        *[]jsonInline `json:"termInlines,omitempty"`
	DescriptionInlines *[]jsonInline `json:"descriptionInlines,omitempty"`
	Blocks             []any         `json:"blocks"`
}
