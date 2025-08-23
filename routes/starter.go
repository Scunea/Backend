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

package routes

import (
	"net/http"

	"github.com/Scunea/Backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func Start(r *gin.Engine, db *sqlx.DB) {
	r.Static("/static", "./files")

	StartAccount(r, db)
	StartActivities(r, db)
	StartGrades(r, db)
	StartInfo(r, db)
	StartLogin(r, db)
	StartMessages(r, db)
	StartNotifications(r, db)
	StartParents(r, db)
	StartPeople(r, db)
	StartReports(r, db)
	StartSchool(r, db)
	StartUpload(r, db)
	StartWebsocket(r, db)
}

func AuthRequired(db *sqlx.DB, skipSchool bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		school := c.GetHeader("school")
		if utils.VerifyToken(token, db) && (skipSchool || utils.VerifySchool(token, school, db)) {
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Not authorized.",
			})
		}
		c.Abort()
	}
}
