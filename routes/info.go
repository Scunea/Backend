package routes

import (
	"encoding/json"
	"net/http"

	"github.com/Scunea/Backend/types"
	"github.com/Scunea/Backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func StartInfo(r *gin.Engine, db *sqlx.DB) {
	r.GET("/info", AuthRequired(db, false), func(c *gin.Context) {
		var schoolId = c.GetHeader("school")

		var myUser = utils.GetMyUser(db, c)
		var school, ok = utils.GetSchool(db, schoolId)

		var parsedChildren = utils.GetChildren(db, myUser)

		var children = []string{}

		for _, child := range parsedChildren {
			children = append(children, child.Id)
		}

		var grades map[string]map[string]types.Grade
		json.Unmarshal([]byte(myUser.Grades), &grades)

		var gradesParsed []types.GradeExtraTwo

		if utils.Type(db, myUser, school.Id) == "student" {
			for key, grade := range grades[school.Id] {
				var teacher, ok = utils.GetUser(db, key, "id")

				if ok {
					var gradeExtra = types.GradeExtraTwo{
						Subject:           utils.ParseStringString(teacher.Teacher)[school.Id],
						Deliberation:      grade.Deliberation,
						Conceptual:        grade.Conceptual,
						AverageFirstFour:  grade.AverageFirstFour,
						AverageSecondFour: grade.AverageSecondFour,
						Final:             grade.Final,
					}

					gradesParsed = append(gradesParsed, gradeExtra)
				}
			}
		} else {
			for _, child := range parsedChildren {
				var grades map[string]map[string]types.Grade
				json.Unmarshal([]byte(child.Grades), &grades)

				for key, grade := range grades[school.Id] {
					var teacher, ok = utils.GetUser(db, key, "id")

					if ok {
						var gradeExtra = types.GradeExtraTwo{
							Subject:           utils.ParseStringString(teacher.Teacher)[school.Id],
							Deliberation:      grade.Deliberation,
							Conceptual:        grade.Conceptual,
							AverageFirstFour:  grade.AverageFirstFour,
							AverageSecondFour: grade.AverageSecondFour,
							Final:             grade.Final,
						}

						gradesParsed = append(gradesParsed, gradeExtra)
					}
				}
			}
		}

		var available = utils.GetAvailable(db, myUser, school.Id)

		if ok {
			var userInfo = types.UserInfo{
				Id:             myUser.Id,
				Name:           myUser.Name,
				SchoolName:     school.Name,
				SchoolLogo:     school.Logo,
				Email:          myUser.Email,
				Tfa:            len(myUser.Tfa) > 0,
				Administrator:  utils.Type(db, myUser, c.GetHeader("school")) == "administrator",
				Teacher:        utils.ParseStringString(myUser.Teacher)[schoolId],
				Parents:        utils.ParseStringArray(myUser.Parents),
				PendingParents: utils.ParseStringArray(myUser.PendingParents),
				Children:       children,
				Grades:         gradesParsed,
				Available:      available,
				Schools:        utils.GetSchoolsWithName(db, myUser.Schools),
				PendingSchools: utils.GetSchoolsWithName(db, myUser.PendingSchools),
			}

			c.JSON(http.StatusOK, userInfo)
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Error getting info.",
			})
		}
	})
}
