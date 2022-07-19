package routes

import (
	"github.com/Scunea/Backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func StartWebsocket(r *gin.Engine, db *sqlx.DB) {
	r.GET("/socket", func(c *gin.Context) {
		utils.HandleWebsocket(c.Writer, c.Request, db)
	})
}
