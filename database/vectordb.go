package database

import (
	"chat-app/config"
	"context"
	"fmt"
	"log"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
)

var (
	VectorDB client.Client
)

func InitMilvus() {
	milvusAddress := fmt.Sprintf("%s:%d", config.AppConfig.MilvusHost, config.AppConfig.MilvusPort)

	var err error
	VectorDB, err = client.NewClient(
		context.Background(),
		client.Config{
			Address: milvusAddress,
		},
	)
	if err != nil {
		log.Fatalf("Failed to connect to Milvus: %v", err)
	}

	log.Printf("Successfully connected to Milvus at %s\n", milvusAddress)
}

// CloseMilvus closes the connection to Milvus
func CloseMilvus() {
	if VectorDB == nil {
		return
	}

	if err := VectorDB.Close(); err != nil {
		log.Printf("Error closing Milvus connection: %v\n", err)
	} else {
		log.Println("Milvus connection closed successfully.")
	}
}
