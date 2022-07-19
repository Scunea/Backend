package utils

import "encoding/json"

func ParseStringArray(jsonString string) []string {
	var arr []string
	json.Unmarshal([]byte(jsonString), &arr)
	return arr
}

func UnParse(jsonString interface{}) string {
	unparsed, _ := json.Marshal(jsonString)
	return string(unparsed)
}

func ParseStringString(jsonString string) map[string]string {
	var obj map[string]string
	json.Unmarshal([]byte(jsonString), &obj)
	return obj
}
