package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"my-project/db"
	"my-project/logs"
	"my-project/models"
)

// --- AWS Client Initialization ---
var snsClient *sns.Client
var ddbClient *dynamodb.Client

func init() {
	// Initialize AWS Clients lazily or on startup
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		logs.Error("Unable to load SDK config, " + err.Error())
	} else {
		snsClient = sns.NewFromConfig(cfg)
		ddbClient = dynamodb.NewFromConfig(cfg)
	}
}

// --- Helper Functions ---

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 16)
	return string(bytes), err
}

func validateEmail(email string) bool {
	// Simple Regex for email validation
	regex := `^[^\s@]+@[^\s@]+\.[^\s@]+$`
	matched, _ := regexp.MatchString(regex, strings.ToLower(email))
	return matched
}

// --- Request Structs ---

type CreateUserRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Password  string `json:"password" binding:"required,min=8,max=12"`
	Username  string `json:"username" binding:"required"` // This is the email
}

type UpdateUserRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Password  string `json:"password" binding:"required"`
}

type DynamoVerifyItem struct {
	Email string `dynamodbav:"email"`
	Token string `dynamodbav:"token"`
	TTL   int64  `dynamodbav:"ttl"`
}

// --- Controllers ---

// VerifyEmail handles the email verification logic via DynamoDB
func VerifyEmail(c *gin.Context) {
	email := c.Query("email")
	token := c.Query("token")

	if email == "" || token == "" {
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte("<h1>Error</h1><p>Email and token are required.</p>"))
		return
	}

	tableName := os.Getenv("DDB_VERIFY_TABLE")

	// Get Item from DynamoDB
	getItemInput := &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"email": &types.AttributeValueMemberS{Value: email},
		},
	}

	result, err := ddbClient.GetItem(context.TODO(), getItemInput)
	if err != nil || result.Item == nil {
		logs.Warn("Verification attempt for invalid email: " + email)
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte("<h1>Error</h1><p>Invalid or expired verification link.</p>"))
		return
	}

	var record DynamoVerifyItem
	err = attributevalue.UnmarshalMap(result.Item, &record)
	if err != nil {
		logs.Error("Failed to unmarshal DynamoDB item: " + err.Error())
		c.Data(http.StatusInternalServerError, "text/html; charset=utf-8", []byte("<h1>Error</h1><p>Internal server error.</p>"))
		return
	}

	if record.Token != token {
		logs.Warn("Invalid token for email: " + email)
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte("<h1>Error</h1><p>Invalid or expired verification link.</p>"))
		return
	}

	if record.TTL < time.Now().Unix() {
		logs.Warn("Expired token for email: " + email)
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte("<h1>Error</h1><p>Verification link has expired. Please register again.</p>"))
		return
	}

	// Delete Item from DynamoDB
	deleteItemInput := &dynamodb.DeleteItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"email": &types.AttributeValueMemberS{Value: email},
		},
	}
	_, err = ddbClient.DeleteItem(context.TODO(), deleteItemInput)
	if err != nil {
		logs.Error("Error deleting token from DynamoDB: " + err.Error())
	}

	logs.Info("Successfully verified email: " + email)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("<h1>Success!</h1><p>Email verified successfully! You can now log in.</p>"))
}

