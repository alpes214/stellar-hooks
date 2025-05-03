package api

import (
	"net/http"
	"strconv"

	"github.com/alpes214/stellar-hooks/internal/models"
	"github.com/alpes214/stellar-hooks/internal/storage"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, store storage.SubscriptionStore) {
	r.GET("/subscriptions", getSubscriptions(store))
	r.POST("/subscriptions", createSubscription(store))
	r.GET("/subscriptions/:id", getSubscriptionByID(store))
	r.PUT("/subscriptions/:id", updateSubscription(store))
	r.DELETE("/subscriptions/:id", deleteSubscription(store))
	r.GET("/subscriptions/status", getStatus(store))
}

// getSubscriptions godoc
// @Summary List subscriptions
// @Description Retrieve all subscriptions
// @Tags subscriptions
// @Produce json
// @Success 200 {array} models.Subscription
// @Failure 500 {object} gin.H
// @Router /subscriptions [get]
func getSubscriptions(store storage.SubscriptionStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		subs, err := store.List()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, subs)
	}
}

// createSubscription godoc
// @Summary Create subscription
// @Description Create a new subscription
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param subscription body models.Subscription true "Subscription body"
// @Success 201 {object} models.Subscription
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /subscriptions [post]
func createSubscription(store storage.SubscriptionStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		var sub models.Subscription
		if err := c.ShouldBindJSON(&sub); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		id, err := store.Create(sub)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		sub.ID = id
		c.JSON(http.StatusCreated, sub)
	}
}

// getSubscriptionByID godoc
// @Summary Get subscription by ID
// @Description Retrieve a subscription by its ID
// @Tags subscriptions
// @Produce json
// @Param id path int true "Subscription ID"
// @Success 200 {object} models.Subscription
// @Failure 400 {object} gin.H
// @Failure 404 {object} gin.H
// @Router /subscriptions/{id} [get]
func getSubscriptionByID(store storage.SubscriptionStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}
		sub, err := store.GetByID(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, sub)
	}
}

// updateSubscription godoc
// @Summary Update subscription
// @Description Update an existing subscription by ID
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path int true "Subscription ID"
// @Param subscription body models.Subscription true "Updated subscription body"
// @Success 200 {object} models.Subscription
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /subscriptions/{id} [put]
func updateSubscription(store storage.SubscriptionStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}
		var sub models.Subscription
		if err := c.ShouldBindJSON(&sub); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		sub.ID = id
		if err := store.Update(sub); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, sub)
	}
}

// deleteSubscription godoc
// @Summary Delete subscription
// @Description Delete a subscription by ID
// @Tags subscriptions
// @Produce json
// @Param id path int true "Subscription ID"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /subscriptions/{id} [delete]
func deleteSubscription(store storage.SubscriptionStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}
		if err := store.Delete(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	}
}

// getStatus godoc
// @Summary Subscription store status
// @Description Returns basic store statistics
// @Tags subscriptions
// @Produce json
// @Success 200 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /subscriptions/status [get]
func getStatus(store storage.SubscriptionStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		count, err := store.Count()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"count": count})
	}
}
