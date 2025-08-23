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

func ParseReports(db *sqlx.DB, reports []types.Report) []types.ReportParsed {
	var reportsParsed = []types.ReportParsed{}

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
