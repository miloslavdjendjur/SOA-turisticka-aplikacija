package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	ConnectDatabase()

	r := gin.Default()

	
	r.POST("/tours", CreateTour)
	r.GET("/tours", GetTours)
	r.POST("/keypoints", AddKeyPoint)
	r.POST("/blogs", CreateBlog)
	r.POST("/tours/purchase", BuyTour)     
	r.POST("/position", UpdatePosition)    
	r.POST("/tours/start", StartTour)     

	r.Run(":8082")
}