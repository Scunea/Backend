package routes

import (
	"net/http"

	"github.com/Scunea/Backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func Start(r *gin.Engine, db *sqlx.DB) {
	go StartAccount(r, db)
	go StartActivities(r, db)
	go StartGrades(r, db)
	go StartInfo(r, db)
	go StartLogin(r, db)
	go StartMessages(r, db)
	go StartNotifications(r, db)
	go StartParents(r, db)
	go StartPeople(r, db)
	go StartReports(r, db)
	go StartSchool(r, db)
	go StartUpload(r, db)
	go StartWebsocket(r, db)
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
