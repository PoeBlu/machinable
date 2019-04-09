package collections

import (
	"net/http"
	"time"

	"github.com/anothrnick/machinable/dsi/models"
	"github.com/gin-gonic/gin"
)

type Usage struct {
	RequestCount      int64         `json:"-"`
	TotalResponseTime int64         `json:"-"`
	AvgResponse       int64         `json:"avg_response"`
	StatusCodes       map[int]int64 `json:"status_codes"`
}

// ListCollectionUsage returns the list of activity logs for a project
func (h *Collections) ListCollectionUsage(c *gin.Context) {
	projectSlug := c.MustGet("project").(string)

	// filter anything within x hours
	old := time.Now().Add(-time.Hour * time.Duration(1))
	filter := &models.Filters{
		"created": models.Value{
			models.GTE: old.Unix(),
		},
		"endpoint_type": models.Value{
			models.EQ: models.EndpointCollection,
		},
	}

	// TODO: base this on the api limit for the customer tier
	iLimit := int64(10000)
	logs, err := h.store.ListProjectLogs(projectSlug, iLimit, 0, filter, nil)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make(map[int64]Usage)

	// transform logs
	for _, log := range logs {
		aligned := log.AlignedCreated

		data, ok := response[aligned]
		if !ok {
			data = Usage{
				StatusCodes: make(map[int]int64),
			}
		}

		data.RequestCount++
		data.TotalResponseTime += log.ResponseTime
		data.StatusCodes[log.StatusCode]++
	}

	// get average response time
	for key, usage := range response {
		usage.AvgResponse = (usage.RequestCount / usage.TotalResponseTime)
		response[key] = usage
	}

	c.PureJSON(http.StatusOK, gin.H{"items": response})
}

// ListResponseTimes returns HTTP response times for collections for the last 1 hour
func (h *Collections) ListResponseTimes(c *gin.Context) {
	projectSlug := c.MustGet("project").(string)

	old := time.Now().Add(-time.Hour * time.Duration(1))
	filter := &models.Filters{
		"timestamp": models.Value{
			models.GTE: old.Unix(),
		},
	}

	responseTimes, err := h.store.ListCollectionResponseTimes(projectSlug, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"response_times": responseTimes})
}

// ListStatusCodes returns HTTP response status codes for collections for the last 1 hour
func (h *Collections) ListStatusCodes(c *gin.Context) {
	projectSlug := c.MustGet("project").(string)

	old := time.Now().Add(-time.Hour * time.Duration(1))
	filter := &models.Filters{
		"timestamp": models.Value{
			models.GTE: old.Unix(),
		},
	}

	statusCodes, err := h.store.ListCollectionStatusCode(projectSlug, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status_codes": statusCodes})
}

// GetStats returns the size of the collections
func (h *Collections) GetStats(c *gin.Context) {
	projectSlug := c.MustGet("project").(string)

	// retrieve the list of collections
	collections, err := h.store.GetCollections(projectSlug)

	if err != nil {
		c.JSON(err.Code(), gin.H{"error": err.Error()})
		return
	}

	totalStats := &models.Stats{}
	collectionStats := map[string]*models.Stats{}
	for _, col := range collections {
		stats, err := h.store.GetCollectionStats(projectSlug, col.Name)
		if err != nil {
			c.JSON(err.Code(), gin.H{"error": err.Error()})
			return
		}

		collectionStats[col.Name] = stats
		totalStats.Size += stats.Size
		totalStats.Count += stats.Count
	}

	c.JSON(http.StatusOK, gin.H{"total": totalStats, "collections": collectionStats})
}