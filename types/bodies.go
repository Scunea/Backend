package types

// Account

type PatchAccountRequest struct {
	Name     string
	Password string
}

type DeleteAccountRequest struct {
	Password string `json:"password"`
	Otp      string `json:"otp"`
}

type PostOtpRequest struct {
	Otp string
}

type DeleteTfaRequest struct {
	Password string `json:"password"`
	Otp      string `json:"otp"`
}

// Activities

type ActivityPostRequest struct {
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Files       []IdAndName `json:"files"`
	Type        string      `json:"type"`
	Delivery    string      `json:"delivery"`
	Expiration  int64       `json:"expiration"`
	Receiver    []string    `json:"receiver"`
}

type ActivityPatchRequest struct {
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Files       []IdAndName `json:"files"`
	Type        string      `json:"type"`
	Delivery    string      `json:"delivery"`
	Expiration  int64       `json:"expiration"`
	Receiver    []string    `json:"receiver"`
}

type ActivityDeliverRequest struct {
	Comments string      `json:"comments"`
	Files    []IdAndName `json:"files"`
}

type ActivityResultRequest struct {
	Result string `json:"result"`
}

// Grades

type PostGradesRequest struct {
	Id                string `json:"id"`
	FullName          string `json:"name"`
	Subject           string `json:"subject"`
	Deliberation      string `json:"deliberation"`
	Conceptual        string `json:"conceptual"`
	AverageFirstFour  string `json:"averageFirstFour"`
	AverageSecondFour string `json:"averageSecondFour"`
	Final             string `json:"final"`
}

// Login

type SignUpRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type VerifyRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Otp      string `json:"otp"`
}

type LoginByTokenRequest struct {
	Token string `json:"token"`
}

// Messages

type MessagePostRequest struct {
	Title    string      `json:"title"`
	Content  string      `json:"content"`
	Files    []IdAndName `json:"files"`
	Receiver []string    `json:"receiver"`
}

// Notifications

type Keys struct {
	P256dh string `json:"p256dh"`
	Auth   string `json:"auth"`
}

type NotificationPostRequest struct {
	Endpoint string `json:"endpoint"`
	Keys     Keys `json:"keys"`
}

// Parents

type ParentsPutRequest struct {
	Email string `json:"email"`
}

// People

type PeoplePutRequest struct {
	Email   string `json:"email"`
	Type    string `json:"type"`
	Subject string `json:"subject"`
}

type PeopleDeleteRequest struct {
	Tos []string `json:"tos"`
}

// Reports

type ReportPostRequest struct {
	Title string    `json:"title"`
	File  IdAndName `json:"file"`
}

type ReportPatchRequest struct {
	Title string `json:"title"`
}
