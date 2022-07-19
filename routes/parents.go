package routes

import (
	"net/http"

	"github.com/Scunea/Backend/types"
	"github.com/Scunea/Backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func StartParents(r *gin.Engine, db *sqlx.DB) {
	r.GET("/pendingchildren", AuthRequired(db, false), func(c *gin.Context) {
		var myUser = utils.GetMyUser(db, c)
		var users = utils.GetUsers(db)

		var pendingChildren = []types.IdAndName{}

		for _, user := range users {
			if utils.ExistsFunc(utils.ParseStringArray(user.Parents), func(parent string) bool {
				return parent == myUser.Id
			}) {
				pendingChildren = append(pendingChildren, types.IdAndName{
					Id:   user.Id,
					Name: user.Name,
				})
			}
		}

		c.JSON(http.StatusOK, pendingChildren)
	})

	r.GET("/pendingparents", AuthRequired(db, false), func(c *gin.Context) {
		var myUser = utils.GetMyUser(db, c)

		var pendingParents = []types.IdAndName{}

		for _, userId := range utils.ParseStringArray(myUser.Parents) {
			var user, ok = utils.GetUser(db, userId, "id")
			if ok {
				pendingParents = append(pendingParents, types.IdAndName{
					Id:   user.Id,
					Name: user.Name,
				})
			}
		}

		c.JSON(http.StatusOK, pendingParents)
	})

	r.PUT("/parents", AuthRequired(db, false), func(c *gin.Context) {
		var body types.ParentsPutRequest
		c.BindJSON(&body)

		if len(body.Email) > 0 {
			var user, ok = utils.GetUser(db, body.Email, "email")
			if ok {
				var myUser = utils.GetMyUser(db, c)
				var pendingParents = utils.ParseStringArray(myUser.PendingParents)
				var parents = utils.ParseStringArray(myUser.Parents)

				if !utils.ExistsFunc(pendingParents, func(parent string) bool {
					return parent == user.Id
				}) && !utils.ExistsFunc(parents, func(parent string) bool {
					return parent == user.Id
				}) {

					pendingParents = append(pendingParents, user.Id)

					_, err := db.Exec("UPDATE users SET pendingparents = $1 WHERE id = $2", utils.UnParse(pendingParents), myUser.Id)

					if err == nil {
						c.JSON(http.StatusOK, types.IdAndName{
							Id:   user.Id,
							Name: user.Name,
						})

						utils.SendWebsocket("", user.Id, types.ParentInvited{
							Event: "parentInvited",
							Parent: types.IdAndName{
								Id:   user.Id,
								Name: user.Name,
							},
						})

						utils.SendWebsocket("", user.Id, types.ChildrenInvited{
							Event: "childrenInvited",
							Children: types.IdAndName{
								Id:   myUser.Id,
								Name: myUser.Name,
							},
						})
					} else {
						c.JSON(http.StatusInternalServerError, gin.H{
							"error": "Failed to add parent.",
						})
					}
				} else {
					c.JSON(http.StatusNotFound, gin.H{
						"error": "User already your parent.",
					})
				}
			} else {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "User not found.",
				})
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Email is empty.",
			})
		}
	})

	r.DELETE("/parents/:id", AuthRequired(db, false), func(c *gin.Context) {
		var myUser = utils.GetMyUser(db, c)
		var userId = c.Param("id")

		var pendingParents = utils.ParseStringArray(myUser.PendingParents)

		if utils.ExistsFunc(pendingParents, func(parent string) bool {
			return parent == userId
		}) {
			var index = utils.IndexFunc(pendingParents, func(parent string) bool {
				return parent == userId
			})

			pendingParents = append(pendingParents[:index], pendingParents[index+1:]...)

			_, err := db.Exec("UPDATE users SET pendingparents = $1 WHERE id = $2", utils.UnParse(pendingParents), myUser.Id)

			if err == nil {
				c.JSON(http.StatusOK, gin.H{})

				utils.SendWebsocket("", userId, types.ParentInviteRemoved{
					Event: "parentInviteRemoved",
					Id:    userId,
				})

				utils.SendWebsocket("", userId, types.ChildrenInviteRemoved{
					Event: "childrenInviteRemoved",
					Id:    myUser.Id,
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to remove parent.",
				})
			}
		} else {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Parent not invited.",
			})
		}
	})

	r.POST("/accept/:id", AuthRequired(db, false), func(c *gin.Context) {
		var myUser = utils.GetMyUser(db, c)
		var child, ok = utils.GetUser(db, c.Param("id"), "id")

		if ok {
			var pendingParents = utils.ParseStringArray(child.PendingParents)

			if utils.ExistsFunc(pendingParents, func(parent string) bool {
				return parent == myUser.Id
			}) {
				var index = utils.IndexFunc(pendingParents, func(parent string) bool {
					return parent == myUser.Id
				})

				pendingParents = append(pendingParents[:index], pendingParents[index+1:]...)

				var parents = utils.ParseStringArray(child.Parents)

				parents = append(parents, myUser.Id)

				_, err := db.Exec("UPDATE users SET pendingparents = $1, parents = $2 WHERE id = $3", utils.UnParse(pendingParents), utils.UnParse(parents), child.Id)

				if err == nil {
					c.JSON(http.StatusOK, gin.H{
						"success": "Child accepted.",
					})

					utils.SendWebsocket("", child.Id, types.ParentInviteRemoved{
						Event: "parentInviteRemoved",
						Id:    myUser.Id,
					})

					utils.SendWebsocket("", myUser.Id, types.ChildrenInviteRemoved{
						Event: "childrenInviteRemoved",
						Id:    child.Id,
					})
				} else {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": "Failed to accept child.",
					})
				}
			} else {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "Parent not invited.",
				})
			}
		} else {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found.",
			})
		}
	})
}
