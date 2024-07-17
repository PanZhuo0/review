package data

import (
	v1 "business/api/review/v1"
	"business/internal/conf"
	"context"

	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/validate"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/google/wire"
	"github.com/hashicorp/consul/api"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewBusinessRepo, NewReviewServiceClient, NewDiscovery)

// Data .
type Data struct {
	// TODO wrapped database client
	// Data 数据的来源,需要调用review-service的服务
	// 嵌入一个gRPC 客户端,通过这个client去调用review-service的服务
	rc v1.ReviewClient
}

// NewData .
func NewData(rc v1.ReviewClient, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{rc: rc}, cleanup, nil
}

// NewDiscovery服务发现对象的构造方法
func NewDiscovery(conf *conf.Registry) registry.Discovery {
	// 1.新建consul连接
	c := api.DefaultConfig()
	//使用配置文件中的注册中心的配置
	c.Address = conf.Consul.Address
	c.Scheme = conf.Consul.Scheme
	conn, err := api.NewClient(c)
	if err != nil {
		panic(err)
	}
	// 2.使用consul连接，实现kratos中的Discoery
	dis := consul.New(conn)
	return dis
}

// 使用注册中心进行服务发现
func NewReviewServiceClient(dis registry.Discovery) v1.ReviewClient {
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithDiscovery(dis),                          //指定使用那个注册中心
		grpc.WithEndpoint("discovery:///review.service"), //指定访问注册中心中的那个服务
		grpc.WithMiddleware(
			recovery.Recovery(),
			validate.Validator(),
		))
	if err != nil {
		panic(err)
	}
	return v1.NewReviewClient(conn)
}
