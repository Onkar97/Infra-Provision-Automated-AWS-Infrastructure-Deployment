package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"my-project/db"
	"my-project/logs"
	"my-project/models"
)

// ProductRequest matches the expected JSON input
type ProductRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Sku          string `json:"sku"`
	Manufacturer string `json:"manufacturer"`
	Quantity     *int   `json:"quantity"` // Pointer to distinguish between 0 and missing
}

// Helper to validate common strict rules (No Query Params, Content-Length check)
func isValidRequest(c *gin.Context, requireBody bool) bool {
	if len(c.Request.URL.Query()) > 0 {
		return false
	}
	// Check for Auth Header presence if needed (Node code checked it in some places)
	// checks body existence
	if requireBody {
		if c.Request.ContentLength == 0 && c.Request.Header.Get("Transfer-Encoding") == "" {
			return false
		}
	} else {
		// If body is NOT allowed (e.g. GET with body)
		if c.Request.ContentLength > 0 {
			return false
		}
	}
	return true
}

// CreateProduct handles creating a new product
func CreateProduct(c *gin.Context) {
	// Authentication check
	authUserInterface, exists := c.Get("user")
	if !exists {
		c.Status(http.StatusUnauthorized)
		return
	}
	authUser := authUserInterface.(*models.User)

	if !isValidRequest(c, true) {
		c.Status(http.StatusBadRequest)
		return
	}

	var req ProductRequest
	// Strict JSON decoding to catch "unexpected keys"
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// Manual Validation to match Node's strict "string only" and existence checks
	if req.Name == "" || req.Description == "" || req.Sku == "" || req.Manufacturer == "" || req.Quantity == nil {
		c.Status(http.StatusBadRequest)
		return
	}
	if *req.Quantity < 0 || *req.Quantity > 100 {
		c.Status(http.StatusBadRequest)
		return
	}

	newProduct := models.Product{
		Name:         req.Name,
		Description:  req.Description,
		Sku:          req.Sku,
		Manufacturer: req.Manufacturer,
		Quantity:     *req.Quantity,
		OwnerUserID:  authUser.ID,
		DateAdded:    time.Now(),
		DateLastUpdated: time.Now(),
	}

	// --- DB: Insert Product (Timer) ---
	startInsert := time.Now()
	
	if err := db.DB.Create(&newProduct).Error; err != nil {
		logs.Error("Product insert failed: " + err.Error())
		c.Status(http.StatusBadRequest) // Generic bad request for db errors (like constraints)
		return
	}

	insertDurationMs := float64(time.Since(startInsert).Milliseconds())
	logs.Info("Query executed in " + strconv.FormatFloat(insertDurationMs, 'f', 2, 64) + "ms")
	// metricsClient.Timing("db.query.latency.createProduct", insertDurationMs)

	c.JSON(http.StatusCreated, newProduct)
}

