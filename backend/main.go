package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"fn-music-dl/api"
	"fn-music-dl/pkg"

	"github.com/gin-gonic/gin"
)

func main() {
	pkg.CM.Load()

	socketPath := strings.TrimSpace(os.Getenv("SOCKET_PATH"))
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(corsMiddleware())

	api.RegisterRoutes(r)

	if socketPath != "" {
		// Unix socket mode — used by FNOS unified gateway
		if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
			log.Fatalf("failed to remove existing socket %s: %v", socketPath, err)
		}
		listener, err := net.Listen("unix", socketPath)
		if err != nil {
			log.Fatalf("failed to listen on unix socket %s: %v", socketPath, err)
		}
		defer os.Remove(socketPath)
		// Make socket accessible by the gateway
		_ = os.Chmod(socketPath, 0777)
		fmt.Printf("Listening on unix socket %s\n", socketPath)
		if err := r.RunListener(listener); err != nil {
			log.Fatalf("server error: %v", err)
		}
	} else {
		// TCP mode — local development
		listenAddr := ":" + port
		fmt.Printf("Listening on http://localhost%s\n", listenAddr)
		if err := r.Run(listenAddr); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Cache-Control, Content-Language, Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
