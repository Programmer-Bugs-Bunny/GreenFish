//go:build ignore

package main

import (
	"io"
	"log"
	"os"

	"go-web-template/models"

	"ariga.io/atlas-provider-gorm/gormschema"
)

func main() {
	stmts, err := gormschema.New("postgres").Load(
		&models.User{},
		&models.Role{},
		// 在这里添加其他模型...
		// &models.Product{},
		// &models.Order{},
	)
	if err != nil {
		log.Fatalf("failed to load gorm schema: %v", err)
	}
	io.WriteString(os.Stdout, stmts)
}
