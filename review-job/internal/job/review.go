package job

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	stdlog "log"
	"review-job/internal/conf"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	kratoslog "github.com/go-kratos/kratos/v2/log"
	"github.com/segmentio/kafka-go"
)

//评价数据流处理任务

//1.从kafka中获取评价数据
//2.将数据写入es中

// 定义一个JobWorker结构体，用于处理评价数据流
type JobWorker struct {
	kafkaReader *kafka.Reader
	ESClient    *ESClient
	Logger      *kratoslog.Helper
}

type ESClient struct {
	*elasticsearch.TypedClient
	index string
}

// 定义数据（与 Canal 等 binlog 同步格式一致：data 为行对象数组）
type Msg struct {
	Type     string        `json:"type"`
	Database string        `json:"database"`
	Table    string        `json:"table"`
	IsDDl    bool          `json:"isDdl"`
	Data     []interface{} `json:"data"`
}

func NewJobWorker(kafkaReader *kafka.Reader, ESClient *ESClient, logger *kratoslog.Helper) *JobWorker {
	return &JobWorker{
		kafkaReader: kafkaReader,
		ESClient:    ESClient,
		Logger:      logger,
	}
}

func NewKafkaReader(cfg *conf.Kafka) (*kafka.Reader, error) {
	if cfg == nil {
		return nil, fmt.Errorf("kafka config is nil")
	}
	if len(cfg.Brokers) == 0 {
		return nil, fmt.Errorf("kafka brokers is empty")
	}
	if cfg.Topic == "" {
		return nil, fmt.Errorf("kafka topic is empty")
	}
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Brokers,
		GroupID:  cfg.GroupId,
		Topic:    cfg.Topic,
		MaxBytes: 10e6, // 10MB
	}), nil
}

func NewESClient(cfg *conf.Elasticsearch) (*ESClient, error) {
	if cfg == nil {
		return nil, fmt.Errorf("elasticsearch config is nil")
	}
	if len(cfg.Addresses) == 0 {
		return nil, fmt.Errorf("elasticsearch addresses is empty")
	}
	client, err := elasticsearch.NewTypedClient(elasticsearch.Config{
		Addresses: cfg.Addresses,
	})
	if err != nil {
		return nil, err
	}
	return &ESClient{
		TypedClient: client,
		index:       cfg.Index,
	}, nil
}
// 程序启动时执行
func (j *JobWorker) Start(ctx context.Context) error {
	j.Logger.Debug("start job worker")
	for {
		m, err := j.kafkaReader.ReadMessage(ctx)
		if errors.Is(err, context.Canceled) {
			break
		}
		if err != nil {
			j.Logger.Error("failed to read message from kafka", "error", err)
			break
		}
		j.Logger.Debug("message at topic/partition/offset %v/%v/%v: %s = %s\n", m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))

		msg := new(Msg)
		if err := json.Unmarshal(m.Value, msg); err != nil {
			j.Logger.Error("failed to unmarshal message from kafka", "error", err)
			continue
		}
		if msg.Type == "INSERT" {
			for i, item := range msg.Data {
				row, ok := item.(map[string]interface{})
				if !ok {
					j.Logger.Error("kafka data row is not an object", "index", i)
					continue
				}
				j.indexDocument(row)
			}
		} else {
			for i, item := range msg.Data {
				row, ok := item.(map[string]interface{})
				if !ok {
					j.Logger.Error("kafka data row is not an object", "index", i)
					continue
				}
				j.updateDocument(row)
			}
		}
	}
	return nil
}

// kratos程序停止时执行
func (j *JobWorker) Stop(ctx context.Context) error {
	j.Logger.Debug("stop job worker")
	if err := j.kafkaReader.Close(); err != nil {
		return err
	}
	return nil
}
func ReadReviewFromKafka() { // 创建一个reader，指定GroupID，从 topic-A 消费消息
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{"localhost:9092", "localhost:9093", "localhost:9094"},
		GroupID:  "consumer-group-id", // 指定消费者组id
		Topic:    "topic-A",
		MaxBytes: 10e6, // 10MB
	})

	// 接收消息
	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			break
		}
		fmt.Printf("message at topic/partition/offset %v/%v/%v: %s = %s\n", m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))
	}

	// 程序退出前关闭Reader
	if err := r.Close(); err != nil {
		stdlog.Fatal("failed to close reader:", err)
	}
}

func conElasticsearch() *elasticsearch.TypedClient { // ES 配置
	cfg := elasticsearch.Config{
		Addresses: []string{
			"http://localhost:9200",
		},
	} 

	// 创建客户端连接
	client, err := elasticsearch.NewTypedClient(cfg)
	if err != nil {
		stdlog.Fatal("failed to create elasticsearch client:", err)
	}
	return client
}

// createIndex 创建索引
func createIndex(client *elasticsearch.TypedClient) {
	resp, err := client.Indices.
		Create("my-review-1").
		Do(context.Background())
	if err != nil {
		fmt.Printf("create index failed, err:%v\n", err)
		return
	}
	fmt.Printf("index:%#v\n", resp.Index)
}

// updateDocument 更新文档
func (j *JobWorker) updateDocument(d map[string]interface{}) {
	// 修改后的结构体变量
	reiewId := d["review_id"].(string)
	if len(reiewId) == 0 {
		j.Logger.Error("review_id is empty")
		return
	}
	resp, err := j.ESClient.Update(j.ESClient.index, reiewId).
		Doc(d). // 使用结构体变量更新
		Do(context.Background())
	if err != nil {
		j.Logger.Error("failed to update document", "error", err)
		return
	}
	j.Logger.Debug("updated document", "result", resp.Result)
}

// indexDocument 索引文档
func (j *JobWorker) indexDocument(d map[string]interface{}) {
	reiewId := d["review_id"].(string)
	// 添加文档
	if len(reiewId) == 0 {
		j.Logger.Error("review_id is empty")
		return
	}
	resp, err := j.ESClient.Index(j.ESClient.index).
		Id(reiewId).
		Document(d).
		Do(context.Background())
	if err != nil {
		j.Logger.Error("failed to index document", "error", err)
		return
	}
	j.Logger.Debug("indexed document", "result", resp.Result)
}

// Tag 评价标签（供 ES 文档序列化使用）
type Tag struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Review 写入 ES 的评价文档结构
type Review struct {
	ID          int       `json:"id"`
	UserID      int64     `json:"user_id"`
	Score       int       `json:"score"`
	Content     string    `json:"content"`
	Tags        []Tag     `json:"tags"`
	Status      int       `json:"status"`
	PublishTime time.Time `json:"publish_time"`
}
