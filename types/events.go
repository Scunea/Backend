package types

// Account

type EditedUserUser struct {
	Id       string   `json:"id"`
	Name     string   `json:"name"`
	Subject  string   `json:"subject"`
	Children []string `json:"children"`
	Type     string   `json:"type"`
}

type EditedUser struct {
	Event string         `json:"event"`
	User  EditedUserUser `json:"user"`
}

// Activities

type NewActivity struct {
	Event    string              `json:"event"`
	Activity ActivityParsedAlone `json:"activity"`
}

type NewActivityTeacher struct {
	Event    string         `json:"event"`
	Activity ActivityParsed `json:"activity"`
}

type EditedActivity struct {
	Event    string              `json:"event"`
	Id       string              `json:"id"`
	Activity ActivityParsedAlone `json:"activity"`
}

type DeletedActivity struct {
	Event string `json:"event"`
	Id    string `json:"id"`
}

type ViewedActivity struct {
	Event string `json:"event"`
	Id    string `json:"id"`
}

type ViewedActivityTeacher struct {
	Event string `json:"event"`
	Id    string `json:"id"`
	User  string `json:"user"`
}

type DeliveredActivity struct {
	Event    string `json:"event"`
	Id       string `json:"id"`
	Delivery string `json:"delivery"`
}

type DeliveredActivityTeacher struct {
	Event    string `json:"event"`
	Id       string `json:"id"`
	User     string `json:"user"`
	Delivery string `json:"delivery"`
}

type ResultActivity struct {
	Event  string `json:"event"`
	Id     string `json:"id"`
	Result string `json:"result"`
}

type ResultActivityTeacher struct {
	Event  string `json:"event"`
	Id     string `json:"id"`
	User   string `json:"user"`
	Result string `json:"result"`
}

// Grades

type NewGrades struct {
	Event  string          `json:"event"`
	Grades []GradeExtraTwo `json:"grades"`
}

// Messages

type NewMessage struct {
	Event   string        `json:"event"`
	Message MessageParsed `json:"message"`
}

type EditedMessage struct {
	Event   string        `json:"event"`
	Id      string        `json:"id"`
	Message MessageParsed `json:"message"`
}

type DeletedMessage struct {
	Event string `json:"event"`
	Id    string `json:"id"`
}

// Parents

type ParentInvited struct {
	Event  string    `json:"event"`
	Parent IdAndName `json:"parent"`
}

type ChildrenInvited struct {
	Event    string    `json:"event"`
	Children IdAndName `json:"children"`
}

type ParentInviteRemoved struct {
	Event string `json:"event"`
	Id    string `json:"id"`
}

type ChildrenInviteRemoved struct {
	Event string `json:"event"`
	Id    string `json:"id"`
}

// Reports

type NewReport struct {
	Event  string       `json:"event"`
	Report ReportParsed `json:"report"`
}

type EditedReport struct {
	Event  string       `json:"event"`
	Id     string       `json:"id"`
	Report ReportParsed `json:"report"`
}

type DeletedReport struct {
	Event string `json:"event"`
	Id    string `json:"id"`
}
