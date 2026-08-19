package middleware

import (
	"bytes"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type IdempotencyRecord struct {
	Key        string    `gorm:"type:varchar(255);primaryKey"`
	StatusCode int       `gorm:"not null"`
	Response   string    `gorm:"type:text;not null"`
	CreatedAt  time.Time
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func IdempotencyMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodPost && c.Request.Method != http.MethodPut && c.Request.Method != http.MethodPatch {
			c.Next()
			return
		}

		key := c.GetHeader("Idempotency-Key")
		if key == "" {
			c.Next()
			return
		}

		var record IdempotencyRecord
		if err := db.First(&record, "key = ?", key).Error; err == nil {
			c.Header("Content-Type", "application/json")
			c.String(record.StatusCode, record.Response)
			c.Abort()
			return
		}

		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		c.Next()

		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			db.Create(&IdempotencyRecord{
				Key:        key,
				StatusCode: c.Writer.Status(),
				Response:   blw.body.String(),
				CreatedAt:  time.Now(),
			})
		}
	}
}
