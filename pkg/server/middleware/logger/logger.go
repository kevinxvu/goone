package logger

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/vuduongtp/go-core/pkg/logging"
	"go.uber.org/zap"
)

// Header keys
const (
	CorrelationIDHeaderKey = "X-Correlation-ID"
	RequestIDHeaderKey     = "X-Request-ID"
	UserInfoHeaderKey      = "X-User-Info"
)

// Context keys
type contextKey string

const (
	CorrelationIDKey contextKey = "correlation_id"
	RequestIDKey     contextKey = "request_id"
	UserInfoKey      contextKey = "user_info"
)

// setEchoContext sets correlation ID, request ID and user info in context
func setEchoContext(c echo.Context) {
	var correlationID, requestID, userInfo string
	ctx := c.Request().Context()

	// Get or generate correlation ID
	correlationID = c.Request().Header.Get(CorrelationIDHeaderKey)
	if correlationID == "" {
		correlationID = generateCorrelationID()
	}

	// Get or generate request ID
	requestID = c.Request().Header.Get(RequestIDHeaderKey)
	if requestID == "" {
		requestID = uuid.NewString()
	}

	// Get user info if available
	userInfo = c.Request().Header.Get(UserInfoHeaderKey)

	// Create logger with request context
	logger := logging.DefaultLogger().With(
		zap.String("correlation_id", correlationID),
		zap.String("request_id", requestID),
	)
	if userInfo != "" {
		logger = logger.With(zap.String("user_info", userInfo))
	}

	// Store logger in context
	ctx = logging.WithContext(ctx, logger)

	// Update request with new context
	c.SetRequest(c.Request().WithContext(ctx))

	// Set response headers
	c.Response().Header().Set(CorrelationIDHeaderKey, correlationID)
	c.Response().Header().Set(RequestIDHeaderKey, requestID)
}

// generateCorrelationID generates a new correlation ID
func generateCorrelationID() string {
	return uuid.NewString()
}

// Middleware returns a middleware that logs HTTP requests using zap logger
func Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()

			// Set context with correlation ID, request ID
			setEchoContext(c)

			start := time.Now()
			var err error
			var errStr string

			// Execute the handler
			if err = next(c); err != nil {
				c.Error(err)
				b, _ := json.Marshal(err.Error())
				if len(b) > 1 {
					b = b[1 : len(b)-1]
				}
				errStr = string(b)
			}

			stop := time.Now()
			latency := stop.Sub(start)

			// Get request size
			reqSizeStr := req.Header.Get(echo.HeaderContentLength)
			if reqSizeStr == "" {
				reqSizeStr = "0"
			}
			reqSize, _ := strconv.ParseInt(reqSizeStr, 10, 64)

			// Get logger from context
			logger := logging.FromContext(c.Request().Context())

			// Build log fields
			fields := []zap.Field{
				zap.String("ip", c.RealIP()),
				zap.String("user_agent", req.UserAgent()),
				zap.String("host", req.Host),
				zap.String("method", req.Method),
				zap.String("path", req.URL.Path),
				zap.String("uri", req.RequestURI),
				zap.Int("status", res.Status),
				zap.Int64("byte_in", reqSize),
				zap.Int64("byte_out", res.Size),
				zap.String("latency", latency.String()),
				zap.Int64("latency_ms", latency.Milliseconds()),
				zap.String("referer", req.Referer()),
			}

			// Log request
			if !strings.EqualFold(errStr, "") {
				fields = append(fields, zap.String("error", errStr))
				logger.Error("HTTP request", fields...)
			} else {
				logger.Info("HTTP request", fields...)
			}

			return err
		}
	}
}
