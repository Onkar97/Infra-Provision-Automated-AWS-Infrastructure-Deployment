package controllers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"my-project/db"
	"my-project/logs"
	"my-project/models"
)

var s3Client *s3.Client

// Initialize S3 Client (This runs automatically when the package loads)
func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		logs.Error("Unable to load SDK config for S3: " + err.Error())
	} else {
		s3Client = s3.NewFromConfig(cfg)
	}
}

// CreateImage handles file upload to S3 and DB insertion
func CreateImage(c *gin.Context) {
	// 1. Authentication Check
	authUserInterface, exists := c.Get("user")
	if !exists {
		c.Status(http.StatusUnauthorized)
		return
	}
	authUser := authUserInterface.(*models.User)

	productIdParam := c.Param("productId")
	productId, _ := strconv.Atoi(productIdParam)

	// 2. Strict Validation (Query Params)
	if len(c.Request.URL.Query()) > 0 {
		c.Status(http.StatusBadRequest)
		return
	}

	// 3. File Handling (Equivalent to Multer)
	fileHeader, err := c.FormFile("file")
	if err != nil {
		logs.Error("Cannot find file")
		c.Status(http.StatusBadRequest)
		return
	}

	// 4. Validate Mime Type (Equivalent to imageFileFilter)
	contentType := fileHeader.Header.Get("Content-Type")
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/jpg" {
		c.Status(http.StatusBadRequest)
		return
	}

	// --- DB: Find Product (Timer) ---
	startFind := time.Now()
	var product models.Product
	if err := db.DB.First(&product, productId).Error; err != nil {
		logs.Info("Cannot find Product")
		c.Status(http.StatusNotFound)
		return
	}
	findDurationMs := float64(time.Since(startFind).Milliseconds())
	logs.Info("Query executed in " + strconv.FormatFloat(findDurationMs, 'f', 2, 64) + "ms")
	// metricsClient.Timing("db.query.latency.findProduct", findDurationMs)

	// Check Ownership
	if product.OwnerUserID != authUser.ID {
		c.Status(http.StatusForbidden)
		return
	}

	// 5. Open File Stream for S3 Upload
	fileContent, err := fileHeader.Open()
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	defer fileContent.Close()

	// 6. Generate Unique Key
	uniqueFileName := fmt.Sprintf("%s-%s", uuid.New().String(), fileHeader.Filename)
	s3Key := fmt.Sprintf("%d/%d/%s", authUser.ID, productId, uniqueFileName)

	// --- S3: Upload (Timer) ---
	startS3 := time.Now()
	_, err = s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(os.Getenv("S3_BUCKET_NAME")),
		Key:         aws.String(s3Key),
		Body:        fileContent, // Stream directly
		ContentType: aws.String(contentType),
	})
	if err != nil {
		logs.Error("S3 Upload failed: " + err.Error())
		c.Status(http.StatusServiceUnavailable)
		return
	}
	s3DurationMs := float64(time.Since(startS3).Milliseconds())
	logs.Info("S3 Upload executed in " + strconv.FormatFloat(s3DurationMs, 'f', 2, 64) + "ms")
	// metricsClient.Timing("s3.upload.latency", s3DurationMs)

	// 7. Insert into DB
	newImage := models.Image{
		ProductID:    uint(productId),
		FileName:     fileHeader.Filename,
		S3BucketPath: s3Key,
		DateCreated:  time.Now(),
	}

	// --- DB: Insert Image (Timer) ---
	startInsert := time.Now()
	if err := db.DB.Create(&newImage).Error; err != nil {
		logs.Error("Image insert failed: " + err.Error())
		c.Status(http.StatusBadRequest)
		return
	}
	insertDurationMs := float64(time.Since(startInsert).Milliseconds())
	logs.Info("Query executed in " + strconv.FormatFloat(insertDurationMs, 'f', 2, 64) + "ms")
	// metricsClient.Timing("db.query.latency.createImage", insertDurationMs)

	c.JSON(http.StatusCreated, newImage)
}

