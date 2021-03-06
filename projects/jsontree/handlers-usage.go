package jsontree

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/machinable/machinable/dsi/models"
)

// Usage wraps usage statistics for JSON keys
type Usage struct {
	RequestCount      int64         `json:"request_count"`
	TotalResponseTime int64         `json:"-"`
	AvgResponse       int64         `json:"avg_response"`
	StatusCodes       map[int]int64 `json:"status_codes"`
}

// ListUsage returns the list of activity logs for a project and endpoint type
func (d *Handlers) ListUsage(c *gin.Context) {
	projectID := c.MustGet("projectId").(string)

	// filter anything within x hours
	old := time.Now().Add(-time.Hour * time.Duration(1))
	filter := &models.Filters{
		"created": models.Value{
			models.GTE: old,
		},
		"endpoint_type": models.Value{
			models.EQ: models.EndpointJSON,
		},
	}

	// TODO: base this on the api limit for the customer tier
	iLimit := int64(10000)
	logs, err := d.db.ListProjectLogs(projectID, iLimit, 0, filter, nil)

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

		response[aligned] = data
	}

	// get average response time
	for key, usage := range response {
		usage.AvgResponse = (usage.TotalResponseTime / usage.RequestCount)
		response[key] = usage
	}

	c.PureJSON(http.StatusOK, gin.H{"items": response})
}
