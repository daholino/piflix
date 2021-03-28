package internal

import (
	"io"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func SetupLogger(path string) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file for logger: %v", err)
	}

	wrt := io.MultiWriter(os.Stdout, file)
	log.SetOutput(wrt)

	gin.DisableConsoleColor()
	gin.DefaultWriter = wrt
}
