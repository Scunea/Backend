package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Scunea/Backend/types"
	"github.com/Scunea/Backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func StartReports(r *gin.Engine, db *sqlx.DB) {
	r.GET("/reports", AuthRequired(db, false), func(c *gin.Context) {
		var schoolId = c.GetHeader("school")
		var reports = utils.GetReports(db, schoolId)

		var reportsParsed = utils.ParseReports(db, reports)

		c.JSON(http.StatusOK, reportsParsed)
	})

	r.POST("/reports", AuthRequired(db, false), func(c *gin.Context) {
		var body types.ReportPostRequest
		c.BindJSON(&body)

		var schoolId = c.GetHeader("school")
		var user = utils.GetMyUser(db, c)

		if utils.Type(db, user, schoolId) == "administrator" {
			if len(body.Title) > 0 && len(body.File.Id) > 0 {

				if utils.FilesExist([]types.IdAndName{body.File}) {

					var report = types.ReportParsed{
						Id:    uuid.NewString(),
						Title: body.Title,
						File:  body.File,
						Author: types.IdAndName{
							Id:   user.Id,
							Name: user.Name,
						},
						Date:   time.Now().UnixMilli(),
						School: schoolId,
					}

					_, err := db.Exec("INSERT INTO reports (id, title, file, author, date, school) VALUES ($1, $2, $3, $4, $5, $6)", report.Id, report.Title, utils.UnParse(report.File), report.Author.Id, fmt.Sprint(report.Date), report.School)
					if err == nil {
						c.JSON(http.StatusOK, gin.H{
							"success": "Report sent.",
						})

						var users = utils.GetUsersFromSchool(db, schoolId)

						var newReport = types.NewReport{
							Event:  "newReport",
							Report: report,
						}

						for _, user := range users {
							utils.SendNotification(db, user.Id, newReport)
						}

						utils.SendWebsocket(schoolId, "", newReport)
					} else {
						c.JSON(http.StatusBadRequest, gin.H{
							"error": "Error sending report.",
						})
					}
				} else {
					c.JSON(http.StatusNotFound, gin.H{
						"error": "File not found.",
					})
				}
			} else {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Title or File is empty.",
				})
			}
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "You are not an administrator.",
			})
		}
	})

	r.PATCH("/reports/:id", AuthRequired(db, false), func(c *gin.Context) {
		var reportId = c.Param("id")

		var body types.ReportPatchRequest
		c.BindJSON(&body)

		var schoolId = c.GetHeader("school")
		var user = utils.GetMyUser(db, c)

		var report, ok = utils.GetReport(db, reportId, schoolId)

		if ok {
			if utils.Type(db, user, schoolId) == "administrator" {
				var file types.IdAndName

				json.Unmarshal([]byte(report.File), &file)

				date, err := strconv.ParseInt(report.Date, 0, 64)

				if err == nil {

					var report = types.ReportParsed{
						Id:    report.Id,
						Title: body.Title,
						File:  file,
						Author: types.IdAndName{
							Id:   user.Id,
							Name: user.Name,
						},
						Date:   date,
						School: schoolId,
					}

					_, err := db.Exec("UPDATE reports SET title = $1 WHERE id = $2", report.Title, report.Id)
					if err == nil {
						c.JSON(http.StatusOK, gin.H{
							"success": "Report edited.",
						})
						utils.SendWebsocket(schoolId, "", types.EditedReport{
							Event:  "editedReport",
							Id:     report.Id,
							Report: report,
						})
					} else {
						c.JSON(http.StatusBadRequest, gin.H{
							"error": "Error editing report.",
						})
					}
				} else {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "Error editing report.",
					})
				}
			} else {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "You are not the author of this report.",
				})
			}
		} else {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Report not found.",
			})
		}
	})

	r.DELETE("/reports/:id", AuthRequired(db, false), func(c *gin.Context) {
		var reportId = c.Param("id")

		var schoolId = c.GetHeader("school")
		var user = utils.GetMyUser(db, c)

		var _, ok = utils.GetReport(db, reportId, schoolId)

		if ok {
			if utils.Type(db, user, schoolId) == "administrator" {
				_, err := db.Exec("DELETE FROM reports WHERE id = $1", reportId)
				if err == nil {
					c.JSON(http.StatusOK, gin.H{
						"success": "Report deleted.",
					})
					utils.SendWebsocket(schoolId, "", types.DeletedReport{
						Event: "deletedReport",
						Id:    reportId,
					})
				} else {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "Error deleting report.",
					})
				}
			} else {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "You are not an administrator.",
				})
			}
		} else {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Report not found.",
			})
		}
	})
}
