package types

type Grade struct {
	Deliberation      string `json:"deliberation"`
	Conceptual        string `json:"conceptual"`
	AverageFirstFour  string `json:"averageFirstFour"`
	AverageSecondFour string `json:"averageSecondFour"`
	Final             string `json:"final"`
}

type GradeExtra struct {
	Id                string `json:"id"`
	FullName          string `json:"name"`
	Subject           string `json:"subject"`
	Deliberation      string `json:"deliberation"`
	Conceptual        string `json:"conceptual"`
	AverageFirstFour  string `json:"averageFirstFour"`
	AverageSecondFour string `json:"averageSecondFour"`
	Final             string `json:"final"`
}

type GradeExtraTwo struct {
	Subject           string `json:"subject"`
	Deliberation      string `json:"deliberation"`
	Conceptual        string `json:"conceptual"`
	AverageFirstFour  string `json:"averageFirstFour"`
	AverageSecondFour string `json:"averageSecondFour"`
	Final             string `json:"final"`
}

type UserParsed struct {
	Token          string                      `json:"token"`
	Id             string                      `json:"id"`
	Email          string                      `json:"email"`
	Verified       bool                        `json:"verified"`
	Verificator    string                      `json:"verificator"`
	Tfa            string                      `json:"tfa"`
	Name           string                      `json:"name"`
	Grades         map[string]map[string]Grade `json:"grades"`
	Password       string                      `json:"password"`
	Administrator  []string                    `json:"administrator"`
	Teacher        map[string]string           `json:"teacher"`
	Parents        []string                    `json:"parents"`
	PendingParents []string                    `json:"pendingparents"`
	Schools        []string                    `json:"schools"`
	PendingSchools []string                    `json:"pendingschools"`
}

type SimpleUser struct {
	Id       string   `json:"id"`
	Name     string   `json:"name"`
	Teacher  string   `json:"teacher"`
	Children []string `json:"children"`
	Type     string   `json:"type"`
}

type UserInfo struct {
	Id             string          `json:"id"`
	Name           string          `json:"name"`
	SchoolName     string          `json:"schoolName"`
	SchoolLogo     string          `json:"schoolLogo"`
	Email          string          `json:"email"`
	Tfa            bool            `json:"tfa"`
	Administrator  bool            `json:"administrator"`
	Teacher        string          `json:"teacher"`
	Parents        []string        `json:"parents"`
	PendingParents []string        `json:"pendingparents"`
	Children       []string        `json:"children"`
	Grades         []GradeExtraTwo `json:"grades"`
	Available      []SimpleUser    `json:"available"`
	Schools        []School        `json:"schools"`
	PendingSchools []School        `json:"pendingschools"`
}

type Person struct {
	Id       string      `json:"id"`
	Name     string      `json:"name"`
	Email    string      `json:"email"`
	Type     string      `json:"type"`
	Subject  string      `json:"subject"`
	Children []IdAndName `json:"children"`
}
