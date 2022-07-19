package routes

import (
	"encoding/json"
	"net/http"

	"github.com/Scunea/Backend/types"
	"github.com/Scunea/Backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func StartPeople(r *gin.Engine, db *sqlx.DB) {
	r.GET("/people", AuthRequired(db, false), func(c *gin.Context) {
		var schoolId = c.GetHeader("school")
		var user = utils.GetMyUser(db, c)

		if utils.Type(db, user, schoolId) == "administrator" {
			var users = utils.GetUsersFromSchool(db, schoolId)

			var peopleParsed []types.Person

			for _, user := range users {
				var children = []types.IdAndName{}

				for _, child := range utils.GetChildren(db, user) {
					children = append(children, types.IdAndName{
						Id:   child.Id,
						Name: child.Name,
					})
				}

				var person = types.Person{
					Id:       user.Id,
					Name:     user.Name,
					Email:    user.Email,
					Type:     utils.Type(db, user, schoolId),
					Subject:  utils.ParseStringString(user.Teacher)[schoolId],
					Children: children,
				}

				peopleParsed = append(peopleParsed, person)
			}

			c.JSON(http.StatusOK, peopleParsed)
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Not authorized.",
			})
		}
	})

	r.PUT("/people", AuthRequired(db, false), func(c *gin.Context) {
		var body types.PeoplePutRequest
		c.BindJSON(&body)

		var schoolId = c.GetHeader("school")
		var user = utils.GetMyUser(db, c)

		if utils.Type(db, user, schoolId) == "administrator" {
			var user, ok = utils.GetUser(db, body.Email, "email")
			if ok {
				if len(body.Type) > 0 {
					if body.Type == "administrator" || body.Type == "teacher" || body.Type == "student" {
						var pendingSchools = utils.ParseStringArray(user.PendingSchools)
						var admin = utils.ParseStringArray(user.Administrator)
						var teacher = utils.ParseStringString(user.Teacher)

						pendingSchools = append(pendingSchools, schoolId)
						user.PendingSchools = utils.UnParse(pendingSchools)

						if body.Type == "administrator" {
							admin = append(admin, schoolId)
							user.Administrator = utils.UnParse(admin)
						}
						if body.Type == "teacher" {
							if len(body.Subject) > 0 {
								teacher[schoolId] = body.Subject
								user.Teacher = utils.UnParse(teacher)
							} else {
								c.JSON(http.StatusBadRequest, gin.H{
									"error": "Teachers must have a subject.",
								})
							}
						}

						_, err := db.Exec("UPDATE users SET pendingschools = $1, administrator = $2, teacher = $3 WHERE id = $4", user.PendingSchools, user.Administrator, user.Teacher, user.Id)

						if err == nil {
							c.JSON(http.StatusOK, gin.H{
								"success": "User invited.",
							})
						} else {
							c.JSON(http.StatusInternalServerError, gin.H{
								"error": "Error inviting user.",
							})
						}
					} else {
						c.JSON(http.StatusBadRequest, gin.H{
							"error": "Invalid type.",
						})
					}
				} else {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "Type is empty.",
					})
				}
			} else {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "User not found.",
				})
			}
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Not authorized.",
			})
		}
	})

	r.DELETE("/people", AuthRequired(db, false), func(c *gin.Context) {
		var body types.PeopleDeleteRequest
		c.BindJSON(&body)

		var schoolId = c.GetHeader("school")
		var myUser = utils.GetMyUser(db, c)

		if utils.Type(db, myUser, schoolId) == "administrator" {
			c.JSON(http.StatusOK, gin.H{
				"success": "Users queued for deletion.",
			})

			for _, id := range body.Tos {
				var user, ok = utils.GetUser(db, id, "id")
				if ok {
					var schools = utils.ParseStringArray(user.Schools)
					var pendingSchools = utils.ParseStringArray(user.PendingSchools)
					var grades map[string]map[string]types.Grade
					var admin = utils.ParseStringArray(user.Administrator)
					var teacher = utils.ParseStringString(user.Teacher)

					json.Unmarshal([]byte(user.Grades), &grades)
					delete(grades, schoolId)
					user.Grades = utils.UnParse(grades)

					var indexSchools = utils.IndexFunc(schools, func(school string) bool {
						return school == schoolId
					})
					schools = append(schools[:indexSchools], schools[indexSchools+1:]...)
					user.Schools = utils.UnParse(schools)

					var indexPendingSchools = utils.IndexFunc(pendingSchools, func(school string) bool {
						return school == schoolId
					})
					pendingSchools = append(pendingSchools[:indexPendingSchools], pendingSchools[indexPendingSchools+1:]...)
					user.PendingSchools = utils.UnParse(pendingSchools)

					if utils.Type(db, user, schoolId) == "administrator" {
						var index = utils.IndexFunc(admin, func(school string) bool {
							return school == schoolId
						})
						admin = append(admin[:index], admin[index+1:]...)
						user.Administrator = utils.UnParse(admin)
					}

					if utils.Type(db, user, schoolId) == "teacher" {
						delete(teacher, schoolId)
						user.Teacher = utils.UnParse(teacher)
					}

					db.Exec("UPDATE users SET schools = $1, pendingschools = $2, grades = $3, administrator = $4, teacher = $5 WHERE id = $6", user.Schools, user.PendingSchools, user.Grades, user.Administrator, user.Teacher, user.Id)
				}
			}
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Not authorized.",
			})
		}
	})
}
