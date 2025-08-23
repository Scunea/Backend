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
