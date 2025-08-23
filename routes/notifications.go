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

	"github.com/Scunea/Backend/types"
	"github.com/Scunea/Backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func StartNotifications(r *gin.Engine, db *sqlx.DB) {
	r.POST("/notifications", AuthRequired(db, false), func(c *gin.Context) {
		var body types.NotificationPostRequest
		c.BindJSON(&body)

		if len(body.Endpoint) > 0 && len(body.Keys.Auth) > 0 && len(body.Keys.P256dh) > 0 {
			var user = utils.GetMyUser(db, c)

			_, err := db.Exec("INSERT INTO notifications (endpoint, p256dh, auth, id) VALUES($1, $2, $3, $4) ON CONFLICT DO NOTHING", body.Endpoint, body.Keys.P256dh, body.Keys.Auth, user.Id)

			if err == nil {
				c.JSON(http.StatusOK, gin.H{
					"success": "Notifications registered.",
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Error registering notifications.",
				})
			}
		}
	})
}
