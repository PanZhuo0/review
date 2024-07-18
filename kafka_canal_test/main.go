package main

import (
	"context"
	"fmt"

	"github.com/segmentio/kafka-go"
)

func main() {
	// 测试Canal是否将MySQL中的Bin日志发往了Kafka
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{`localhost:9092`},
		Topic:     `example`,
		Partition: 0,
	})
	for {
		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			panic(err)
		}
		fmt.Printf("OFFSET:%v Key:%v Value:%v", m.Offset, string(m.Key), string(m.Value))
	}
}