// GetProduct retrieves a single product by ID
func GetProduct(c *gin.Context) {
	if c.Request.Method == "HEAD" {
		c.Status(http.StatusMethodNotAllowed)
		return
	}

	productIdParam := c.Param("productId")
	id, err := strconv.Atoi(productIdParam)

	// Strict validation: ID must be int, no body, no query string
	if err != nil || !isValidRequest(c, false) {
		c.Status(http.StatusBadRequest)
		return
	}

	// --- DB: Find Product (Timer) ---
	startFind := time.Now()
	
	var product models.Product
	result := db.DB.First(&product, id)

	findDurationMs := float64(time.Since(startFind).Milliseconds())
	logs.Info("Query executed in " + strconv.FormatFloat(findDurationMs, 'f', 2, 64) + "ms")
	// metricsClient.Timing("db.query.latency.getProduct", findDurationMs)

	if result.Error != nil {
		c.Status(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, product)
}

// GetAllProduct retrieves all products (Public route?)
// Note: Node code selects "product", effectively selecting all fields
func GetAllProduct(c *gin.Context) {
	if c.Request.Method == "HEAD" {
		c.Status(http.StatusMethodNotAllowed)
		return
	}

	// Validation: No body, no query string
	if !isValidRequest(c, false) {
		c.Status(http.StatusBadRequest)
		return
	}

	// --- DB: Find All (Timer) ---
	startFind := time.Now()

	var products []models.Product
	db.DB.Find(&products) // GetRawMany equivalent

	findDurationMs := float64(time.Since(startFind).Milliseconds())
	logs.Info("Query executed in " + strconv.FormatFloat(findDurationMs, 'f', 2, 64) + "ms")
	// metricsClient.Timing("db.query.latency.getAllProduct", findDurationMs)

	c.JSON(http.StatusOK, products)
}

// UpdatePutProduct handles full updates (PUT)
func UpdatePutProduct(c *gin.Context) {
	authUserInterface, exists := c.Get("user")
	if !exists {
		c.Status(http.StatusUnauthorized)
		return
	}
	authUser := authUserInterface.(*models.User)

	productIdParam := c.Param("productId")
	id, err := strconv.Atoi(productIdParam)

	if err != nil || !isValidRequest(c, true) {
		c.Status(http.StatusBadRequest)
		return
	}

	// Check existence and ownership
	var product models.Product
	if err := db.DB.First(&product, id).Error; err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	if product.OwnerUserID != authUser.ID {
		c.Status(http.StatusForbidden)
		return
	}

	var req ProductRequest
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// PUT requires ALL fields to be present
	if req.Name == "" || req.Description == "" || req.Sku == "" || req.Manufacturer == "" || req.Quantity == nil {
		c.Status(http.StatusBadRequest)
		return
	}
	if *req.Quantity < 0 || *req.Quantity > 100 {
		c.Status(http.StatusBadRequest)
		return
	}

	// --- DB: Update Product (Timer) ---
	startUpdate := time.Now()

	updates := map[string]interface{}{
		"name":              req.Name,
		"description":       req.Description,
		"sku":               req.Sku,
		"manufacturer":      req.Manufacturer,
		"quantity":          *req.Quantity,
		"date_last_updated": time.Now(),
	}

	if err := db.DB.Model(&product).Updates(updates).Error; err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	updateDurationMs := float64(time.Since(startUpdate).Milliseconds())
	logs.Info("Query executed in " + strconv.FormatFloat(updateDurationMs, 'f', 2, 64) + "ms")
	// metricsClient.Timing("db.query.latency.updatePutProduct", updateDurationMs)

	c.Status(http.StatusNoContent)
}

// UpdatePatchProduct handles partial updates (PATCH)
func UpdatePatchProduct(c *gin.Context) {
	authUserInterface, exists := c.Get("user")
	if !exists {
		c.Status(http.StatusUnauthorized)
		return
	}
	authUser := authUserInterface.(*models.User)

	productIdParam := c.Param("productId")
	id, err := strconv.Atoi(productIdParam)

	if err != nil || !isValidRequest(c, true) {
		c.Status(http.StatusBadRequest)
		return
	}

	// Check existence and ownership
	var product models.Product
	if err := db.DB.First(&product, id).Error; err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	if product.OwnerUserID != authUser.ID {
		c.Status(http.StatusForbidden)
		return
	}

	// Use a map for PATCH to know exactly which fields were sent
	var reqMap map[string]interface{}
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&reqMap); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// Validate keys against allowed fields
	allowedFields := map[string]bool{"name": true, "description": true, "sku": true, "manufacturer": true, "quantity": true}
	for key := range reqMap {
		if !allowedFields[key] {
			c.Status(http.StatusBadRequest)
			return
		}
	}

	// Validate Quantity if present
	if val, ok := reqMap["quantity"]; ok {
		// JSON numbers are float64 by default in map[string]interface{}
		qFloat, ok := val.(float64)
		if !ok || qFloat < 0 || qFloat > 100 || qFloat != float64(int(qFloat)) {
			c.Status(http.StatusBadRequest)
			return
		}
	}

	reqMap["date_last_updated"] = time.Now()

	// --- DB: Update Product (Timer) ---
	startUpdate := time.Now()

	if err := db.DB.Model(&product).Updates(reqMap).Error; err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	updateDurationMs := float64(time.Since(startUpdate).Milliseconds())
	logs.Info("Query executed in " + strconv.FormatFloat(updateDurationMs, 'f', 2, 64) + "ms")
	// metricsClient.Timing("db.query.latency.updatePatchProduct", updateDurationMs)

	c.Status(http.StatusNoContent)
}

// DeleteProduct handles deletion
func DeleteProduct(c *gin.Context) {
	authUserInterface, exists := c.Get("user")
	if !exists {
		c.Status(http.StatusUnauthorized) // Node code returns 401 if !req.user
		return
	}
	authUser := authUserInterface.(*models.User)

	productIdParam := c.Param("productId")
	id, err := strconv.Atoi(productIdParam)

	// Validation: No query string, No body
	if err != nil || len(c.Request.URL.Query()) > 0 || c.Request.ContentLength > 0 {
		c.Status(http.StatusBadRequest)
		return
	}

	// --- DB: Find Product ---
	startFind := time.Now()
	var product models.Product
	if err := db.DB.First(&product, id).Error; err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	findDurationMs := float64(time.Since(startFind).Milliseconds())
	logs.Info("Query executed in " + strconv.FormatFloat(findDurationMs, 'f', 2, 64) + "ms")
	// metricsClient.Timing("db.query.latency.findProduct", findDurationMs)

	if product.OwnerUserID != authUser.ID {
		c.Status(http.StatusForbidden)
		return
	}

	// --- DB: Delete Images (Timer) ---
	startDeleteImg := time.Now()
	
	// Delete all images associated with this product
	db.DB.Where("product_id = ?", id).Delete(&models.Image{})

	delImgDurationMs := float64(time.Since(startDeleteImg).Milliseconds())
	logs.Info("Query executed in " + strconv.FormatFloat(delImgDurationMs, 'f', 2, 64) + "ms")
	// metricsClient.Timing("db.query.latency.deleteImages", delImgDurationMs)

	// --- DB: Delete Product (Timer) ---
	startDeleteProd := time.Now()

	db.DB.Delete(&product)

	delProdDurationMs := float64(time.Since(startDeleteProd).Milliseconds())
	logs.Info("Query executed in " + strconv.FormatFloat(delProdDurationMs, 'f', 2, 64) + "ms")
	// metricsClient.Timing("db.query.latency.deleteProduct", delProdDurationMs)

	c.Status(http.StatusNoContent)
}