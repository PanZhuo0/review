package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/update"
	"github.com/elastic/go-elasticsearch/v8/typedapi/some"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
)

/*
	go 操作ES

1.创建连接
2.创建Index
3.增加文档
4.检索文档
5.搜索文档
*/
func connect() *elasticsearch.TypedClient {
	// ES 配置
	cfg := elasticsearch.Config{
		Addresses: []string{
			"http://localhost:9200",
		},
	}
	// 创建客户端连接
	client, err := elasticsearch.NewTypedClient(cfg) //带类型的Client对象
	if err != nil {
		fmt.Printf("elasticsearch.NewTypedClient failed, err:%v\n", err)
		return nil
	}
	return client
}

func createIndx() {
	// 使用typedclient对象创建index
	client := connect()
	response, err := client.Indices.Create("test").Do(context.Background())
	if err != nil {
		fmt.Println("创建index失败,err:", err)
	}
	fmt.Println(response)

}

type Review struct {
	ID          int64     `json:id`
	UserID      int64     `json:"userID"`
	Score       uint8     `json:"score"`
	Content     string    `json:"content"`
	Tags        []Tag     `json:"tags"`
	Status      int       `json:"status"`
	PublishTime time.Time `json:"publishTime"`
}

type Tag struct {
	Code  int    `json:"code"`
	Title string `json:"title"`
}

// 往INDEX中Upquery数据
func indexDocument() {
	c := connect()
	d := Review{
		ID:      1,
		UserID:  19089080,
		Score:   2,
		Content: "这是一个特别好评",
		Tags: []Tag{
			{1000, "好评"},
			{2000, "物超所值"},
			{3000, "有图"},
		},
	}
	resp, err := c.Index("test").Id(fmt.Sprint(d.ID)).Document(d).Do(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println(resp.Result)
}

// 根据ID获取文档
func getDocumentByID() {
	c := connect()
	resp, err := c.Get("test", "1").Do(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println(string(resp.Source_))
}

// 搜索文档
func searchDocument() {
	c := connect()
	resp, err := c.Search().Index("test").Request(&search.Request{
		// 搜索的条件
		Query: &types.Query{
			// 查询所有的文档
			MatchAll: &types.MatchAllQuery{},
			// MatchPhrase: map[string]types.MatchPhraseQuery{
			// 	"content": {Query: "好评"},
			// },
		},
	}).Do(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println(resp.Hits.Total.Value)
	for _, hit := range resp.Hits.Hits {
		fmt.Printf("%s\n", hit.Source_)
	}
}

// 聚合-求好评的平均值
func aggregate() {
	c := connect()
	response, err := c.Search().Index("test").Request(
		&search.Request{
			Size: some.Int(0),
			Aggregations: map[string]types.Aggregations{
				"avg_score": {
					Avg: &types.AverageAggregation{
						Field: some.String("score"),
					},
				},
			},
		},
	).Do(context.Background())
	if err != nil {
		fmt.Printf("aggregation failed,err:%v\n", err)
		return
	}
	fmt.Printf("avgScore:%v\n", response.Aggregations["avg_score"])

}

// 通过RawJSON更新文档
func updateByRawJson() {
	c := connect()
	resp, err := c.Update("test", "1").Request(
		&update.Request{
			Doc: json.RawMessage(
				`{
					  "ID": 1,
					  "userID": 222222222,
					  "score": 2,
					  "content": "这是一个特别好评"}
				 `),
		},
	).Do(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println(resp.Result)
}

// 删除Index中的某个document
func delete() {
	c := connect()
	resp, err := c.Delete("test", "2").Do(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println(resp.Result)
}
func main() {
	// createIndx()
	// indexDocument()
	// getDocumentByID()
	// searchDocument()
	// aggregate()
	// updateByRawJson()
	delete()
}
