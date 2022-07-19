package types

type ReportParsed struct {
	Id     string    `json:"id"`
	Title  string    `json:"title"`
	File   IdAndName `json:"file"`
	Author IdAndName `json:"author"`
	Date   int64     `json:"date"`
	School string    `json:"school"`
}
