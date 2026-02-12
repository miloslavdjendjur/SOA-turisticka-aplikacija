package main

import (
	"fmt"
	"math"
	"net/http"
	"time"
	"encoding/json"

	"github.com/gin-gonic/gin"
)


func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 
	dLat := (lat2 - lat1) * (math.Pi / 180.0)
	dLon := (lon2 - lon1) * (math.Pi / 180.0)
	lat1Rad := lat1 * (math.Pi / 180.0)
	lat2Rad := lat2 * (math.Pi / 180.0)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

// --- STANDARDNI HANDLERI ---

func CreateTour(c *gin.Context) {
	var input Tour
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.Status = Draft
	input.Distance = 0
	input.Price = 0
	if err := DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, input)
}


func GetAllBlogs(c *gin.Context) {
	touristID := c.Query("userId")

	if touristID == "" {
		c.JSON(http.StatusOK, []Blog{}) 
		return
	}

	followersURL := fmt.Sprintf("http://followers:8083/followers?userId=%s", touristID)
	resp, err := http.Get(followersURL)
	
	var followedIDs []int
	if err == nil && resp.StatusCode == 200 {
		json.NewDecoder(resp.Body).Decode(&followedIDs)
		resp.Body.Close()
	}

	var myID int
	fmt.Sscanf(touristID, "%d", &myID)
	followedIDs = append(followedIDs, myID)

	
	var blogs []Blog
	
	if err := DB.Where("author_id IN ?", followedIDs).Find(&blogs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, blogs)
}

func AddKeyPoint(c *gin.Context) {
	var input KeyPoint
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, input)
}

func CreateBlog(c *gin.Context) {
	var input Blog
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.DateCreated = time.Now()
	if err := DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, input)
}

func GetTours(c *gin.Context) {
	var tours []Tour
	DB.Preload("KeyPoints").Find(&tours)
	c.JSON(http.StatusOK, tours)
}


func BuyTour(c *gin.Context) {
	var purchase TourPurchase
	if err := c.ShouldBindJSON(&purchase); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var exist TourPurchase
	if err := DB.Where("tourist_id = ? AND tour_id = ?", purchase.TouristID, purchase.TourID).First(&exist).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{"message": "Vec ste kupili ovu turu", "token": exist.Token})
		return
	}

	purchase.Token = fmt.Sprintf("TOKEN-%d-%d-%d", purchase.TouristID, purchase.TourID, time.Now().Unix())
	purchase.PurchaseDate = time.Now()

	if err := DB.Create(&purchase).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, purchase)
}


func UpdatePosition(c *gin.Context) {
	var pos TouristPosition
	if err := c.ShouldBindJSON(&pos); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	pos.UpdatedAt = time.Now()
	var existing TouristPosition
	if err := DB.Where("tourist_id = ?", pos.TouristID).First(&existing).Error; err == nil {
		existing.Latitude = pos.Latitude
		existing.Longitude = pos.Longitude
		existing.UpdatedAt = time.Now()
		DB.Save(&existing)
		c.JSON(http.StatusOK, existing)
	} else {
		DB.Create(&pos)
		c.JSON(http.StatusCreated, pos)
	}
}


func StartTour(c *gin.Context) {
	var exec TourExecution
	if err := c.ShouldBindJSON(&exec); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var purchase TourPurchase
	if err := DB.Where("tourist_id = ? AND tour_id = ?", exec.TouristID, exec.TourID).First(&purchase).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Tura nije kupljena!"})
		return
	}

	exec.Status = "STARTED"
	exec.StartTime = time.Now()
	exec.LastActivity = time.Now()

	if err := DB.Create(&exec).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Greska pri kreiranju sesije"})
		return
	}
	c.JSON(http.StatusCreated, exec)
}

