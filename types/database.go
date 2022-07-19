package types

type User struct {
	Token          string `json:"token"`
	Id             string `json:"id"`
	Email          string `json:"email"`
	Verified       bool   `json:"verified"`
	Verificator    string `json:"verificator"`
	Tfa            string `json:"tfa"`
	Name           string `json:"name"`
	Grades         string `json:"grades"`
	Password       string `json:"password"`
	Administrator  string `json:"administrator"`
	Teacher        string `json:"teacher"`
	Parents        string `json:"parents"`
	PendingParents string `json:"pendingparents"`
	Schools        string `json:"schools"`
	PendingSchools string `json:"pendingschools"`
}

type School struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Logo string `json:"logo"`
}

type Activity struct {
	Id          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Files       string `json:"files"`
	Type        string `json:"type"`
	Delivery    string `json:"delivery"`
	Author      string `json:"author"`
	Date        string `json:"date"`
	Expiration  string `json:"expiration"`
	Receiver    string `json:"receiver"`
	Delivered   string `json:"delivered"`
	Result      string `json:"result"`
	Viewed      string `json:"viewed"`
	School      string `json:"school"`
}

type Message struct {
	Id       string `json:"id"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	Files    string `json:"files"`
	Author   string `json:"author"`
	Date     string `json:"date"`
	Receiver string `json:"receiver"`
	School   string `json:"school"`
}

type Notification struct {
	Endpoint string `json:"endpoint"`
	P256dh   string `json:"p256dh"`
	Auth     string `json:"auth"`
	Id       string `json:"id"`
}

type Report struct {
	Id     string `json:"id"`
	Title  string `json:"title"`
	File   string `json:"file"`
	Author string `json:"author"`
	Date   string `json:"date"`
	School string `json:"school"`
}
