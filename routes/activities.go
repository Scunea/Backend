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

func StartActivities(r *gin.Engine, db *sqlx.DB) {
	r.GET("/activities", AuthRequired(db, false), func(c *gin.Context) {
		var user = utils.GetMyUser(db, c)

		var activities = utils.GetActivities(db, c.GetHeader("school"))
		for i, activity := range activities {
			if utils.Type(db, user, c.GetHeader("school")) == "teacher" {
				if activity.Author != user.Id {
					activities = append(activities[:i], activities[i+1:]...)
				}
			} else if utils.Type(db, user, c.GetHeader("school")) != "administrator" {
				if !utils.ExistsFunc(utils.ParseStringArray(activity.Receiver), func(s string) bool {
					return s == user.Id
				}) {
					activities = append(activities[:i], activities[i+1:]...)
				}
			}
		}

		var parsedActivities = utils.ParseActivities(db, activities)

		if utils.Type(db, user, c.GetHeader("school")) == "teacher" {
			c.JSON(http.StatusOK, parsedActivities)
		} else {
			var parsedActivitiesAlone = utils.AlonifyActivities(parsedActivities, user.Id)
			c.JSON(http.StatusOK, parsedActivitiesAlone)
		}

	})

	r.POST("/activities", AuthRequired(db, false), func(c *gin.Context) {
		var body types.ActivityPostRequest
		err := c.BindJSON(&body)

		var schoolId = c.GetHeader("school")

		if err == nil {
			var user = utils.GetMyUser(db, c)

			var files = []types.IdAndName{}

			if body.Files != nil {
				files = body.Files
			}

			filesString, _ := json.Marshal(files)

			var activity = types.ActivityParsed{
				Id:          uuid.NewString(),
				Title:       body.Title,
				Description: body.Description,
				Files:       files,
				Type:        body.Type,
				Delivery:    body.Delivery,
				Author: types.IdAndName{
					Id:   user.Id,
					Name: user.Name,
				},
				Date:       time.Now().UnixMilli(),
				Expiration: body.Expiration,
				Receiver:   body.Receiver,
				Delivered:  map[string]types.Deliver{},
				Result:     map[string]string{},
				Viewed:     []string{},
				School:     schoolId,
			}

			if utils.Type(db, user, schoolId) == "teacher" {

				var users = utils.GetUsersFromSchool(db, schoolId)

				if len(utils.FilterFunc(activity.Receiver, func(receiver string) bool {
					return utils.ExistsFunc(users, func(user types.User) bool {
						return receiver == user.Id
					})
				})) < len(activity.Receiver) {
					c.JSON(http.StatusNotFound, gin.H{
						"error": "Receiver not found",
					})
					return
				}

				if len(activity.Title) > 0 && len(activity.Receiver) > 0 && utils.FilesExist(activity.Files) {
					_, err := db.Exec("INSERT INTO activities (id, title, description, files, type, delivery, author, date, expiration, receiver, delivered, result, viewed, school) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)", activity.Id, activity.Title, activity.Description, string(filesString), activity.Type, activity.Delivery, user.Id, activity.Date, activity.Expiration, utils.UnParse(activity.Receiver), utils.UnParse(activity.Delivered), utils.UnParse(activity.Result), utils.UnParse(activity.Viewed), activity.School)
					if err == nil {
						c.JSON(http.StatusCreated, activity)
						utils.SendWebsocket(activity.School, user.Id, types.NewActivityTeacher{
							Event:    "newActivity",
							Activity: activity,
						})

						for _, receiver := range activity.Receiver {
							var newActivity = types.NewActivity{
								Event:    "newActivity",
								Activity: utils.AlonifyActivities([]types.ActivityParsed{activity}, receiver)[0],
							}

							utils.SendNotification(db, receiver, newActivity)

							utils.SendWebsocket(activity.School, receiver, newActivity)
						}
					} else {
						print(err.Error())
						c.JSON(http.StatusInternalServerError, gin.H{
							"error": "Error creating activity",
						})
					}
				} else {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "Title or Receiver is empty.",
					})
				}
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "You are not authorized to create an activity.",
				})
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request.",
			})
		}
	})

	r.PATCH("/activities/:id", AuthRequired(db, false), func(c *gin.Context) {
		var body types.ActivityPatchRequest
		err := c.BindJSON(&body)

		var schoolId = c.GetHeader("school")

		if err == nil {
			var user = utils.GetMyUser(db, c)

			if utils.Type(db, user, schoolId) == "teacher" {

				var activity, ok = utils.GetActivity(db, c.Param("id"), schoolId)

				var activityParsed = utils.ParseActivities(db, []types.Activity{activity})[0]

				var oldReceiver = activityParsed.Receiver

				if ok {

					var users = utils.GetUsersFromSchool(db, schoolId)

					if body.Receiver != nil {

						if len(utils.FilterFunc(body.Receiver, func(receiver string) bool {
							return utils.ExistsFunc(users, func(user types.User) bool {
								return receiver == user.Id
							})
						})) < len(body.Receiver) {
							c.JSON(http.StatusNotFound, gin.H{
								"error": "Receiver not found",
							})
							return
						}
					}

					if len(body.Title) > 0 {
						activityParsed.Title = body.Title
					}

					if len(body.Description) > 0 {
						activityParsed.Description = body.Description
					}

					if len(body.Files) > 0 {
						activityParsed.Files = body.Files
					}

					filesString, _ := json.Marshal(activityParsed.Files)

					if len(body.Type) > 0 {
						activityParsed.Type = body.Type
					}

					if len(body.Delivery) > 0 {
						activityParsed.Delivery = body.Delivery
					}

					if len(fmt.Sprint(body.Expiration)) > 0 {
						activityParsed.Expiration = body.Expiration
					}

					if len(body.Receiver) > 0 {
						activityParsed.Receiver = body.Receiver
					}

					_, err := db.Exec("UPDATE activities SET title = $1, description = $2, files = $3, type = $4, delivery = $5, expiration = $6, receiver = $7 WHERE id = $8", activityParsed.Title, activityParsed.Description, string(filesString), activityParsed.Type, activityParsed.Delivery, activityParsed.Expiration, utils.UnParse(activityParsed.Receiver), activityParsed.Id)
					if err == nil {
						c.JSON(http.StatusOK, activityParsed)
						for _, receiver := range activityParsed.Receiver {
							utils.SendWebsocket(activityParsed.School, receiver, types.EditedActivity{
								Event:    "editedActivity",
								Id:       activityParsed.Id,
								Activity: utils.AlonifyActivities([]types.ActivityParsed{activityParsed}, receiver)[0],
							})
						}

						for _, oldReceiver := range utils.FilterFunc(oldReceiver, func(receiver string) bool {
							return !utils.ExistsFunc(activityParsed.Receiver, func(r string) bool {
								return r == receiver
							})
						}) {
							utils.SendWebsocket(activityParsed.School, oldReceiver, types.DeletedActivity{
								Event: "deletedActivity",
								Id:    activityParsed.Id,
							})
						}
					} else {
						print(err.Error())
						c.JSON(http.StatusInternalServerError, gin.H{
							"error": "Error updating activity",
						})
					}
				} else {
					c.JSON(http.StatusNotFound, gin.H{
						"error": "Activity not found.",
					})
				}
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "You are not authorized to create an activity.",
				})
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request.",
			})
		}
	})

	r.DELETE("/activities/:id", AuthRequired(db, false), func(c *gin.Context) {
		var user = utils.GetMyUser(db, c)
		var activity, ok = utils.GetActivity(db, c.Param("id"), c.GetHeader("school"))
		if ok {
			if utils.Type(db, user, c.GetHeader("school")) == "administrator" || activity.Author == user.Id {
				_, err := db.Exec("DELETE FROM activities WHERE id = $1", c.Param("id"))
				if err == nil {
					c.JSON(http.StatusOK, gin.H{
						"success": "Activity deleted.",
					})
					for _, receiver := range utils.ParseStringArray(activity.Receiver) {
						utils.SendWebsocket(activity.School, receiver, types.DeletedActivity{
							Event: "deletedActivity",
							Id:    activity.Id,
						})
					}
				} else {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": "Error deleting activity.",
					})
				}
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "You are not authorized to delete this activity.",
				})
			}
		} else {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Activity not found.",
			})
		}
	})

	r.POST("/activities/view/:id", AuthRequired(db, false), func(c *gin.Context) {
		var user = utils.GetMyUser(db, c)
		var activity, ok = utils.GetActivity(db, c.Param("id"), c.GetHeader("school"))
		if ok {
			var viewed = utils.ParseStringArray(activity.Viewed)
			if utils.Type(db, user, c.GetHeader("school")) == "student" && utils.ExistsFunc(utils.ParseStringArray(activity.Receiver), func(v string) bool {
				return v == user.Id
			}) && !utils.ExistsFunc(viewed, func(v string) bool {
				return v == user.Id
			}) {
				viewed = append(viewed, user.Id)
				_, err := db.Exec("UPDATE activities SET viewed = $1 WHERE id = $2", utils.UnParse(viewed), activity.Id)
				if err == nil {
					c.JSON(http.StatusOK, gin.H{
						"success": "Activity viewed.",
					})
					utils.SendWebsocket(activity.School, user.Id, types.ViewedActivity{
						Event: "viewedActivity",
						Id:    activity.Id,
					})
					utils.SendWebsocket(activity.School, activity.Author, types.ViewedActivityTeacher{
						Event: "viewedActivity",
						Id:    activity.Id,
						User:  user.Id,
					})
				} else {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": "Error viewing activity.",
					})
				}
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "You are not authorized to view this activity.",
				})
			}
		} else {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Activity not found.",
			})
		}
	})

	r.POST("/activities/deliver/:id", AuthRequired(db, false), func(c *gin.Context) {
		var body types.ActivityDeliverRequest
		err := c.BindJSON(&body)

		if err == nil {
			if utils.FilesExist(body.Files) {
				var user = utils.GetMyUser(db, c)
				var activity, ok = utils.GetActivity(db, c.Param("id"), c.GetHeader("school"))
				if ok {
					expiration, err := strconv.ParseInt(activity.Expiration, 0, 64)

					if err == nil {

						if expiration == 0 || time.Now().UnixMilli() > expiration {

							var delivered map[string]types.Deliver
							json.Unmarshal([]byte(activity.Delivered), &delivered)

							var available = true

							for key := range delivered {
								if key == user.Id {
									available = false
									break
								}
							}

							if utils.Type(db, user, c.GetHeader("school")) == "student" && utils.ExistsFunc(utils.ParseStringArray(activity.Receiver), func(v string) bool {
								return v == user.Id
							}) && utils.ExistsFunc(utils.ParseStringArray(activity.Viewed), func(v string) bool {
								return v == user.Id
							}) && available {
								delivered[user.Id] = types.Deliver{
									Comments: body.Comments,
									Files:    body.Files,
									Date:     time.Now().UnixMilli(),
								}
								var result map[string]string
								json.Unmarshal([]byte(activity.Result), &result)
								result[user.Id] = "Unchecked"
								_, err := db.Exec("UPDATE activities SET delivered = $1, result = $2 WHERE id = $3", utils.UnParse(delivered), utils.UnParse(result), activity.Id)
								if err == nil {
									c.JSON(http.StatusOK, gin.H{
										"success": "Activity delivered.",
									})
									utils.SendWebsocket(activity.School, user.Id, types.DeliveredActivity{
										Event:    "deliveredActivity",
										Id:       activity.Id,
										Delivery: activity.Delivery,
									})
									utils.SendWebsocket(activity.School, activity.Author, types.DeliveredActivityTeacher{
										Event:    "deliveredActivity",
										Id:       activity.Id,
										User:     user.Id,
										Delivery: activity.Delivery,
									})
								} else {
									c.JSON(http.StatusInternalServerError, gin.H{
										"error": "Error delivering activity.",
									})
								}
							} else {
								c.JSON(http.StatusUnauthorized, gin.H{
									"error": "You are not authorized to deliver this activity.",
								})
							}
						} else {
							c.JSON(http.StatusBadRequest, gin.H{
								"error": "Activity expired.",
							})
						}
					} else {
						c.JSON(http.StatusNotFound, gin.H{
							"error": "Activity not found.",
						})
					}
				} else {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "Error delivering activity.",
					})
				}
			} else {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "Files not found.",
				})
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Error delivering activity.",
			})
		}
	})

	r.POST("/activities/result/:activityId/:userId", AuthRequired(db, false), func(c *gin.Context) {
		var body types.ActivityResultRequest
		err := c.BindJSON(&body)

		if err == nil {
			if body.Result == "Accepted" || body.Result == "Rejected" {
				var myUser = utils.GetMyUser(db, c)
				var user, ok = utils.GetUser(db, c.Param("userId"), c.GetHeader("school"))
				if ok {
					var activity, ok = utils.GetActivity(db, c.Param("activityId"), c.GetHeader("school"))
					if ok {
						if activity.Author == myUser.Id {
							var result map[string]string
							json.Unmarshal([]byte(activity.Result), &result)
							if result[user.Id] == "Unchecked" {
								result[user.Id] = body.Result
								_, err := db.Exec("UPDATE activities SET result = $1 WHERE id = $2", utils.UnParse(result), activity.Id)
								if err == nil {
									c.JSON(http.StatusOK, gin.H{
										"success": "Activity result updated.",
									})
									utils.SendWebsocket(activity.School, user.Id, types.ResultActivity{
										Event:  "resultActivity",
										Id:     activity.Id,
										Result: body.Result,
									})

									utils.SendWebsocket(activity.School, myUser.Id, types.ResultActivityTeacher{
										Event:  "resultActivity",
										Id:     activity.Id,
										User:   user.Id,
										Result: body.Result,
									})
								} else {
									c.JSON(http.StatusInternalServerError, gin.H{
										"error": "Error updating activity result.",
									})
								}
							} else {
								c.JSON(http.StatusBadRequest, gin.H{
									"error": "You are not authorized to update this activity result.",
								})
							}
						} else {
							c.JSON(http.StatusUnauthorized, gin.H{
								"error": "You are not authorized to update this activity result.",
							})
						}
					} else {
						c.JSON(http.StatusNotFound, gin.H{
							"error": "Activity not found.",
						})
					}
				} else {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "Error updating activity result.",
					})
				}
			} else {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Result must be Accepted or Rejected.",
				})
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Error updating activity result.",
			})
		}
	})
}
