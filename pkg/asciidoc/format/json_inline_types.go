package format

type jsonInline struct {
	Kind          string        `json:"kind"`
	Source        jsonSpan      `json:"source"`
	ContentSource *jsonSpan     `json:"contentSource,omitempty"`
	Value         *string       `json:"value,omitempty"`
	Children      *[]jsonInline `json:"children,omitempty"`
	TargetSource  *jsonSpan     `json:"targetSource,omitempty"`
	LabelSource   *jsonSpan     `json:"labelSource,omitempty"`
	Target        *string       `json:"target,omitempty"`
}
