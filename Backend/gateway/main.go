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
		if c.Request.Method == "OPTIONS" {
			setCORSHeaders(c.Writer)
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	stakeholdersTarget := "http://stakeholders:8081"
	toursTarget := "http://tours:8082"
	followersTarget := "http://followers:8083"

	// --- RUTE ---
	r.Any("/api/users", reverseProxy(stakeholdersTarget))       
	r.Any("/api/users/*path", reverseProxy(stakeholdersTarget)) 

	r.Any("/tours", reverseProxy(toursTarget))
	r.Any("/keypoints", reverseProxy(toursTarget))
	r.Any("/blogs", reverseProxy(toursTarget))
	r.Any("/comments", reverseProxy(toursTarget)) 
	r.Any("/tours/purchase", reverseProxy(toursTarget))
	r.Any("/position", reverseProxy(toursTarget))
	r.Any("/tours/start", reverseProxy(toursTarget))
	r.Any("/tours/end", reverseProxy(toursTarget))
	r.Any("/tours/check", reverseProxy(toursTarget))
	r.Any("/cart", reverseProxy(toursTarget))
	r.Any("/cart/*path", reverseProxy(toursTarget))

	r.Any("/followers", reverseProxy(followersTarget))
	r.Any("/followers/*path", reverseProxy(followersTarget)) 

	log.Println("Gateway running on port 8000")
	r.Run(":8000")
}

func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
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

		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			setCORSHeaders(w) // Dodaj headers da browser ne blokira
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte("Gateway Error: Service Unreachable"))
		}

		proxy.ModifyResponse = func(r *http.Response) error {
			r.Header.Del("Access-Control-Allow-Origin")
			r.Header.Del("Access-Control-Allow-Credentials")
			r.Header.Del("Access-Control-Allow-Headers")
			r.Header.Del("Access-Control-Allow-Methods")
			
			r.Header.Set("Access-Control-Allow-Origin", "http://localhost:4200")
			r.Header.Set("Access-Control-Allow-Credentials", "true")
			r.Header.Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
			r.Header.Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
			return nil
		}

		proxy.ServeHTTP(c.Writer, c.Request)
	}
}