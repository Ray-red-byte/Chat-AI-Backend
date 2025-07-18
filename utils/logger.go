package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// Kafka producer singleton
var kafkaProd *kafka.Producer

type StructuredLogger struct {
	producer *kafka.Producer
	topic    string
}

var Logger *StructuredLogger

func init() {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "kafka:9092",
	})
	if err != nil {
		log.Fatalf("❌ Failed to create Kafka producer: %v", err)
	}

	go func() {
		for e := range producer.Events() {
			if m, ok := e.(*kafka.Message); ok && m.TopicPartition.Error != nil {
				log.Printf("Kafka delivery error: %v", m.TopicPartition.Error)
			}
		}
	}()

	Logger = &StructuredLogger{
		producer: producer,
		topic:    "chatai-logs",
	}
}

func (l *StructuredLogger) Info(format string, args ...any) {
	l.send("INFO", format, args...)
}

func (l *StructuredLogger) Warn(format string, args ...any) {
	l.send("WARN", format, args...)
}

func (l *StructuredLogger) Error(format string, args ...any) {
	l.send("ERROR", format, args...)
}

func (l *StructuredLogger) Println(args ...any) {
	msg := fmt.Sprintln(args...)
	l.send("INFO", "%s", msg)
}

func (l *StructuredLogger) send(level, format string, args ...any) {
	logMsg := map[string]interface{}{
		"level":     level,
		"timestamp": time.Now().Format(time.RFC3339),
		"message":   fmt.Sprintf(format, args...),
	}

	// 取得呼叫端檔案與行數
	if _, file, line, ok := runtime.Caller(3); ok {
		logMsg["caller"] = fmt.Sprintf("%s:%d", filepath.Base(file), line)
	}

	data, err := json.Marshal(logMsg)
	if err != nil {
		log.Printf("Failed to marshal log: %v", err)
		return
	}

	err = l.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &l.topic,
			Partition: kafka.PartitionAny,
		},
		Value: data,
	}, nil)
	if err != nil {
		log.Printf("Failed to send log to Kafka: %v", err)
	}
}
