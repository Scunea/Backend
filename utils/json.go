/*
This file is part of Scunea.
Copyright (C) 2025  Mateo Hideg

Scunea is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

Scunea is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with Scunea.  If not, see <https://www.gnu.org/licenses/>.
*/

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
