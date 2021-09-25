package main

import (
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"davidcralph.co.uk/gosharex/util"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	ratelimit "github.com/zcong1993/gin-ratelimit"

	_ "github.com/joho/godotenv/autoload"
)

var startTime = time.Now() // used for uptime tracking

func setupRouter() *gin.Engine {
	router := gin.New()

	if os.Getenv("LOGGING") == "true" {
		router.Use(gin.Logger())
		f, _ := os.Create("gin.log")
		gin.DefaultWriter = io.MultiWriter(f)
	}

	if os.Getenv("RATELIMIT") == "true" {
		duration, _ := strconv.ParseInt(os.Getenv("RATELIMIT_DURATION"), 10, 64)
		count, _ := strconv.ParseInt(os.Getenv("RATELIMIT_COUNT"), 10, 64)

		router.Use(ratelimit.New(ratelimit.Config{Duration: duration, RateLimit: count}))
	}

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello World"})
	})

	router.POST("/upload", func(c *gin.Context) {
		if c.Request.Header.Get("Authorization") != os.Getenv("SECRET") { // Make sure the token is correct
			c.JSON(401, gin.H{"message": "Invalid token"})
			return
		}

		file, _ := c.FormFile("file")

		sizelimit, _ := strconv.ParseInt(os.Getenv("SIZE_LIMIT"), 10, 64)
		if os.Getenv("SIZE_LIMIT_ENABLED") == "true" && file.Size > sizelimit { // size limit
			c.JSON(413, gin.H{"message": "File too large"})
			return
		}

		filename := util.RandomString(4) + "." + strings.Split(file.Filename, ".")[1] // Create random filename...
		c.SaveUploadedFile(file, os.Getenv("FOLDER")+filename)                        // ...and move the file to the uploads folder
		c.JSON(200, gin.H{"file": filename})
	})

	router.GET("/delete", func(c *gin.Context) {
		if c.Request.Header.Get("Authorization") != os.Getenv("SECRET") { // Make sure the token is correct
			c.JSON(401, gin.H{"message": "Invalid token"})
			return
		}

		if len(c.Query("file")) < 1 { // Check for empty file query
			c.JSON(400, gin.H{"message": "Missing \"file\" query"})
			return
		}

		if _, err := os.Stat(os.Getenv("FOLDER") + c.Query("file")); os.IsNotExist(err) { // Make sure file exists
			c.JSON(404, gin.H{"message": "File doesn't exist"})
			return
		}

		os.Remove(os.Getenv("FOLDER") + c.Query("file")) // Delete file
		c.JSON(200, gin.H{"message": "Deleted successfully"})
	})

	router.GET("/stats", func(c *gin.Context) { // File count etc
		files, _ := ioutil.ReadDir(os.Getenv("FOLDER"))
		c.JSON(200, gin.H{"fileCount": len(files), "uptime": time.Since(startTime)})
	})

	router.Use(static.Serve("/web", static.LocalFile(os.Getenv("FOLDER"), true))) // Static files in ./uploads/

	if os.Getenv("WEB") == "true" {
		router.LoadHTMLGlob("web/templates/*")

		router.Use(static.Serve("/", static.LocalFile("/web/*/*", true))) // Static files in ./uploads/

		router.GET("/web", func(c *gin.Context) {
			c.HTML(200, "index.njk", gin.H{})
		})

		router.GET("/web/stats", func(c *gin.Context) {
			files, _ := ioutil.ReadDir(os.Getenv("FOLDER"))
			c.HTML(200, "stats.njk", gin.H{"fileCount": len(files), "uptime": time.Since(startTime)})
		})
	}

	return router
}

func main() {
	// Create ./uploads if it doesn't already exist
	if _, err := os.Stat(os.Getenv("FOLDER")); os.IsNotExist(err) {
		os.Mkdir(os.Getenv("FOLDER"), 0777)
	}

	// Start webserver
	router := setupRouter()
	router.Run(":" + os.Getenv("PORT"))
}
