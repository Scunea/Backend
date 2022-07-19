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
