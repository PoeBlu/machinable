package documents

import (
	"net/http"

	"bitbucket.org/nsjostrom/machinable/dsi/interfaces"
	"github.com/gin-gonic/gin"
)

// New returns a pointer to a new `Documents` struct
func New(db interfaces.Datastore) *Documents {
	return &Documents{
		store: db,
	}
}

// Documents contains the datastore and any HTTP handlers for project resource documents
type Documents struct {
	store interfaces.Datastore
}

// AddObject creates a new document of the resource definition
func (h *Documents) AddObject(c *gin.Context) {
	resourcePathName := c.Param("resourcePathName")
	projectSlug := c.MustGet("project").(string)
	fieldValues := make(map[string]interface{})

	c.BindJSON(&fieldValues)

	newID, err := h.store.AddDefDocument(projectSlug, resourcePathName, fieldValues)
	if err != nil {
		c.JSON(err.Code(), gin.H{"error": err.Error()})
		return
	}

	// Set the inserted ID for the response
	fieldValues["id"] = newID

	c.JSON(http.StatusCreated, fieldValues)
}

// ListObjects returns the list of objects for a resource
func (h *Documents) ListObjects(c *gin.Context) {
	resourcePathName := c.Param("resourcePathName")
	projectSlug := c.MustGet("project").(string)

	documents, err := h.store.ListDefDocuments(projectSlug, resourcePathName, 0, 0, nil)

	if err != nil {
		c.JSON(err.Code(), gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"items": documents, "count": len(documents)})
}

// GetObject returns a single object with the resourceID for this resource
func (h *Documents) GetObject(c *gin.Context) {
	resourcePathName := c.Param("resourcePathName")
	resourceID := c.Param("resourceID")
	projectSlug := c.MustGet("project").(string)

	document, err := h.store.GetDefDocument(projectSlug, resourcePathName, resourceID)

	if err != nil {
		c.JSON(err.Code(), gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, document)
}

// DeleteObject deletes the object from the collection
func (h *Documents) DeleteObject(c *gin.Context) {
	resourcePathName := c.Param("resourcePathName")
	resourceID := c.Param("resourceID")
	projectSlug := c.MustGet("project").(string)

	err := h.store.DeleteDefDocument(projectSlug, resourcePathName, resourceID)

	if err != nil {
		c.JSON(err.Code(), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{})
}