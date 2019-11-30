package ginzap

import (
	"bytes"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	counter uint64
	bpool   = &bytesPool{
		pool: sync.Pool{New: func() interface{} {
			return bytes.NewBuffer(nil)
		}},
	}
)

type (
	// hijackWriter 通过代理模式拦截
	hijackWriter struct {
		gin.ResponseWriter
		cache *bytes.Buffer
	}

	bytesPool struct {
		pool sync.Pool
	}
)

func (bp *bytesPool) get() *bytes.Buffer {
	return bp.pool.Get().(*bytes.Buffer)
}

func (bp *bytesPool) put(cache *bytes.Buffer) {
	cache.Reset()
	bp.pool.Put(cache)
}

func (hw *hijackWriter) Write(b []byte) (int, error) {
	hw.cache.Write(b)
	return hw.ResponseWriter.Write(b)
}

// Ginzap returns a gin.HandlerFunc (middleware) that logs requests using uber-go/zap.
//
// Requests with errors are logged using zap.Error().
// Requests without errors are logged using zap.Info().
//
// It receives:
//   1. A time package format string (e.g. time.RFC3339).
//   2. A boolean stating whether to use UTC time zone or local.
func Ginzap(logger *zap.Logger, mergeLog bool, dumpResponse bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		index := atomic.AddUint64(&counter, 1)
		start := time.Now()
		// 避免被其他中间件修改
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		if !mergeLog {
			logger.Info("request-"+strconv.FormatUint(index, 10),
				zap.String("method", c.Request.Method),
				zap.String("path", path),
				zap.String("query", query),
				zap.String("ip", c.ClientIP()),
				zap.String("user-agent", c.Request.UserAgent()),
			)
		}

		if dumpResponse {
			cache := bpool.get()
			defer bpool.put(cache)
			c.Writer = &hijackWriter{cache: cache, ResponseWriter: c.Writer}
		}

		c.Next()

		latency := time.Since(start).Milliseconds()
		var respBody []byte

		if dumpResponse {
			if c.Writer.Size() > 0 {
				respBody = c.Writer.(*hijackWriter).cache.Bytes()
			}
		}

		switch {
		case len(c.Errors) > 0:
			for _, e := range c.Errors.Errors() {
				logger.Error(e, zap.Uint64("index", index))
			}

		case !mergeLog && !dumpResponse:
			logger.Info("response-"+strconv.FormatUint(index, 10),
				zap.Int("status-code", c.Writer.Status()),
				zap.Int64("latency", latency),
			)

		case !mergeLog && dumpResponse:
			logger.Info("response-"+strconv.FormatUint(index, 10),
				zap.Int("status", c.Writer.Status()),
				zap.ByteString("response-body", respBody),
				zap.Int64("latency", latency),
			)

		case mergeLog && !dumpResponse:
			logger.Info("request-"+strconv.FormatUint(index, 10),
				zap.String("method", c.Request.Method),
				zap.String("path", path),
				zap.String("query", query),
				zap.String("ip", c.ClientIP()),
				zap.String("user-agent", c.Request.UserAgent()),
				zap.Int("status", c.Writer.Status()),
				zap.Int64("latency", latency),
			)

		case mergeLog && dumpResponse:
			logger.Info("request-"+strconv.FormatUint(index, 10),
				zap.String("method", c.Request.Method),
				zap.String("path", path),
				zap.String("query", query),
				zap.String("ip", c.ClientIP()),
				zap.String("user-agent", c.Request.UserAgent()),
				zap.Int("status", c.Writer.Status()),
				zap.ByteString("response-body", respBody),
				zap.Int64("latency", latency),
			)
		}
	}
}

// RecoveryWithZap returns a gin.HandlerFunc (middleware)
// that recovers from any panics and logs requests using uber-go/zap.
// All errors are logged using zap.Error().
// stack means whether output the stack info.
// The stack info is easy to find where the error occurs but the stack info is too large.
func RecoveryWithZap(logger *zap.Logger, stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					logger.Error(c.Request.URL.Path,
						zap.Any("error", err),
						zap.ByteString("request", httpRequest),
					)
					// If the connection is dead, we can't write a status to it.
					c.Error(err.(error)) // nolint: errcheck
					c.Abort()
					return
				}

				if stack {
					logger.Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.ByteString("request", httpRequest),
						zap.ByteString("stack", debug.Stack()),
					)
				} else {
					logger.Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.ByteString("request", httpRequest),
					)
				}
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()

		c.Next()
	}
}