func EndTour(c *gin.Context) {
	var req struct {
		TouristID int    `json:"touristId"`
		TourID    int    `json:"tourId"`
		Status    string `json:"status"` 
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var exec TourExecution
	if err := DB.Where("tourist_id = ? AND tour_id = ? AND status = ?", req.TouristID, req.TourID, "STARTED").First(&exec).Error; err != nil {
		c.JSON(404, gin.H{"error": "Nema aktivne sesije"})
		return
	}

	
	exec.Status = req.Status 
	exec.EndTime = time.Now()
	exec.LastActivity = time.Now()

	DB.Save(&exec) 

	c.JSON(http.StatusOK, gin.H{"message": "Sesija zavrsena", "status": exec.Status})
}

func CheckProximity(c *gin.Context) {
	var req struct {
		TouristID int `json:"touristId"`
		TourID    int `json:"tourId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var exec TourExecution
	if err := DB.Where("tourist_id = ? AND tour_id = ? AND status = ?", req.TouristID, req.TourID, "STARTED").First(&exec).Error; err != nil {
		c.JSON(404, gin.H{"error": "Sesija nije pronadjena"})
		return
	}

	// Bez obzira na ishod, belezi se last activity (Tacka 17)
	exec.LastActivity = time.Now()
	DB.Save(&exec)

	var pos TouristPosition
	if err := DB.Where("tourist_id = ?", req.TouristID).First(&pos).Error; err != nil {
		c.JSON(400, gin.H{"error": "Lokacija nije pronadjena"})
		return
	}

	var keypoints []KeyPoint
	DB.Where("tour_id = ?", req.TourID).Find(&keypoints)

	found := false
	nearbyPoint := ""

	for _, kp := range keypoints {
		dist := calculateDistance(pos.Latitude, pos.Longitude, kp.Latitude, kp.Longitude)
		if dist < 0.1 { // Blizu tacke (100m)
			found = true
			nearbyPoint = kp.Name
			// Ovde bi se u sesiji dodalo da je tacka kompletirana
			break
		}
	}

	c.JSON(200, gin.H{"nearKeyPoint": found, "pointName": nearbyPoint})
}

func Checkout(c *gin.Context) {
	var req struct {
		TouristID int `json:"touristId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var cart ShoppingCart
	if err := DB.Preload("Items").Where("tourist_id = ?", req.TouristID).First(&cart).Error; err != nil {
		c.JSON(400, gin.H{"error": "Korpa prazna"})
		return
	}

	for _, item := range cart.Items {
		purchase := TourPurchase{
			TouristID:    req.TouristID,
			TourID:       item.TourID,
			Token:        fmt.Sprintf("TKN-%d", time.Now().UnixNano()),
			PurchaseDate: time.Now(),
		}
		DB.Create(&purchase)
	}

	DB.Where("shopping_cart_id = ?", cart.ID).Delete(&OrderItem{})
	cart.TotalPrice = 0
	DB.Save(&cart)

	c.JSON(200, gin.H{"message": "Kupljeno! Korpa je prazna."})
}


func GetCart(c *gin.Context) {
	touristID := c.Query("touristId")
	var cart ShoppingCart

	if err := DB.Preload("Items").Where("tourist_id = ?", touristID).First(&cart).Error; err != nil {
		
		c.JSON(http.StatusOK, gin.H{"items": []interface{}{}, "totalPrice": 0})
		return
	}
	c.JSON(http.StatusOK, cart)
}


func AddToCart(c *gin.Context) {
	var req struct {
		TouristID int     `json:"touristId"`
		TourID    uint    `json:"tourId"`
		TourName  string  `json:"tourName"`
		Price     float64 `json:"price"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var cart ShoppingCart
	if err := DB.Where("tourist_id = ?", req.TouristID).First(&cart).Error; err != nil {
		cart = ShoppingCart{TouristID: req.TouristID, TotalPrice: 0}
		DB.Create(&cart)
	}


	item := OrderItem{
		ShoppingCartID: cart.ID,
		TourID:         req.TourID,
		TourName:       req.TourName,
		Price:          req.Price,
	}
	DB.Create(&item)

	
	cart.TotalPrice += req.Price
	DB.Save(&cart)

	c.JSON(http.StatusOK, gin.H{"message": "Dodato u korpu"})
}

// func Checkout(c *gin.Context) {
// 	var req struct {
// 		TouristID int `json:"touristId"`
// 	}
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}


// 	var cart ShoppingCart
// 	if err := DB.Preload("Items").Where("tourist_id = ?", req.TouristID).First(&cart).Error; err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Korpa je prazna"})
// 		return
// 	}

// 	if len(cart.Items) == 0 {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Korpa je prazna"})
// 		return
// 	}


// 	for _, item := range cart.Items {
// 		purchase := TourPurchase{
// 			TouristID:    req.TouristID,
// 			TourID:       item.TourID,
// 			Token:        fmt.Sprintf("TOKEN-%d-%d-%d", req.TouristID, item.TourID, time.Now().Unix()),
// 			PurchaseDate: time.Now(),
// 		}

// 		var exists TourPurchase
// 		if err := DB.Where("tourist_id = ? AND tour_id = ?", req.TouristID, item.TourID).First(&exists).Error; err != nil {
// 			DB.Create(&purchase)
// 		}
// 	}

	
// 	DB.Where("shopping_cart_id = ?", cart.ID).Delete(&OrderItem{})
// 	cart.TotalPrice = 0
// 	DB.Save(&cart)

// 	c.JSON(http.StatusOK, gin.H{"message": "Uspešna kupovina! Tokeni generisani."})
// }

func CreateComment(c *gin.Context) {
	var comm Comment
	if err := c.ShouldBindJSON(&comm); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}


	var blog Blog
	if err := DB.First(&blog, comm.BlogID).Error; err != nil {
		c.JSON(404, gin.H{"error": "Blog ne postoji"})
		return
	}


	if comm.AuthorID != blog.AuthorID {

		followersURL := "http://followers:8083"
		checkURL := fmt.Sprintf("%s/followers/check?followerId=%d&followedId=%d", followersURL, comm.AuthorID, blog.AuthorID)
		
		resp, err := http.Get(checkURL)
		if err != nil || resp.StatusCode != 200 {
			c.JSON(500, gin.H{"error": "Greska pri proveri pracenja"})
			return
		}
		defer resp.Body.Close()

		var result struct {
			IsFollowing bool `json:"isFollowing"`
		}
		json.NewDecoder(resp.Body).Decode(&result)

		if !result.IsFollowing {
			c.JSON(403, gin.H{"error": "Morate zapratiti autora da biste komentarisali!"})
			return
		}
	}

	comm.DateCreated = time.Now()
	DB.Create(&comm)
	c.JSON(201, comm)
}

func GetComments(c *gin.Context) {
	blogID := c.Query("blogId")
	var comments []Comment
	DB.Where("blog_id = ?", blogID).Find(&comments)
	c.JSON(200, comments)
}