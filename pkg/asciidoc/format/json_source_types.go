package format

type jsonSpan struct {
	Start jsonPosition `json:"start"`
	End   jsonPosition `json:"end"`
}

type jsonPosition struct {
	Offset int `json:"offset"`
	Line   int `json:"line"`
	Column int `json:"column"`
}
