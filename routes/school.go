package routes

import (
	"net/http"

	"github.com/Scunea/Backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func StartSchool(r *gin.Engine, db *sqlx.DB) {
	r.PATCH("/school", AuthRequired(db, false), func(c *gin.Context) {
		type PatchSchoolRequest struct {
			Name string
			Logo string
		}

		type EditedSchool struct {
			Event string `json:"event"`
			Name  string `json:"name"`
			Logo  string `json:"logo"`
		}

		var body PatchSchoolRequest
		err := c.BindJSON(&body)

		if err == nil {
			var user = utils.GetMyUser(db, c)
			if utils.Type(db, user, c.GetHeader("school")) == "administrator" {
				if len(body.Name) > 0 || len(body.Logo) > 0 {
					school, _ := utils.GetSchool(db, c.GetHeader("school"))
					if len(body.Name) > 0 {
						school.Name = body.Name
					}
					if len(body.Logo) > 0 && utils.FileExists(body.Logo) {
						school.Logo = body.Logo
					}
					_, err := db.Exec("UPDATE schools SET name = $1, logo = $2 WHERE id = $3", school.Name, school.Logo, school.Id)
					if err == nil {
						c.JSON(http.StatusOK, gin.H{
							"success": "School updated.",
						})
						utils.SendWebsocket(school.Id, "", EditedSchool{
							Event: "editedSchool",
							Name:  body.Name,
							Logo:  body.Logo,
						})
					} else {
						c.JSON(http.StatusInternalServerError, gin.H{
							"error": "Error updating school.",
						})
					}
				} else {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "No data to update",
					})
				}
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "You are not authorized to edit this school",
				})
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request.",
			})
		}
	})

	r.POST("/create", AuthRequired(db, true), func(c *gin.Context) {
		type CreateSchoolRequest struct {
			Name string
		}

		var body CreateSchoolRequest
		err := c.BindJSON(&body)

		if err == nil {
			if len(body.Name) > 0 {
				_, err := db.Exec("INSERT INTO schools (id, name, logo) VALUES ($1, $2, $3)", uuid.NewString(), body.Name, "")
				if err == nil {
					c.JSON(http.StatusOK, gin.H{
						"success": "School created.",
					})
				} else {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": "Error creating school.",
					})
				}
			} else {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "No data to update",
				})
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request.",
			})
		}
	})

	r.POST("/join/:id", AuthRequired(db, true), func(c *gin.Context) {
		type DeletedUser struct {
			Event  string `json:"event"`
			UserId string `json:"userId"`
		}

		type NewUserUser struct {
			Id       string   `json:"id"`
			Name     string   `json:"name"`
			Email    string   `json:"email"`
			Subject  string   `json:"subject"`
			Children []string `json:"children"`
			Type     string   `json:"type"`
		}

		type NewUser struct {
			Event string      `json:"event"`
			User  NewUserUser `json:"user"`
		}

		schoolId := c.Param("id")

		var user = utils.GetMyUser(db, c)

		var pendingSchools = utils.ParseStringArray(user.PendingSchools)
		var schools = utils.ParseStringArray(user.Schools)
		if utils.ExistsFunc(pendingSchools, func(school string) bool {
			return school == schoolId
		}) {
			pendingSchools = utils.Remove(pendingSchools, utils.IndexFunc(pendingSchools, func(school string) bool {
				return school == schoolId
			}))
			schools = append(schools, schoolId)
			_, err := db.Exec("UPDATE users SET pendingschools = $1, schools = $2 WHERE id = $3", utils.UnParse(pendingSchools), utils.UnParse(schools), user.Id)
			if err == nil {
				c.JSON(http.StatusOK, gin.H{
					"success": "You have joined the school.",
				})
				var users = utils.GetUsers(db)
				for _, userToSend := range users {
					if utils.Type(db, user, schoolId) == "teacher" || utils.Type(db, userToSend, schoolId) == "administrator" {
						utils.SendWebsocket("", userToSend.Id, DeletedUser{
							Event:  "deletedUser",
							UserId: user.Id,
						})
						var childrenIds []string
						for _, child := range utils.GetChildren(db, user) {
							childrenIds = append(childrenIds, child.Id)
						}
						utils.SendWebsocket("", userToSend.Id, NewUser{
							Event: "newUser",
							User: NewUserUser{
								Id:       user.Id,
								Name:     user.Name,
								Email:    user.Email,
								Subject:  utils.ParseStringString(user.Teacher)[schoolId],
								Children: childrenIds,
								Type:     utils.Type(db, user, schoolId),
							},
						})
					}
				}
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Error joining school.",
				})
			}
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "You are not authorized to join this school",
			})
		}
	})
}
