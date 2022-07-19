package types

type Deliver struct {
	Comments string      `json:"comments"`
	Files    []IdAndName `json:"files"`
	Date     int64       `json:"date"`
}

type ActivityParsed struct {
	Id          string             `json:"id"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Files       []IdAndName        `json:"files"`
	Type        string             `json:"type"`
	Delivery    string             `json:"delivery"`
	Author      IdAndName          `json:"author"`
	Date        int64              `json:"date"`
	Expiration  int64              `json:"expiration"`
	Receiver    []string           `json:"receiver"`
	Delivered   map[string]Deliver `json:"delivered"`
	Result      map[string]string  `json:"result"`
	Viewed      []string           `json:"viewed"`
	School      string             `json:"school"`
}

type ActivityParsedAlone struct {
	Id          string      `json:"id"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Files       []IdAndName `json:"files"`
	Type        string      `json:"type"`
	Delivery    string      `json:"delivery"`
	Author      IdAndName   `json:"author"`
	Date        int64       `json:"date"`
	Expiration  int64       `json:"expiration"`
	Receiver    []string    `json:"receiver"`
	Delivered   Deliver     `json:"delivered"`
	Result      string      `json:"result"`
	Viewed      bool        `json:"viewed"`
	School      string      `json:"school"`
}
