package routes

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func StartUpload(r *gin.Engine, db *sqlx.DB) {
	r.POST("/upload", AuthRequired(db, false), func(c *gin.Context) {
		type File struct {
			Id  string `json:"id"`
			Url string `json:"url"`
		}

		form, _ := c.MultipartForm()
		files := form.File["upload"]

		if len(files) > 0 {
			var filesToSend []File
			for _, file := range files {
				realFile, err := file.Open()
				if err == nil {
					fileBytes := make([]byte, file.Size)
					realFile.Read(fileBytes)
					realFile.Close()
					if err == nil {
						var extension string
						var splitted = strings.Split(file.Filename, ".")
						if len(splitted) > 1 {
							extension = "." + splitted[len(splitted)-1]
						}
						var fileName = uuid.NewString() + extension
						os.WriteFile("./files/"+fileName, fileBytes, 0600)
						filesToSend = append(filesToSend, File{
							Id:  fileName,
							Url: c.Request.URL.Scheme + c.Request.Host + "/static/" + fileName,
						})
					} else {
						break
					}
				} else {
					break
				}
			}
			if c.GetHeader("simple") != "" {
				c.JSON(http.StatusOK, filesToSend[0])
			} else {
				c.JSON(http.StatusOK, filesToSend)
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "No file uploaded",
			})
		}
	})
}
