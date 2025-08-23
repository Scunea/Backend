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

import (
	"encoding/json"
	"strconv"

	"github.com/Scunea/Backend/types"
	"github.com/jmoiron/sqlx"
)

func ParseMessages(db *sqlx.DB, messages []types.Message) []types.MessageParsed {
	var messagesParsed = []types.MessageParsed{}

	for _, message := range messages {
		var author, ok = GetUser(db, message.Author, "id")

		if ok {
			date, err := strconv.ParseInt(message.Date, 0, 64)

			var files = []types.IdAndName{}
			json.Unmarshal([]byte(message.Files), &files)

			var content = message.Content

			var pdf struct {
				Pdf string `json:"pdf"`
			}

			errPdf := json.Unmarshal([]byte(content), &pdf)

			if errPdf == nil {
				//content = ""
			}

			if err == nil {
				var messageParsed = types.MessageParsed{
					Id:      message.Id,
					Title:   message.Title,
					Content: content,
					Pdf:     pdf.Pdf,
					Files:   files,
					Author: types.IdAndName{
						Id:   author.Id,
						Name: author.Name,
					},
					Date:     date,
					Receiver: ParseStringArray(message.Receiver),
					School:   message.School,
				}
				messagesParsed = append(messagesParsed, messageParsed)
			}
		}
	}
	return messagesParsed
}
