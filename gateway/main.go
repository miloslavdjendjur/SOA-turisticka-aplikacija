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

	// Rute za Followers (Go + Neo4j)
	r.Any("/followers", reverseProxy(followersTarget))

	log.Println("Gateway running on port 8000")
	r.Run(":8000")
}

// Pomocna funkcija koja pravi Reverse Proxy
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