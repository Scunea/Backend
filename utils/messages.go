package utils

import (
	"encoding/json"
	"strconv"

	"github.com/Scunea/Backend/types"
	"github.com/jmoiron/sqlx"
)

func ParseMessages(db *sqlx.DB, messages []types.Message) []types.MessageParsed {
	var messagesParsed []types.MessageParsed

	for _, message := range messages {
		var author, ok = GetUser(db, message.Author, "id")

		if ok {
			date, err := strconv.ParseInt(message.Date, 0, 64)

			var files []types.IdAndName
			json.Unmarshal([]byte(message.Files), &files)

			var content = message.Content

			var pdf struct {
				Pdf string `json:"pdf"`
			}

			errPdf := json.Unmarshal([]byte(content), &pdf)

			if errPdf == nil {
				content = ""
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