// CreateUser handles user registration
func CreateUser(c *gin.Context) {
	// Validate Headers
	if c.Request.ContentLength == 0 && c.Request.Header.Get("Transfer-Encoding") == "" {
		c.Status(http.StatusBadRequest)
		return
	}
	if len(c.Request.URL.Query()) > 0 || c.GetHeader("Authorization") != "" {
		c.Status(http.StatusBadRequest)
		return
	}

	var req CreateUserRequest
	// Strict JSON unmarshalling
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// Validate Email
	if !validateEmail(req.Username) {
		logs.Info("Invalid Email Address")
		c.Status(http.StatusBadRequest)
		return
	}

	// --- DB: Find User (Timer) ---
	startFind := time.Now()
	var existingUser models.User
	result := db.DB.Where("username = ?", req.Username).First(&existingUser)

	// Calculate Metrics
	findDurationMs := float64(time.Since(startFind).Milliseconds())
	logs.Info("Query executed in " + strconv.FormatFloat(findDurationMs, 'f', 2, 64) + "ms")
	logs.Client.Timing("db.query.latency.findUser", findDurationMs) // Uncomment when metrics package is ready

	if result.RowsAffected > 0 {
		c.Status(http.StatusBadRequest) // User already exists
		return
	}

	hashedPassword, _ := hashPassword(req.Password)

	newUser := models.User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Password:  hashedPassword,
		Username:  strings.ToLower(req.Username),
	}

	// --- DB: Insert User (Timer) ---
	startInsert := time.Now()
	if err := db.DB.Create(&newUser).Error; err != nil {
		logs.Error("User insert failed: " + err.Error())
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "23505") {
			c.Status(http.StatusBadRequest)
		} else {
			c.Status(http.StatusServiceUnavailable)
		}
		return
	}
	insertDurationMs := float64(time.Since(startInsert).Milliseconds())
	logs.Info("Query executed in " + strconv.FormatFloat(insertDurationMs, 'f', 2, 64) + "ms")
	logs.Client.Timing("db.query.latency.createUser", insertDurationMs)

	// --- SNS Publish ---
	if os.Getenv("GO_ENV") != "test" {
		snsMessage := map[string]string{
			"email":      newUser.Username,
			"first_name": newUser.FirstName,
		}
		msgBytes, _ := json.Marshal(snsMessage)
		msgString := string(msgBytes)

		startSNS := time.Now()
		_, err := snsClient.Publish(context.TODO(), &sns.PublishInput{
			TopicArn: aws.String(os.Getenv("SNS_TOPIC_ARN")),
			Message:  aws.String(msgString),
		})
		snsDuration := time.Since(startSNS).Milliseconds()

		if err != nil {
			logs.Error("SNS Publish failed: " + err.Error())
		} else {
			logs.Info("Successfully published registration message for " + newUser.Username + " to SNS.")
			logs.Client.Timing("sns.publish.latency", float64(snsDuration))
		}
	}

	c.JSON(http.StatusCreated, newUser)
}

// GetUser retrieves user details
func GetUser(c *gin.Context) {
	if c.Request.Method == "HEAD" {
		c.Status(http.StatusMethodNotAllowed)
		return
	}

	// Authorization Check (Assuming Middleware sets "user")
	authUserInterface, exists := c.Get("user")
	if !exists {
		c.Status(http.StatusUnauthorized)
		return
	}
	authUser := authUserInterface.(*models.User)

	userIdParam := c.Param("userId") // In Go router, usually :userID
	userIdInt, err := strconv.ParseUint(userIdParam, 10, 32)

	// Validation
	if err != nil || len(c.Request.URL.Query()) > 0 || c.Request.ContentLength > 0 {
		c.Status(http.StatusBadRequest)
		return
	}

	if uint(userIdInt) != authUser.ID {
		c.Status(http.StatusForbidden)
		return
	}

	// Return User
	c.JSON(http.StatusOK, authUser)
}

// UpdateUser handles user updates
func UpdateUser(c *gin.Context) {
	authUserInterface, exists := c.Get("user")
	if !exists {
		c.Status(http.StatusUnauthorized)
		return
	}
	authUser := authUserInterface.(*models.User)

	userIdParam := c.Param("userId")
	userIdInt, err := strconv.ParseUint(userIdParam, 10, 32)

	if err != nil || len(c.Request.URL.Query()) > 0 || c.Request.ContentLength == 0 {
		c.Status(http.StatusBadRequest)
		return
	}

	if uint(userIdInt) != authUser.ID {
		c.Status(http.StatusForbidden)
		return
	}

	var req UpdateUserRequest
	// Go's strict typing handles the "unexpected keys" check automatically via JSON binding
	// if we set DisallowUnknownFields. However, standard BindJSON ignores extras.
	// To match your strict "400 on unexpected fields" logic, we use a custom decoder:

	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields() // Crucial to match your Node logic
	if err := decoder.Decode(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	newPassword, _ := hashPassword(req.Password)

	// --- DB: Update User (Timer) ---
	startUpdate := time.Now()

	updates := map[string]interface{}{
		"first_name":      req.FirstName,
		"last_name":       req.LastName,
		"password":        newPassword,
		"account_updated": time.Now(),
	}

	if err := db.DB.Model(&models.User{}).Where("id = ?", authUser.ID).Updates(updates).Error; err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	updateDurationMs := float64(time.Since(startUpdate).Milliseconds())
	logs.Info("Query executed in " + strconv.FormatFloat(updateDurationMs, 'f', 2, 64) + "ms")
	logs.Client.Timing("db.query.latency.updateUser", updateDurationMs)

	c.Status(http.StatusNoContent)
}
