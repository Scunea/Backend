package routes

import (
	"net/http"
	"net/mail"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/Scunea/Backend/types"
	"github.com/Scunea/Backend/utils"
)

func StartLogin(r *gin.Engine, db *sqlx.DB) {

	r.POST("/signup", func(c *gin.Context) {
		var body types.SignUpRequest
		err := c.BindJSON(&body)

		if err == nil {
			if len(body.Email) > 0 && len(body.Name) > 0 && len(body.Password) > 0 {
				_, err := mail.ParseAddress(body.Email)
				if err == nil {
					var _, ok = utils.GetUser(db, body.Email, "email")
					if !ok {
						_, err := db.Exec("INSERT INTO users (token, id, email, verified, verificator, tfa, name, grades, password, administrator, teacher, parents, pendingparents, schools, pendingschools) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)", "NEW", uuid.NewString(), body.Email, false, uuid.NewString(), "", body.Name, "{}", utils.CreatePassword(body.Password), "[]", "{}", "[]", "[]", "[]", "[]")
						if err == nil {
							c.JSON(http.StatusOK, gin.H{
								"success": "User created.",
							})
						} else {
							c.JSON(http.StatusInternalServerError, gin.H{
								"error": "Error creating user.",
							})
						}
					} else {
						c.JSON(http.StatusBadRequest, gin.H{
							"error": "Email already exists.",
						})
					}
				}
			} else {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Email, name or password is empty.",
				})
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request.",
			})
		}
	})

	r.POST("/verify/:verificator", func(c *gin.Context) {
		var body types.VerifyRequest
		err := c.BindJSON(&body)

		if err == nil {
			verificator := c.Param("verificator")

			if len(body.Email) > 0 && len(body.Password) > 0 && len(verificator) > 0 {
				var user, ok = utils.GetUser(db, body.Email, "email")
				if ok {
					if user.Verificator == verificator {
						if user.Verified {
							c.JSON(http.StatusForbidden, gin.H{
								"error": "User already verified.",
							})
						} else {
							_, err := db.Exec("UPDATE users SET token = $1, verified = $2, verificator = $3 WHERE id = $4", utils.CreateToken(user.Id), true, "", user.Id)

							if err == nil {
								c.JSON(http.StatusOK, gin.H{
									"id":             user.Id,
									"email":          user.Email,
									"tfa":            len(user.Tfa) > 0,
									"name":           user.Name,
									"schools":        utils.GetSchoolsWithName(db, user.Schools),
									"pendingschools": utils.GetSchoolsWithName(db, user.PendingSchools),
									"token":          user.Token,
								})
							} else {
								c.JSON(http.StatusInternalServerError, gin.H{
									"error": "Error verifying user.",
								})
							}
						}
					} else {
						c.JSON(http.StatusBadRequest, gin.H{
							"error": "Invalid token.",
						})
					}
				} else {
					c.JSON(http.StatusNotFound, gin.H{
						"error": "User does not exist",
					})
				}
			} else {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Email or password is empty.",
				})
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request.",
			})
		}
	})

	r.POST("/login", func(c *gin.Context) {
		var body types.LoginRequest
		err := c.BindJSON(&body)

		if err == nil {
			if len(body.Email) > 0 && len(body.Password) > 0 {
				var user, ok = utils.GetUser(db, body.Email, "email")
				if ok {

					if len(user.Tfa) < 1 || (len(user.Tfa) > 0 && utils.CheckOtp(body.Otp, user)) {
						if utils.CheckPassword(body.Password, user.Password) {
							var token = user.Token
							if !utils.VerifyToken(user.Token, db) {
								token = utils.CreateToken(user.Id)
								_, err := db.Exec("UPDATE users SET token = $1 WHERE id = $2", token, user.Id)
								if err != nil {
									c.JSON(http.StatusInternalServerError, gin.H{
										"error": "Error updating token.",
									})
									return
								}
							}
							c.JSON(http.StatusOK, gin.H{
								"id":             user.Id,
								"email":          user.Email,
								"tfa":            len(user.Tfa) > 0,
								"name":           user.Name,
								"schools":        utils.GetSchoolsWithName(db, user.Schools),
								"pendingschools": utils.GetSchoolsWithName(db, user.PendingSchools),
								"token":          token,
							})
						} else {
							c.JSON(http.StatusUnauthorized, gin.H{
								"error": "Invalid password.",
							})
						}
					} else {
						c.JSON(http.StatusUnauthorized, gin.H{
							"error": "Invalid OTP.",
						})
					}
				} else {
					c.JSON(http.StatusNotFound, gin.H{
						"error": "User does not exist",
					})
				}
			} else {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Email, password or OTP is empty.",
				})
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request.",
			})
		}
	})

	r.POST("/loginbytoken", func(c *gin.Context) {
		var body types.LoginByTokenRequest
		err := c.BindJSON(&body)

		if err == nil {
			if len(body.Token) > 0 {
				var user, ok = utils.GetUser(db, body.Token, "token")
				if ok {
					c.JSON(http.StatusOK, gin.H{
						"id":             user.Id,
						"email":          user.Email,
						"tfa":            len(user.Tfa) > 0,
						"name":           user.Name,
						"schools":        utils.GetSchoolsWithName(db, user.Schools),
						"pendingschools": utils.GetSchoolsWithName(db, user.PendingSchools),
						"token":          user.Token,
					})
				} else {
					c.JSON(http.StatusNotFound, gin.H{
						"error": "User does not exist",
					})
				}
			} else {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Token is empty.",
				})
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request.",
			})
		}
	})
}
