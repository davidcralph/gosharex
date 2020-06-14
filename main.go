package main

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/bramvdbogaerde/go-randomstring"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"

	_ "github.com/joho/godotenv/autoload"
)

func setupRouter() *gin.Engine {
	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "gotta go fast"})
	})

	router.POST("/upload", func(c *gin.Context) {
		if c.Request.Header.Get("Authorization") != os.Getenv("SECRET") { // Make sure the token is correct
			c.JSON(401, gin.H{"message": "Invalid token!"})
			return
		}
		file, _ := c.FormFile("file")
		filename := randomstring.New() + "." + strings.Split(file.Filename, ".")[1] // Create random filename...
		c.SaveUploadedFile(file, "./uploads/"+filename)                             // ...and move the file to ./uploads/
		c.JSON(200, gin.H{"file": filename})
	})

	router.GET("/stats", func(c *gin.Context) { // File count etc
		files, _ := ioutil.ReadDir("./uploads")
		c.JSON(200, gin.H{"fileCount": len(files)})
	})

	router.Use(static.Serve("/", static.LocalFile("./uploads", true))) // Static files in ./uploads/

	return router
}

func main() {
	// Create ./uploads if it doesn't already exist
	if _, err := os.Stat("./uploads/"); os.IsNotExist(err) {
		os.Mkdir("./uploads/", 777)
	}

	// Start webserver
	router := setupRouter()
	router.Run(os.Getenv("PORT"))
}
