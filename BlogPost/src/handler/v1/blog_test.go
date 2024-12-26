package v1_test

import (
	v1 "blog-post-task/src/handler/v1"
	"blog-post-task/src/models"
	"blog-post-task/src/utils/constants"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	host     = "127.0.0.1"
	port     = "3306"
	user     = "root"
	password = "root"
	dbName   = "blog_post"
)

func setupTestDB() (*gorm.DB, error) {
	var dsn string
	// Use the password if it is set
	if password != "" {
		dsn = user + ":" + password + "@(" + host + ":" + port + ")/" + dbName + "?parseTime=true"
	} else {
		// If no password, omit it from the DSN
		dsn = user + "@(" + host + ":" + port + ")/" + dbName + "?parseTime=true"
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err.Error(),
			"service": constants.ServiceName,
		}).Warn("failed to connect to database")
		return db, err
	}

	// Auto-migrate the necessary models
	db.AutoMigrate(&models.Article{})
	db.AutoMigrate(&models.Comment{})

	return db, nil
}

func TestGetAllArticles(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	d, _ := db.DB()
	defer d.Close()
	r := gin.Default()
	handler, err := v1.NewBlogPostHandler(db)
	assert.NoError(t, err)
	r.GET("/articles", handler.GetAllArticles)
	req, err := http.NewRequest("GET", "/articles", nil)
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	response := []models.Article{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	if reflect.DeepEqual([]models.Article{}, response) {
		t.Errorf("Error on response")
	}
}

func JSONStringify(data interface{}) string {
	byteData, _ := json.Marshal(data)
	stringData := string(byteData)
	return stringData
}

func TestPostArticle(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	d, _ := db.DB()
	defer d.Close()
	r := gin.Default()
	handler, err := v1.NewBlogPostHandler(db)
	assert.NoError(t, err)
	r.POST("/article", handler.PostArticle)
	PostArticle := models.Article{
		Title:        "Sample Article",
		Content:      "This is a sample article content.",
		Nickname:     "JohnDoe",
		CreationDate: time.Now(),
	}
	strData := JSONStringify(PostArticle)
	req, err := http.NewRequest("POST", "/article", strings.NewReader(strData))
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var response models.Article
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotNil(t, response.ArticleID)
	assert.Equal(t, PostArticle.Title, response.Title)
	assert.Equal(t, PostArticle.Content, response.Content)
	assert.Equal(t, PostArticle.Nickname, response.Nickname)
}

func TestGetArticle(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	d, _ := db.DB()
	defer d.Close()
	r := gin.Default()
	handler, err := v1.NewBlogPostHandler(db)
	assert.NoError(t, err)
	r.GET("/articles/:id", handler.GetArticle)
	req, err := http.NewRequest("GET", "/articles/1", nil)
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	response := models.Article{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	if reflect.DeepEqual(models.Article{}, response) {
		t.Errorf("Error on response")
	}
}
func TestAddComment(t *testing.T) {
	// Setup test database
	db, err := setupTestDB()
	assert.NoError(t, err)

	// Ensure that an article exists for adding a comment (seeding data)
	article := models.Article{
		Title:        "Sample Article",
		Content:      "This is a sample article content.",
		Nickname:     "JohnDoe",
		CreationDate: time.Now(),
	}
	err = db.Create(&article).Error
	assert.NoError(t, err)

	// Initialize Gin router and handler
	r := gin.Default()
	handler, err := v1.NewBlogPostHandler(db)
	assert.NoError(t, err)

	// Define the route
	r.POST("/article/comment", handler.AddComment)

	// Prepare the comment data
	addComment := models.Comment{
		ArticleID:       uint(article.ArticleID), // Ensure it's the valid ArticleID
		ParentCommentID: 4,                       // Assuming no parent comment for this test
		Content:         "This is a sample comment.",
		Nickname:        "JohnDoe",
		CreationDate:    time.Now(),
	}
	strData := JSONStringify(addComment)

	// Create a new POST request
	req, err := http.NewRequest("POST", "/article/comment", strings.NewReader(strData))
	assert.NoError(t, err)

	// Record the response
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Ensure the response status is OK
	assert.Equal(t, http.StatusOK, w.Code)

	// Unmarshal the response to check if the comment was added
	var response models.Comment
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Assert the response has the correct data
	assert.NotNil(t, response.ArticleID)
	assert.Equal(t, addComment.ArticleID, response.ArticleID)
	assert.Equal(t, addComment.ParentCommentID, response.ParentCommentID)
	assert.Equal(t, addComment.Content, response.Content)
	assert.Equal(t, addComment.Nickname, response.Nickname)

	// Optionally, check if the comment was saved in the DB
	var savedComment models.Comment
	err = db.First(&savedComment, "article_id = ? AND nickname = ?", addComment.ArticleID, addComment.Nickname).Error
	assert.NoError(t, err)
	assert.Equal(t, addComment.Content, savedComment.Content)
}

func TestGetArticleComments(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	d, _ := db.DB()
	defer d.Close()
	r := gin.Default()
	handler, err := v1.NewBlogPostHandler(db)
	assert.NoError(t, err)
	r.GET("/comments/:article_id", handler.GetArticleComments)
	req, err := http.NewRequest("GET", "/comments/1", nil)
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	response := []models.Comment{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	if reflect.DeepEqual([]models.Comment{}, response) {
		t.Errorf("Error on response")
	}
}

func TestGetAllComments(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)
	d, _ := db.DB()
	defer d.Close()
	r := gin.Default()
	handler, err := v1.NewBlogPostHandler(db)
	assert.NoError(t, err)
	r.GET("/comments", handler.GetAllComments)
	req, err := http.NewRequest("GET", "/comments", nil)
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	response := []models.Comment{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	if reflect.DeepEqual([]models.Comment{}, response) {
		t.Errorf("Error on response")
	}
}

func TestGetCommentOnComment(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)

	r := gin.Default()
	handler, err := v1.NewBlogPostHandler(db)
	assert.NoError(t, err)

	// Ensure the handler is working
	r.GET("/comments/:article_id/:id", handler.GetComentOnComment)

	// Assuming there's a comment on article 1 with ID 1 for this test
	req, err := http.NewRequest("GET", "/comments/6/4", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Ensure the status code is OK
	assert.Equal(t, http.StatusOK, w.Code)

	// Unmarshal the response body into a slice of comments
	var response []models.Comment
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Make sure the response isn't empty (assuming data exists)
	if len(response) == 0 {
		t.Errorf("Expected comments, but got none")
	}
}
