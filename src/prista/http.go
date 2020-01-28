package prista

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// initialize and start HTTP server
func initHttpServer(wg *sync.WaitGroup) bool {
	listenPort := AppConfig.GetInt32("server.http.listen_port", 0)
	if listenPort <= 0 {
		log.Println("No valid [server.http.listen_port] configured, HTTP server is disabled.")
		return false
	}
	listenAddr := AppConfig.GetString("server.http.listen_addr", "127.0.0.1")
	e := echo.New()

	requestTimeout := AppConfig.GetTimeDuration("server.request_timeout", time.Duration(0))
	if requestTimeout > 0 {
		e.Server.ReadTimeout = requestTimeout
	}

	bodyLimit := AppConfig.GetByteSize("server.max_request_size")
	if bodyLimit != nil && bodyLimit.Int64() > 0 {
		e.Use(middleware.BodyLimit(bodyLimit.String()))
	}

	e.POST("/api/log", httpHandlerLog)
	e.PUT("/api/log", httpHandlerLog)

	log.Printf("Starting [%s] HTTP server on [%s:%d]...\n", AppConfig.GetString("app.name")+" v"+AppConfig.GetString("app.version"), listenAddr, listenPort)
	go func() {
		err := e.Start(fmt.Sprintf("%s:%d", listenAddr, listenPort))
		if err != nil {
			log.Println(err)
		}
		wg.Done()
	}()
	return true
}

func extractString(source map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if v, ok := source[key]; ok {
			switch v.(type) {
			case string:
				return strings.TrimSpace(v.(string))
			}
		}
	}
	return ""
}

func httpHandlerLog(c echo.Context) error {
	requestBodyData := map[string]interface{}{}
	if err := c.Bind(&requestBodyData); err != nil {
		log.Printf(fmt.Sprintf("Error while parsing request body as Json: %e", err))
		return c.HTML(http.StatusBadRequest, err.Error())
	}
	category := extractString(requestBodyData, "category", "cat", "c")
	message := extractString(requestBodyData, "message", "msg", "m")
	if category == "" || message == "" {
		return c.HTML(http.StatusBadRequest, "Missing parameter [category] and/or [message]")
	}
	payload := strings.ToLower(category) + "\t" + message
	if err := handleIncomingMessage([]byte(payload)); err != nil {
		return c.HTML(http.StatusInternalServerError, err.Error())
	}
	return c.HTML(http.StatusOK, "Ok")
}