// GetImage retrieves image metadata
func GetImage(c *gin.Context) {
	if c.Request.Method == "HEAD" {
		c.Status(http.StatusMethodNotAllowed)
		return
	}

	productIdParam := c.Param("productId")
	imageIdParam := c.Param("imageId")

	pId, errP := strconv.Atoi(productIdParam)
	iId, errI := strconv.Atoi(imageIdParam)

	// Validation
	if errP != nil || errI != nil || len(c.Request.URL.Query()) > 0 || c.Request.ContentLength > 0 || c.GetHeader("Authorization") != "" {
		c.Status(http.StatusBadRequest)
		return
	}

	// --- DB: Find Image (Timer) ---
	startFind := time.Now()
	var image models.Image
	// Note: We check both image_id and product_id to match your logic, though image_id is PK
	result := db.DB.Where("image_id = ? AND product_id = ?", iId, pId).First(&image)

	findDurationMs := float64(time.Since(startFind).Milliseconds())
	logs.Info("Query executed in " + strconv.FormatFloat(findDurationMs, 'f', 2, 64) + "ms")
	// metricsClient.Timing("db.query.latency.getImage", findDurationMs)

	if result.Error != nil {
		c.Status(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, image)
}

// GetAllImage retrieves all images (Logic copied from Node, returns ALL images in table)
func GetAllImage(c *gin.Context) {
	if c.Request.Method == "HEAD" {
		c.Status(http.StatusMethodNotAllowed)
		return
	}

	if len(c.Request.URL.Query()) > 0 || c.Request.ContentLength > 0 || c.GetHeader("Authorization") != "" {
		c.Status(http.StatusBadRequest)
		return
	}

	// --- DB: Find All (Timer) ---
	startFind := time.Now()
	var images []models.Image
	db.DB.Find(&images)

	findDurationMs := float64(time.Since(startFind).Milliseconds())
	logs.Info("Query executed in " + strconv.FormatFloat(findDurationMs, 'f', 2, 64) + "ms")
	// metricsClient.Timing("db.query.latency.getAllImages", findDurationMs)

	c.JSON(http.StatusOK, images)
}

// DeleteImage handles deletion from S3 and DB
func DeleteImage(c *gin.Context) {
	authUserInterface, exists := c.Get("user")
	if !exists {
		c.Status(http.StatusUnauthorized)
		return
	}
	authUser := authUserInterface.(*models.User)

	productIdParam := c.Param("productId")
	imageIdParam := c.Param("imageId")

	pId, _ := strconv.Atoi(productIdParam)
	iId, _ := strconv.Atoi(imageIdParam)

	if len(c.Request.URL.Query()) > 0 || c.Request.ContentLength > 0 {
		c.Status(http.StatusBadRequest)
		return
	}

	// --- DB: Find Product (Timer) ---
	startFindProd := time.Now()
	var product models.Product
	if err := db.DB.First(&product, pId).Error; err != nil {
		c.Status(http.StatusNotFound) // Product must exist
		return
	}
	findProdDurationMs := float64(time.Since(startFindProd).Milliseconds())
	logs.Info("Query executed in " + strconv.FormatFloat(findProdDurationMs, 'f', 2, 64) + "ms")
	// metricsClient.Timing("db.query.latency.findProduct", findProdDurationMs)

	if product.OwnerUserID != authUser.ID {
		c.Status(http.StatusForbidden)
		return
	}

	// --- DB: Find Image (Timer) ---
	startFindImg := time.Now()
	var image models.Image
	if err := db.DB.Where("image_id = ? AND product_id = ?", iId, pId).First(&image).Error; err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	findImgDurationMs := float64(time.Since(startFindImg).Milliseconds())
	logs.Info("Query executed in " + strconv.FormatFloat(findImgDurationMs, 'f', 2, 64) + "ms")
	// metricsClient.Timing("db.query.latency.findImage", findImgDurationMs)

	// --- S3: Delete (Timer) ---
	startS3Del := time.Now()
	_, err := s3Client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(os.Getenv("S3_BUCKET_NAME")),
		Key:    aws.String(image.S3BucketPath),
	})
	if err != nil {
		logs.Error("Failed to delete image from S3: " + err.Error())
		c.Status(http.StatusServiceUnavailable)
		return
	}
	s3DelDurationMs := float64(time.Since(startS3Del).Milliseconds())
	logs.Info("S3 Delete executed in " + strconv.FormatFloat(s3DelDurationMs, 'f', 2, 64) + "ms")
	// metricsClient.Timing("s3.delete.latency", s3DelDurationMs)

	// --- DB: Delete Image (Timer) ---
	startDelDB := time.Now()
	if err := db.DB.Delete(&image).Error; err != nil {
		logs.Error("Failed to delete image record: " + err.Error())
		c.Status(http.StatusServiceUnavailable)
		return
	}
	delDBDurationMs := float64(time.Since(startDelDB).Milliseconds())
	logs.Info("Query executed in " + strconv.FormatFloat(delDBDurationMs, 'f', 2, 64) + "ms")
	// metricsClient.Timing("db.query.latency.deleteImage", delDBDurationMs)

	c.Status(http.StatusNoContent)
}
