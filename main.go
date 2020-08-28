package main

import (
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/bramvdbogaerde/go-randomstring"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"

	_ "github.com/joho/godotenv/autoload"
)

var startTime = time.Now()

func setupRouter() *gin.Engine {
	router := gin.New()

	if os.Getenv("LOGGING") == "true" {
		router.Use(gin.Logger())
		f, _ := os.Create("gin.log")
		gin.DefaultWriter = io.MultiWriter(f)
	}

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
		c.SaveUploadedFile(file, os.Getenv("FOLDER")+filename)                      // ...and move the file to ./uploads/
		c.JSON(200, gin.H{"file": filename})
	})

	router.GET("/delete", func(c *gin.Context) {
		if c.Request.Header.Get("Authorization") != os.Getenv("SECRET") { // Make sure the token is correct
			c.JSON(401, gin.H{"message": "Invalid token!"})
			return
		}

		if len(c.Query("file")) < 1 { // Check for empty file query
			c.JSON(400, gin.H{"message": "Missing \"file\" query!"})
			return
		}

		if _, err := os.Stat(os.Getenv("FOLDER") + c.Query("file")); os.IsNotExist(err) { // Make sure file exists
			c.JSON(404, gin.H{"message": "File doesn't exist!"})
			return
		}

		os.Remove(os.Getenv("FOLDER") + c.Query("file")) // Delete file
		c.JSON(200, gin.H{"message": "Deleted successfully"})
	})

	router.GET("/stats", func(c *gin.Context) { // File count etc
		files, _ := ioutil.ReadDir(os.Getenv("FOLDER"))
		c.JSON(200, gin.H{"fileCount": len(files), "uptime": time.Since(startTime)})
	})

	router.Use(static.Serve("/", static.LocalFile(os.Getenv("FOLDER"), true))) // Static files in ./uploads/

	return router
}

func main() {
	// Create ./uploads if it doesn't already exist
	if _, err := os.Stat(os.Getenv("FOLDER")); os.IsNotExist(err) {
		os.Mkdir(os.Getenv("FOLDER"), 777)
	}

	// Start webserver
	router := setupRouter()
	router.Run(os.Getenv("PORT"))
}
