package job

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"review_job/internal/conf"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/segmentio/kafka-go"
)

// 评价数据流处理任务

// JOB自定义执行JOB结构体,实现Transport.Server接口
type JobWorker struct {
	kafkaReader *kafka.Reader
	esClient    *ESClient
	log         *log.Helper
}

type ESClient struct {
	client *elasticsearch.TypedClient
	index  string
}

func NewJobWorker(k *kafka.Reader, e *ESClient, logger log.Logger) *JobWorker {
	return &JobWorker{
		kafkaReader: k,
		esClient:    e,
		log:         log.NewHelper(logger),
	}
}

func NewKafkaReader(cfg *conf.Kafka) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers: cfg.GetBrokers(),
		GroupID: cfg.GetGroupId(),
		Topic:   cfg.GetTopic(),
	})
}

func NewESClient(cfg *conf.ES) (*ESClient, error) {
	client, err := elasticsearch.NewTypedClient(elasticsearch.Config{
		Addresses: cfg.GetAddresses(),
	})
	if err != nil {
		return nil, err
	}
	return &ESClient{
		client: client,
		index:  cfg.GetIndex(),
	}, nil
}

// KAFKA 从Canal中收到的消息
type Msg struct {
	Type     string                   `json:"type"`
	Database string                   `json:"database"`
	Table    string                   `json:"table"`
	IsDDL    bool                     `json:"isDdl"`
	Data     []map[string]interface{} `json:"data"`
}

// Start kratos程序启动之后会调用的方法
// ctx 是kratos框架启动的时候传入的ctx，是带有退出取消的
func (jw JobWorker) Start(ctx context.Context) error {
	jw.log.Debug("JobWorker start....")
	// 1. 从kafka中获取MySQL中的数据变更消息
	// 接收消息
	for {
		m, err := jw.kafkaReader.ReadMessage(ctx)
		if errors.Is(err, context.Canceled) {
			return nil
		}
		if err != nil {
			jw.log.Errorf("readMessage from kafka failed, err:%v", err)
			break
		}
		// jw.log.Debugf("message at topic/partition/offset %v/%v/%v: %s = %s\n", m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))
		// 2. 将完整评价数据写入ES
		msg := new(Msg)
		if err := json.Unmarshal(m.Value, msg); err != nil {
			jw.log.Errorf("unmarshal msg from kafka failed, err:%v", err)
			continue
		}

		// 补充！
		// 实际的业务场景可能需要在这增加一个步骤：对数据做业务处理
		// 例如：把两张表的数据合成一个文档写入ES

		if msg.Type == "INSERT" {
			// 往ES中新增文档
			for idx := range msg.Data {
				jw.indexDocument(msg.Data[idx])
			}
		} else {
			// 往ES中更新文档
			for idx := range msg.Data {
				jw.updateDocument(msg.Data[idx])
			}
		}
	}

	return nil
}

// Stop kratos结束之后会调用的
func (jw JobWorker) Stop(context.Context) error {
	jw.log.Debug("JobWorker stop....")
	// 程序退出前关闭Reader
	return jw.kafkaReader.Close()
}

// indexDocument 索引文档
func (jw JobWorker) indexDocument(d map[string]interface{}) {
	reviewID := d["review_id"].(string)
	// 添加文档
	resp, err := jw.esClient.client.Index(jw.esClient.index).
		Id(reviewID).
		Document(d).
		Do(context.Background())
	if err != nil {
		jw.log.Errorf("indexing document failed, err:%v\n", err)
		return
	}
	jw.log.Debugf("result:%#v\n", resp.Result)
}

// updateDocument 更新文档
func (jw JobWorker) updateDocument(d map[string]interface{}) {
	fmt.Println(d)
	reviewID := d["review_id"].(string)
	fmt.Println(reviewID)
	resp, err := jw.esClient.client.Update(jw.esClient.index, reviewID).
		Doc(d). // 使用结构体变量更新
		Do(context.Background())
	if err != nil {
		jw.log.Errorf("update document failed, err:%v\n", err)
		return
	}
	jw.log.Debugf("result:%v\n", resp.Result)
}
