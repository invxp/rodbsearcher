package rodbsearcher

import (
	"fmt"
	"github.com/gin-gonic/gin"
	internal "github.com/invxp/rodbsearcher/internal/http"
	"github.com/invxp/rodbsearcher/internal/util/system"
	"log"
	"net/http"
	"time"
)

//logf 打印日志(如果没有启用则打到控制台)
func (searcher *RODBSearcher) logf(format string, v ...interface{}) {
	if searcher.logger == nil {
		log.Printf(format, v...)
	} else {
		searcher.logger.Printf(format, v...)
	}
}

func (searcher *RODBSearcher) panic(v interface{}) {
	panic(v)
}

//httpStatics HTTP统计执行时间(中间件)
func (searcher *RODBSearcher) httpStatics() gin.HandlerFunc {
	return func(c *gin.Context) {
		lastTime := time.Now()
		c.Next()
		searcher.logf("HTTP %d %s@%s: %s - LATENCY: %v, HEADERS: %v", c.Writer.Status(), c.ClientIP(), c.Request.Method, c.FullPath(), time.Since(lastTime), c.Request.Header)
	}
}

//httpAuth HTTP校验是否非法(中间件）
func (searcher *RODBSearcher) httpAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("Auth") == "FALSE" {
			c.AbortWithStatusJSON(http.StatusBadRequest, internal.Response{Code: internal.StatusCodeAuthFailed, Description: "Auth Failed", MessageID: system.UniqueID()})
		}
		c.Next()
	}
}

//crateService 随机出错(只是个测试)
func (searcher *RODBSearcher) createService(serviceName, modPath, installDir string) error {
	if time.Now().Unix()%2 == 0 {
		return fmt.Errorf("create %s.%s-%s error", modPath, serviceName, installDir)
	}
	return nil
}
