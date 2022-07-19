package utils

import (
	"encoding/json"
	"strconv"

	"github.com/Scunea/Backend/types"
	"github.com/jmoiron/sqlx"
)

func ParseReports(db *sqlx.DB, reports []types.Report) []types.ReportParsed {
	var reportsParsed []types.ReportParsed

	for _, report := range reports {
		var author, ok = GetUser(db, report.Author, "id")

		if ok {
			date, err := strconv.ParseInt(report.Date, 0, 64)

			var file types.IdAndName
			json.Unmarshal([]byte(report.File), &file)

			if err == nil {
				var reportParsed = types.ReportParsed{
					Id:    report.Id,
					Title: report.Title,
					File:  file,
					Author: types.IdAndName{
						Id:   author.Id,
						Name: author.Name,
					},
					Date:   date,
					School: report.School,
				}
				reportsParsed = append(reportsParsed, reportParsed)
			}
		}
	}
	return reportsParsed
}
