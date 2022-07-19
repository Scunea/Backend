package types

type MessageParsed struct {
	Id       string      `json:"id"`
	Title    string      `json:"title"`
	Content  string      `json:"content"`
	Pdf      string      `json:"pdf"`
	Files    []IdAndName `json:"files"`
	Author   IdAndName   `json:"author"`
	Date     int64       `json:"date"`
	Receiver []string    `json:"receiver"`
	School   string      `json:"school"`
}
