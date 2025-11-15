package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SubscriptionInput struct {
	ID          int    `json:"id"`
	ServiceName string `json:"service_name"`
	Price       int    `json:"price"`
	UserID      string `json:"user_id"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
}

var subscriptions = []SubscriptionInput{}
var nextID int

func main() {
	r := gin.Default()

	r.GET("/subscriptions", getSubscriptionsHandle)
	r.GET("/subscriptions/:id", getSubscriptionByIDHandle)
	r.POST("/subscriptions", postSubscriptionHandle)
	r.DELETE("//subscriptions/:id", deleteSubscriptionHandle)
	r.PUT("/subscriptions/:id", putSubscriptionHandle)

	r.Run(":8080")
}

func putSubscriptionHandle(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "couldnt convert string to int",
		})
		return
	}

	var input SubscriptionInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "couldnt parse json",
		})

		return
	}

	for i := range subscriptions {
		if subscriptions[i].ID == id {
			subscriptions[i] = SubscriptionInput{
				ID:          id,
				ServiceName: input.ServiceName,
				Price:       input.Price,
				UserID:      input.UserID,
				StartDate:   input.StartDate,
				EndDate:     input.EndDate,
			}

			c.JSON(http.StatusOK, gin.H{
				"status":  "success",
				"message": subscriptions[i],
			})

			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"status":  "error",
		"message": "subscription not found",
	})
}

func postSubscriptionHandle(c *gin.Context) {
	var input SubscriptionInput
	input.ID = nextID

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "failed to bind json",
		})
		return
	}

	subscriptions = append(subscriptions, input)
	nextID++

	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"data":   input,
	})
}

func getSubscriptionsHandle(c *gin.Context) {
	c.JSON(http.StatusOK, subscriptions)
}

func getSubscriptionByIDHandle(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.Atoi(idStr)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "invalid id",
		})
		return
	}

	for _, sub := range subscriptions {
		if sub.ID == id {
			c.JSON(http.StatusOK, sub)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"status":  "error",
		"message": "subscription not found",
	})
}

func deleteSubscriptionHandle(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  "error",
			"message": "cant convert id",
		})
		return
	}

	for i := 0; i < len(subscriptions); i++ {
		if subscriptions[i].ID == id {
			subscriptions = append(subscriptions[:i], subscriptions[i+1:]...)
			c.JSON(http.StatusOK, gin.H{
				"status":  "success",
				"message": "deleted subscription",
			})

			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"status":  "error",
		"message": "subscription not found",
	})
}
