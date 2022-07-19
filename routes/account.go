package routes

import (
	"crypto/rand"
	"encoding/base32"
	"net/http"

	"github.com/Scunea/Backend/types"
	"github.com/Scunea/Backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/jltorresm/otpgo"
	"github.com/jltorresm/otpgo/config"
	"github.com/jmoiron/sqlx"
)

func StartAccount(r *gin.Engine, db *sqlx.DB) {

	r.PATCH("/account", AuthRequired(db, false), func(c *gin.Context) {
		var body types.PatchAccountRequest
		err := c.BindJSON(&body)

		if err == nil {
			if len(body.Name) > 0 && len(body.Password) > 0 {
				var user = utils.GetMyUser(db, c)
				if len(body.Name) > 0 {
					user.Name = body.Name
				}
				if len(body.Password) > 0 {
					user.Password = utils.CreatePassword(body.Password)
				}
				_, err := db.Exec("UPDATE users SET name = $1, password = $2 WHERE id = $3", user.Name, user.Password, user.Id)
				if err == nil {
					c.JSON(http.StatusOK, gin.H{
						"success": "Account updated.",
					})

					var users = utils.GetUsers(db)
					var schoolId = c.GetHeader("school")
					for _, userToSend := range users {
						if userToSend.Id == user.Id || utils.Type(db, userToSend, schoolId) == "administrator" {
							var childrenIds []string
							for _, child := range utils.GetChildren(db, user) {
								childrenIds = append(childrenIds, child.Id)
							}
							utils.SendWebsocket("", userToSend.Id, types.EditedUser{
								Event: "editedUser",
								User: types.EditedUserUser{
									Id:       user.Id,
									Name:     user.Name,
									Subject:  utils.ParseStringString(user.Teacher)[schoolId],
									Children: childrenIds,
									Type:     utils.Type(db, user, schoolId),
								},
							})
						}
					}
				} else {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": "Error updating account.",
					})
				}
			} else {
				c.JSON(http.StatusBadRequest, gin.H{
					"message": "No data to update",
				})
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request.",
			})
		}
	})

	r.DELETE("/account", AuthRequired(db, false), func(c *gin.Context) {
		var body types.DeleteAccountRequest
		err := c.BindJSON(&body)

		if err == nil {
			var user = utils.GetMyUser(db, c)

			if len(body.Password) > 0 && utils.CheckPassword(body.Password, user.Password) {
				if len(user.Tfa) > 1 && len(body.Otp) < 1 {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "OTP required.",
					})
					return
				}
				if !utils.CheckOtp(body.Otp, user) {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "Invalid OTP.",
					})
					return
				}
				_, err := db.Exec("DELETE FROM users WHERE id = $1", user.Id)
				if err == nil {
					c.JSON(http.StatusOK, gin.H{
						"success": "Account deleted.",
					})
				} else {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": "Error deleting account.",
					})
				}
			} else {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid password",
				})
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request.",
			})
		}
	})

	r.POST("/otp", AuthRequired(db, false), func(c *gin.Context) {
		var user = utils.GetMyUser(db, c)

		if len(user.Tfa) < 1 {
			var keyByte = make([]byte, 64)
			rand.Reader.Read(keyByte)

			var key = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(keyByte)

			otp := otpgo.TOTP{
				Algorithm: config.HmacSHA1,
				Key:       key,
			}
			keyUri := otp.KeyUri(user.Email, "Scunea")
			qr, err := keyUri.QRCode()
			if err == nil {
				c.JSON(http.StatusOK, gin.H{
					"secret": key,
					"qr":     qr,
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Error generating OTP.",
				})
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "2FA already enabled.",
			})
		}
	})

	r.POST("/otp/:secret", AuthRequired(db, false), func(c *gin.Context) {
		secret := c.Param("secret")

		var body types.PostOtpRequest

		err := c.BindJSON(&body)

		if err == nil {

			var user = utils.GetMyUser(db, c)

			if len(user.Tfa) < 1 {
				if utils.CheckOtp(body.Otp, user) {
					_, err := db.Exec("UPDATE users SET tfa = $1 WHERE id = $2", secret, user.Id)
					if err == nil {
						c.JSON(http.StatusOK, gin.H{
							"success": "2FA enabled.",
						})

						var schoolId = c.GetHeader("school")

						var childrenIds []string
						for _, child := range utils.GetChildren(db, user) {
							childrenIds = append(childrenIds, child.Id)
						}

						utils.SendWebsocket("", user.Id, types.EditedUser{
							Event: "editedUser",
							User: types.EditedUserUser{
								Id:       user.Id,
								Name:     user.Name,
								Subject:  utils.ParseStringString(user.Teacher)[schoolId],
								Children: childrenIds,
								Type:     utils.Type(db, user, schoolId),
							},
						})
					} else {
						c.JSON(http.StatusInternalServerError, gin.H{
							"error": "Error enabling 2FA.",
						})
					}
				} else {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "Invalid OTP.",
					})
				}
			} else {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "2FA already enabled.",
				})
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request.",
			})
		}
	})

	r.DELETE("/otp", AuthRequired(db, false), func(c *gin.Context) {
		var body types.DeleteTfaRequest
		err := c.BindJSON(&body)

		if err == nil {
			var user = utils.GetMyUser(db, c)

			if len(user.Tfa) > 0 {
				if len(body.Password) > 1 && utils.CheckPassword(body.Password, user.Password) && utils.CheckOtp(body.Otp, user) {
					_, err := db.Exec("UPDATE users SET tfa = '' WHERE id = $1", user.Id)
					if err == nil {
						c.JSON(http.StatusOK, gin.H{
							"success": "2FA disabled.",
						})
					} else {
						c.JSON(http.StatusInternalServerError, gin.H{
							"error": "Error disabling 2FA.",
						})
					}
				} else {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "Invalid password or OTP.",
					})
				}
			} else {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "2FA not enabled.",
				})
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request.",
			})
		}
	})
}
