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

func StartMessages(r *gin.Engine, db *sqlx.DB) {
	r.GET("/messages", AuthRequired(db, false), func(c *gin.Context) {
		var schoolId = c.GetHeader("school")
		var user = utils.GetMyUser(db, c)
		var messages = utils.GetMessages(db, schoolId)

		if utils.Type(db, user, schoolId) != "administrator" {
			for i, message := range messages {
				if !utils.ExistsFunc(utils.ParseStringArray(message.Receiver), func(receiver string) bool {
					return receiver == user.Id
				}) {
					messages = append(messages[:i], messages[i+1:]...)
				}
			}
		}

		var messagesParsed = utils.ParseMessages(db, messages)

		c.JSON(http.StatusOK, messagesParsed)
	})

	r.POST("/messages", AuthRequired(db, false), func(c *gin.Context) {
		var body types.MessagePostRequest
		c.BindJSON(&body)

		var schoolId = c.GetHeader("school")
		var user = utils.GetMyUser(db, c)

		if len(body.Title) > 0 && len(body.Content) > 0 && len(body.Receiver) > 0 {

			var files = []types.IdAndName{}

			if body.Files != nil {
				files = body.Files
			}

			if utils.FilesExist(files) {

				var available = utils.GetAvailable(db, user, schoolId)

				if utils.ExistsFunc(body.Receiver, func(receiver string) bool {
					return !utils.ExistsFunc(available, func(available types.SimpleUser) bool {
						return available.Id == receiver
					})
				}) {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "One or more receivers are not available.",
					})
					return
				}

				var message = types.MessageParsed{
					Id:       uuid.NewString(),
					Title:    body.Title,
					Content:  body.Content,
					Files:    files,
					Receiver: body.Receiver,
					Author: types.IdAndName{
						Id:   user.Id,
						Name: user.Name,
					},
					Date:   time.Now().UnixMilli(),
					School: schoolId,
				}

				_, err := db.Exec("INSERT INTO messages (id, title, content, files, receiver, author, date, school) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", message.Id, message.Title, message.Content, utils.UnParse(message.Files), utils.UnParse(message.Receiver), message.Author.Id, fmt.Sprint(message.Date), message.School)
				if err == nil {
					c.JSON(http.StatusOK, gin.H{
						"success": "Message sent.",
					})

					var newMessage = types.NewMessage{
						Event:   "newMessage",
						Message: message,
					}

					utils.SendWebsocket(message.School, user.Id, newMessage)

					for _, receiver := range message.Receiver {
						utils.SendNotification(db, receiver, newMessage)
						utils.SendWebsocket(message.School, receiver, newMessage)
					}
				} else {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "Error sending message.",
					})
				}
			} else {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "Files not found.",
				})
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Title, Content or Receiver is empty.",
			})
		}
	})

	r.PATCH("/messages/:id", AuthRequired(db, false), func(c *gin.Context) {
		var messageId = c.Param("id")

		var body types.MessagePostRequest
		c.BindJSON(&body)

		var schoolId = c.GetHeader("school")
		var user = utils.GetMyUser(db, c)

		var message, ok = utils.GetMessage(db, messageId, schoolId)

		if ok {
			if utils.Type(db, user, schoolId) == "administrator" || message.Author == user.Id {
				var files []types.IdAndName
				json.Unmarshal([]byte(message.Files), &files)

				if body.Files != nil {
					files = body.Files
				}

				if utils.FilesExist(files) {

					var available = utils.GetAvailable(db, user, schoolId)

					if utils.ExistsFunc(body.Receiver, func(receiver string) bool {
						return !utils.ExistsFunc(available, func(available types.SimpleUser) bool {
							return available.Id == receiver
						})
					}) {
						c.JSON(http.StatusBadRequest, gin.H{
							"error": "One or more receivers are not available.",
						})
						return
					}

					date, err := strconv.ParseInt(message.Date, 0, 64)

					if err == nil {

						var message = types.MessageParsed{
							Id:       message.Id,
							Title:    message.Title,
							Content:  message.Content,
							Files:    files,
							Receiver: utils.ParseStringArray(message.Receiver),
							Author: types.IdAndName{
								Id:   user.Id,
								Name: user.Name,
							},
							Date:   date,
							School: schoolId,
						}

						if len(body.Title) > 0 {
							message.Title = body.Title
						}

						if len(body.Content) > 0 {
							message.Content = body.Content
						}

						var oldReceivers []string

						if len(body.Receiver) > 0 {
							oldReceivers = message.Receiver
							message.Receiver = body.Receiver
						}

						_, err := db.Exec("UPDATE messages SET title = $1, content = $2, files = $3, receiver = $4 WHERE id = $5", message.Title, message.Content, utils.UnParse(message.Files), utils.UnParse(message.Receiver), message.Id)
						if err == nil {
							c.JSON(http.StatusOK, gin.H{
								"success": "Message edited.",
							})
							utils.SendWebsocket(message.School, user.Id, types.EditedMessage{
								Event:   "editedMessage",
								Id:      message.Id,
								Message: message,
							})
							for _, receiver := range message.Receiver {
								utils.SendWebsocket(message.School, receiver, types.EditedMessage{
									Event:   "editedMessage",
									Id:      message.Id,
									Message: message,
								})
							}

							for _, oldReceiver := range utils.FilterFunc(oldReceivers, func(oldReceiver string) bool {
								return !utils.ExistsFunc(message.Receiver, func(receiver string) bool {
									return receiver == oldReceiver
								})
							}) {
								utils.SendWebsocket(message.School, oldReceiver, types.DeletedMessage{
									Event: "deletedMessage",
									Id:    message.Id,
								})
							}
						} else {
							c.JSON(http.StatusBadRequest, gin.H{
								"error": "Error editing message.",
							})
						}
					} else {
						c.JSON(http.StatusBadRequest, gin.H{
							"error": "Error editing message.",
						})
					}
				} else {
					c.JSON(http.StatusNotFound, gin.H{
						"error": "Files not found.",
					})
				}
			} else {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "You are not the author of this message.",
				})
			}
		} else {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Message not found.",
			})
		}
	})

	r.DELETE("/messages/:id", func(c *gin.Context) {
		var messageId = c.Param("id")
		var schoolId = c.GetHeader("school")
		var user = utils.GetMyUser(db, c)

		var message, ok = utils.GetMessage(db, messageId, schoolId)

		if ok {
			if utils.Type(db, user, schoolId) == "administrator" || message.Author == user.Id {
				_, err := db.Exec("DELETE FROM messages WHERE id = $1", message.Id)
				if err == nil {
					c.JSON(http.StatusOK, gin.H{
						"success": "Message deleted.",
					})
					utils.SendWebsocket(message.School, user.Id, types.DeletedMessage{
						Event: "deletedMessage",
						Id:    message.Id,
					})
					for _, receiver := range utils.ParseStringArray(message.Receiver) {
						utils.SendWebsocket(message.School, receiver, types.DeletedMessage{
							Event: "deletedMessage",
							Id:    message.Id,
						})
					}
				} else {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "Error deleting message.",
					})
				}
			} else {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "You are not the author of this message.",
				})
			}
		} else {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Message not found.",
			})
		}
	})
}
