package format

type jsonResult struct {
	Document    jsonDocument     `json:"document"`
	Diagnostics []jsonDiagnostic `json:"diagnostics"`
}

type jsonDocument struct {
	Kind   string             `json:"kind"`
	Source jsonSpan           `json:"source"`
	Title  *jsonDocumentTitle `json:"title"`
	Blocks []any              `json:"blocks"`
}

type jsonDocumentTitle struct {
	Kind        string   `json:"kind"`
	Source      jsonSpan `json:"source"`
	TitleSource jsonSpan `json:"titleSource"`
	Text        string   `json:"text"`
}

type jsonDiagnostic struct {
	Severity string   `json:"severity"`
	Message  string   `json:"message"`
	Source   jsonSpan `json:"source"`
}
