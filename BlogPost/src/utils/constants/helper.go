package constants

import (
	"blog-post-task/src/models"
	"errors"
)

// ValidateArticle checks if the required fields are non-empty
func ValidateArticle(article *models.Article) error {
	if article.Title == "" {
		return errors.New("title cannot be empty")
	}
	if article.Nickname == "" {
		return errors.New("nickname cannot be empty")
	}
	// For CreationDate, we assume that zero time means an invalid date
	if article.CreationDate.IsZero() {
		return errors.New("CreationDate cannot be empty")
	}
	return nil
}
