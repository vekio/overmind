package format

type jsonBlockMetadata struct {
	Source         *jsonSpan           `json:"source"`
	Title          *jsonBlockTitle     `json:"title"`
	Anchor         *jsonAnchor         `json:"anchor"`
	AttributeLists []jsonAttributeList `json:"attributeLists"`
}

type jsonBlockTitle struct {
	Source      jsonSpan `json:"source"`
	TitleSource jsonSpan `json:"titleSource"`
	Text        string   `json:"text"`
}

type jsonAnchor struct {
	Source        jsonSpan  `json:"source"`
	IDSource      jsonSpan  `json:"idSource"`
	RefTextSource *jsonSpan `json:"refTextSource,omitempty"`
	ID            string    `json:"id"`
	RefText       string    `json:"refText,omitempty"`
}

type jsonAttributeList struct {
	Source  jsonSpan        `json:"source"`
	Entries []jsonAttribute `json:"entries"`
}

type jsonAttribute struct {
	Source jsonSpan `json:"source"`
	Value  string   `json:"value"`
}
