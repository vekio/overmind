package format

type jsonTable struct {
	Kind          string            `json:"kind"`
	Source        jsonSpan          `json:"source"`
	OpeningSource jsonSpan          `json:"openingSource"`
	ContentSource jsonSpan          `json:"contentSource"`
	ClosingSource *jsonSpan         `json:"closingSource,omitempty"`
	Marker        string            `json:"marker"`
	Format        string            `json:"format"`
	Separator     string            `json:"separator"`
	Metadata      jsonBlockMetadata `json:"metadata"`
	Columns       []jsonTableColumn `json:"columns"`
	Rows          []jsonTableRow    `json:"rows"`
	Content       string            `json:"content"`
	Header        bool              `json:"header"`
	Closed        bool              `json:"closed"`
}

type jsonTableColumn struct {
	Spec string `json:"spec"`
}

type jsonTableRow struct {
	Kind   string          `json:"kind"`
	Source jsonSpan        `json:"source"`
	Header bool            `json:"header"`
	Cells  []jsonTableCell `json:"cells"`
}

type jsonTableCell struct {
	Kind          string       `json:"kind"`
	Source        jsonSpan     `json:"source"`
	MarkerSource  jsonSpan     `json:"markerSource"`
	ContentSource jsonSpan     `json:"contentSource"`
	Text          string       `json:"text"`
	Inlines       []jsonInline `json:"inlines"`
}
