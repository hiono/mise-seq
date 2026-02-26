package main

import (
	"context"
	"fmt"
	"log"

	"github.com/mise-seq/config-loader/config"
	"github.com/mise-seq/config-loader/mise"
)

func main() {
	ctx := context.Background()

	loader := config.NewLoader()
	cfg, err := loader.Parse("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	miseClient := mise.NewClient()

	fmt.Println("Installing tools with hooks...")
	if err := miseClient.InstallAllWithHooks(ctx, cfg); err != nil {
		log.Fatalf("Installation failed: %v", err)
	}

	fmt.Println("Done!")
}
