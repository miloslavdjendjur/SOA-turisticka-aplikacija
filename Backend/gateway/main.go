package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200") // Dozvoli Angularu
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	stakeholdersTarget := "http://stakeholders:8081"
	toursTarget := "http://tours:8082"
	followersTarget := "http://followers:8083"

	// Rute za Stakeholders (Java)
	r.Any("/api/users/*path", reverseProxy(stakeholdersTarget))

	// Rute za Tours (Go)
	r.Any("/tours", reverseProxy(toursTarget))
	r.Any("/keypoints", reverseProxy(toursTarget))
	r.Any("/blogs", reverseProxy(toursTarget))
	r.Any("/tours/purchase", reverseProxy(toursTarget))
	r.Any("/position", reverseProxy(toursTarget))
	r.Any("/tours/start", reverseProxy(toursTarget))
	r.Any("/tours/end", reverseProxy(toursTarget))
	r.Any("/tours/check", reverseProxy(toursTarget))
	r.Any("/cart", reverseProxy(toursTarget))
	r.Any("/cart/*path", reverseProxy(toursTarget))
	r.Any("/comments", reverseProxy(toursTarget))

	// Rute za Followers (Go + Neo4j)
	r.Any("/followers", reverseProxy(followersTarget))
	r.Any("/followers/recommendations", reverseProxy(followersTarget))

	log.Println("Gateway running on port 8000")
	r.Run(":8000")
}

func reverseProxy(target string) gin.HandlerFunc {
	return func(c *gin.Context) {
		remote, err := url.Parse(target)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid proxy target"})
			return
		}

		proxy := httputil.NewSingleHostReverseProxy(remote)
		proxy.Director = func(req *http.Request) {
			req.Header = c.Request.Header
			req.Host = remote.Host
			req.URL.Scheme = remote.Scheme
			req.URL.Host = remote.Host
			req.URL.Path = c.Request.URL.Path
		}

		proxy.ServeHTTP(c.Writer, c.Request)
	}
}