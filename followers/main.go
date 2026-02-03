package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var driver neo4j.DriverWithContext

func initNeo4j() {
	uri := os.Getenv("NEO4J_URI")
	username := os.Getenv("NEO4J_USER")
	password := os.Getenv("NEO4J_PASSWORD")

	var err error
	
	driver, err = neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		log.Fatal("Failed to create Neo4j driver: ", err)
	}

	
	err = driver.VerifyConnectivity(context.Background())
	if err != nil {
		log.Println("Failed to connect to Neo4j, but continuing... (Service might wait for DB)")
	} else {
		log.Println("Connected to Neo4j successfully!")
	}
}


type FollowRequest struct {
	FollowerID int `json:"followerId"` 
	FollowedID int `json:"followedId"` 
}

// TACKA 9: Zaprati korisnika
func FollowUser(c *gin.Context) {
	var req FollowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	
	
	query := `
		MERGE (u1:User {id: $followerId})
		MERGE (u2:User {id: $followedId})
		MERGE (u1)-[:FOLLOWS]->(u2)
		RETURN u1, u2
	`
	params := map[string]any{
		"followerId": req.FollowerID,
		"followedId": req.FollowedID,
	}

	_, err := session.Run(ctx, query, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User followed successfully"})
}

func GetFollowedUsers(c *gin.Context) {
    
	followerID := c.Query("userId") 
    
	ctx := context.Background()
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	query := `
		MATCH (u1:User {id: $id})-[:FOLLOWS]->(u2:User)
		RETURN u2.id as followedId
	`
  
	result, err := session.Run(ctx, query, map[string]any{"id": followerID}) // Paznja: ovde treba cast u int ako neo4j cuva int
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var followedIDs []int
	for result.Next(ctx) {
		record := result.Record()
		val, _ := record.Get("followedId")
        
		if id, ok := val.(int64); ok {
			followedIDs = append(followedIDs, int(id))
		}
	}

	c.JSON(http.StatusOK, followedIDs)
}

func main() {
	initNeo4j()
	r := gin.Default()

	r.POST("/followers", FollowUser)
	r.GET("/followers", GetFollowedUsers)

	r.Run(":8083")
}