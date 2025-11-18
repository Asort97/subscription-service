package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

var db *pgxpool.Pool

type SubscriptionInput struct {
	ID          int    `json:"id"`
	ServiceName string `json:"service_name"`
	Price       int    `json:"price"`
	UserID      string `json:"user_id"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
}

func main() {
	ctx := context.Background()

	dbURL := "postgres://app:app@localhost:5432/subscriptions?sslmode=disable"

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("failed to create db pool: %v", err)
	}

	ctxPing, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctxPing); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}

	db = pool

	r := gin.Default()

	r.GET("/subscriptions", getSubscriptionsHandle)
	r.GET("/subscriptions/:id", getSubscriptionByIDHandle)
	r.GET("/subscriptions/summary", getSubscriptionsSummaryHandle)
	r.POST("/subscriptions", postSubscriptionHandle)
	r.DELETE("//subscriptions/:id", deleteSubscriptionHandle)
	r.PUT("/subscriptions/:id", putSubscriptionHandle)

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("server error: %v", err)
	}
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

	result, err := db.Exec(
		c.Request.Context(),
		`UPDATE subscriptions 
         SET service_name = $1, price = $2, user_id = $3, start_date = $4, end_date = $5
         WHERE id = $6`,
		input.ServiceName, input.Price, input.UserID, input.StartDate, input.EndDate, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "failed to update",
		})
		return
	}

	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": "subscription not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "updated",
	})
}

func postSubscriptionHandle(c *gin.Context) {
	var input SubscriptionInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "failed to bind json",
		})
		return
	}

	var id int
	err := db.QueryRow(
		c.Request.Context(),
		`INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date)
         VALUES ($1, $2, $3, $4, $5)
         RETURNING id`,
		input.ServiceName,
		input.Price,
		input.UserID,
		input.StartDate,
		input.EndDate,
	).Scan(&id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "failed to insert subscription",
		})
		return
	}

	sub := SubscriptionInput{
		ID:          id,
		ServiceName: input.ServiceName,
		Price:       input.Price,
		UserID:      input.UserID,
		StartDate:   input.StartDate,
		EndDate:     input.EndDate,
	}

	c.JSON(http.StatusCreated, sub)
}

func getSubscriptionsHandle(c *gin.Context) {
	rows, err := db.Query(c.Request.Context(), `SELECT id, service_name, price, user_id, start_date, end_date FROM subscriptions`)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "failed to fetch subscriptions",
		})
		return
	}

	defer rows.Close()

	var result []SubscriptionInput

	for rows.Next() {
		var s SubscriptionInput
		err := rows.Scan(
			&s.ID,
			&s.ServiceName,
			&s.Price,
			&s.UserID,
			&s.StartDate,
			&s.EndDate,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "failed to scan row",
			})
			return
		}

		result = append(result, s)
	}

	c.JSON(http.StatusOK, result)
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

	var sub SubscriptionInput
	err = db.QueryRow(
		c.Request.Context(),
		`SELECT id, service_name, price, user_id, start_date, end_date 
         FROM subscriptions 
         WHERE id = $1`,
		id,
	).Scan(&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID, &sub.StartDate, &sub.EndDate)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": "subscription not found",
		})
		return
	}

	c.JSON(http.StatusOK, sub)
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

	result, err := db.Exec(
		c.Request.Context(),
		`DELETE FROM subscriptions WHERE id = $1`,
		id,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": "subscription not found",
		})
		return
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": "subscription not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "deleted",
	})
}

func getSubscriptionsSummaryHandle(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")

	if from == "" || to == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "from and to are required",
		})
		return
	}

	userID := c.Query("user_id")
	serviceName := c.Query("service_name")

	fromYear, fromMonth, err := parseYearMonth(from)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": "failed parse year and month",
		})
		return
	}

	toYear, toMonth, err := parseYearMonth(to)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": "failed parse year and month",
		})
		return
	}

	periodFrom := ymToInt(fromYear, fromMonth)
	periodTo := ymToInt(toYear, toMonth)

	rows, err := db.Query(c.Request.Context(), `SELECT id, service_name, price, user_id, start_date, end_date FROM subscriptions`)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "failed to fetch subscriptions",
		})
		return
	}

	defer rows.Close()

	var result []SubscriptionInput

	for rows.Next() {
		var s SubscriptionInput
		err := rows.Scan(
			&s.ID,
			&s.ServiceName,
			&s.Price,
			&s.UserID,
			&s.StartDate,
			&s.EndDate,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "failed to scan row",
			})
			return
		}

		result = append(result, s)
	}

	total := 0

	for _, sub := range result {
		log.Println("filter user_id =", userID, "row user_id =", sub.UserID)
		log.Println("filter service_name =", serviceName, "row service_name =", sub.ServiceName)

		if userID != "" && strings.TrimSpace(userID) != strings.TrimSpace(sub.UserID) {
			continue
		}

		if serviceName != "" && strings.EqualFold(strings.TrimSpace(serviceName), strings.TrimSpace(sub.ServiceName)) == false {
			continue
		}

		subFromY, subFromM, err := parseYearMonth(sub.StartDate)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "error",
				"message": "failed parse year and month",
			})
			return
		}

		subPeriodFrom := ymToInt(subFromY, subFromM)

		var subPeriodTo int

		if sub.EndDate == "" {
			subPeriodTo = ymToInt(9999, 12)
		} else {
			subToY, subToM, err := parseYearMonth(sub.EndDate)
			if err != nil {
				continue
			}
			subPeriodTo = ymToInt(subToY, subToM)
		}

		start := periodFrom
		if subPeriodFrom > start {
			start = subPeriodFrom
		}

		end := periodTo
		if subPeriodTo < end {
			end = subPeriodTo
		}

		if start > end {
			continue
		}

		months := end - start + 1
		total += months * sub.Price
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"total":  total,
	})
}

func ymToInt(year, month int) int {
	return year*12 + month
}

func parseYearMonth(s string) (year int, month int, err error) {
	parts := strings.Split(s, "-")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("bad date format")
	}

	year, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, err
	}

	month, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, err
	}

	return year, month, nil
}
