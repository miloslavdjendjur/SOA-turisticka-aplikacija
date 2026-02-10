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
	r.GET("/blogs", GetAllBlogs)
	r.POST("/tours/purchase", BuyTour)     
	r.POST("/position", UpdatePosition)    
	r.POST("/tours/start", StartTour)     
	r.POST("/tours/end", EndTour)
	r.POST("/tours/check", CheckProximity)
	r.GET("/cart", GetCart)          
	r.POST("/cart/add", AddToCart)   
	r.POST("/cart/checkout", Checkout) 

	r.Run(":8082")
}