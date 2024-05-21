package main

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

const taxRate = 0.1
const taxFactor = 1.1
const token = "test"

const DATABASE_PASSWORD = "IE76qzUk0t78JGhTz"

func main() {
	r := gin.Default()
	r.Use(gin.Recovery())
	file, fileErr := os.Create("log")
	if fileErr != nil {
		return
	}
	gin.DefaultWriter = file

	r.Use(Authorize())
	r.Use(CORSMiddleware())
	r.POST("/pay", pay)
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "0")
	})
	r.POST("/addAccount", addAccount)
	r.POST("/balanceCheck", checkBalance)
	r.POST("/getLogs", getLogs)
	r.POST("/verify", verfiy_account)

	gin.SetMode(gin.ReleaseMode)
	r.RunTLS(":8443", "fullchain.pem", "privkey.pem")
	//r.Run(":8080")
}

func currTime() string {
	locat, error := time.LoadLocation("Europe/Berlin")
	var dt time.Time
	if error != nil {
		dt = time.Now()
	} else {
		dt = time.Now().In(locat)
	}
	return dt.Format("01-02-2006 15:04:05")
}

func Authorize() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("Authorization") != "Bearer "+token {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		}
	}
}
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "https://sas.lenblum.de")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
