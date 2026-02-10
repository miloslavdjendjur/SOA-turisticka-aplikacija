package main

import (
	"time"
	"github.com/lib/pq"
)

type TourStatus string
const (
	Draft     TourStatus = "DRAFT"
	Published TourStatus = "PUBLISHED"
	Archived  TourStatus = "ARCHIVED"
)

type TourDifficulty string
const (
	Easy   TourDifficulty = "EASY"
	Medium TourDifficulty = "MEDIUM"
	Hard   TourDifficulty = "HARD"
)

type Tour struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	AuthorID    int            `json:"authorId"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Difficulty  TourDifficulty `json:"difficulty"`
	Tags        pq.StringArray `gorm:"type:text[]" json:"tags"`
	Status      TourStatus     `json:"status"`
	Price       float64        `json:"price"`
	Distance    float64        `json:"distance"`
	PublishDate time.Time      `json:"publishDate"`
	ArchiveDate time.Time      `json:"archiveDate"`
	
	KeyPoints   []KeyPoint     `gorm:"foreignKey:TourID" json:"keyPoints"`
}

type KeyPoint struct {
	ID          uint    `gorm:"primaryKey" json:"id"`
	TourID      uint    `json:"tourId"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Image       string  `json:"image"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Order       int     `json:"order"`
}

type Blog struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	AuthorID    int       `json:"authorId"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	DateCreated time.Time `json:"dateCreated"`
	Image       string    `json:"image"`
}


type TourPurchase struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TouristID int       `json:"touristId"`
	TourID    uint      `json:"tourId"`
	Token     string    `json:"token"` 
	PurchaseDate time.Time `json:"purchaseDate"`
}

type TouristPosition struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TouristID int       `json:"touristId"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type TourExecution struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	TouristID    int       `json:"touristId"`
	TourID       uint      `json:"tourId"`
	Status       string    `json:"status"` // "STARTED", "COMPLETED", "ABANDONED"
	StartTime    time.Time `json:"startTime"`
	EndTime      time.Time `json:"endTime"`
	LastActivity time.Time `json:"lastActivity"`
}

// TACKA 16: KORPA I STAVKE
type ShoppingCart struct {
	ID         uint        `gorm:"primaryKey" json:"id"`
	TouristID  int         `json:"touristId"`
	TotalPrice float64     `json:"totalPrice"`
	Items      []OrderItem `gorm:"foreignKey:ShoppingCartID" json:"items"`
}

type OrderItem struct {
	ID             uint    `gorm:"primaryKey" json:"id"`
	ShoppingCartID uint    `json:"shoppingCartId"`
	TourID         uint    `json:"tourId"`
	TourName       string  `json:"tourName"`
	Price          float64 `json:"price"`
}