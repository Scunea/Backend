package routes

import (
	"encoding/json"
	"net/http"

	"github.com/Scunea/Backend/types"
	"github.com/Scunea/Backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func StartGrades(r *gin.Engine, db *sqlx.DB) {
	r.GET("/grades", AuthRequired(db, false), func(c *gin.Context) {
		var schoolId = c.GetHeader("school")
		var myUser = utils.GetMyUser(db, c)
		if utils.Type(db, myUser, schoolId) == "teacher" {
			var users = utils.GetUsersFromSchool(db, schoolId)
			var grades []types.GradeExtra

			for _, user := range users {
				if utils.Type(db, user, schoolId) == "student" {
					var grade map[string]map[string]types.Grade
					json.Unmarshal([]byte(user.Grades), &grade)

					var gradeExtra = types.GradeExtra{
						Id:                user.Id,
						FullName:          user.Name,
						Subject:           utils.ParseStringString(myUser.Teacher)[schoolId],
						Deliberation:      grade[schoolId][myUser.Id].Deliberation,
						Conceptual:        grade[schoolId][myUser.Id].Conceptual,
						AverageFirstFour:  grade[schoolId][myUser.Id].AverageFirstFour,
						AverageSecondFour: grade[schoolId][myUser.Id].AverageSecondFour,
						Final:             grade[schoolId][myUser.Id].Final,
					}

					grades = append(grades, gradeExtra)
				}
			}

			c.JSON(http.StatusOK, grades)
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "You are not authorized to view every grade.",
			})
		}
	})

	r.GET("/grades/:id", AuthRequired(db, false), func(c *gin.Context) {
		var schoolId = c.GetHeader("school")
		var myUser = utils.GetMyUser(db, c)
		if utils.Type(db, myUser, schoolId) == "administrator" {
			var user, ok = utils.GetUser(db, c.Param("id"), "id")
			if ok {
				if utils.Type(db, user, schoolId) == "student" && utils.ExistsFunc(utils.ParseStringArray(user.Schools), func(school string) bool {
					return school == schoolId
				}) {
					var grade map[string]map[string]types.Grade
					json.Unmarshal([]byte(user.Grades), &grade)

					var gradesExtra []types.GradeExtraTwo

					for key, grade := range grade[schoolId] {
						var teacher, ok = utils.GetUser(db, key, "id")

						if ok {
							var gradeExtra = types.GradeExtraTwo{
								Subject:           utils.ParseStringString(teacher.Teacher)[schoolId],
								Deliberation:      grade.Deliberation,
								Conceptual:        grade.Conceptual,
								AverageFirstFour:  grade.AverageFirstFour,
								AverageSecondFour: grade.AverageSecondFour,
								Final:             grade.Final,
							}
							gradesExtra = append(gradesExtra, gradeExtra)
						}
					}

					c.JSON(http.StatusOK, gradesExtra)
				} else {
					c.JSON(http.StatusUnauthorized, gin.H{
						"error": "Only students have grades.",
					})
				}
			} else {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "User not found.",
				})
			}
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "You are not authorized to access this user's grades.",
			})
		}
	})

	r.POST("/grades", AuthRequired(db, false), func(c *gin.Context) {
		var body []types.PostGradesRequest
		c.BindJSON(&body)

		var schoolId = c.GetHeader("school")
		var myUser = utils.GetMyUser(db, c)
		if utils.Type(db, myUser, schoolId) == "teacher" {
			c.JSON(http.StatusOK, gin.H{
				"success": "Grades queued for update.",
			})

			var users = utils.GetUsersFromSchool(db, schoolId)
			for _, user := range users {
				if utils.Type(db, user, schoolId) == "student" {
					var grade map[string]map[string]types.Grade
					json.Unmarshal([]byte(user.Grades), &grade)

					var newGrade = utils.FilterFunc(body, func(grade types.PostGradesRequest) bool {
						return grade.Id == user.Id
					})[0]

					grade[schoolId][myUser.Id] = types.Grade{
						Deliberation:      newGrade.Deliberation,
						Conceptual:        newGrade.Conceptual,
						AverageFirstFour:  newGrade.AverageFirstFour,
						AverageSecondFour: newGrade.AverageSecondFour,
						Final:             newGrade.Final,
					}

					db.Exec("UPDATE users SET grades = $1 WHERE id = $2", utils.UnParse(grade), user.Id)

					var newGrades = types.NewGrades{
						Event: "grades",
						Grades: []types.GradeExtraTwo{
							{
								Subject:           utils.ParseStringString(myUser.Teacher)[schoolId],
								Deliberation:      newGrade.Deliberation,
								Conceptual:        newGrade.Conceptual,
								AverageFirstFour:  newGrade.AverageFirstFour,
								AverageSecondFour: newGrade.AverageSecondFour,
								Final:             newGrade.Final,
							},
						},
					}

					utils.SendNotification(db, user.Id, newGrades)

					utils.SendWebsocket(schoolId, user.Id, newGrades)
				}
			}
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "You are not authorized to edit grades.",
			})
		}
	})
}
