package data

import (
	"context"
	v1 "operator/api/review/v1"
	"operator/internal/conf"

	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/google/wire"
	"github.com/hashicorp/consul/api"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewOperationRepo, NewDiscovery, NewReviewServiceClient)

// Data .
type Data struct {
	// TODO wrapped database client
	// 运营端的数据源
	// ...
	// reveiw-service 的 RPC
	rc  v1.ReviewClient
	log *log.Helper
}

// NewData .
func NewData(c *conf.Data, rc v1.ReviewClient, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{
		rc:  rc,
		log: log.NewHelper(logger),
	}, cleanup, nil
}

func NewReviewServiceClient(dis registry.Discovery) v1.ReviewClient {
	// 建立GRPC连接,使用服务发现dis对象连接RCP服务
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint("discovery:///review.service"), //使用kratos提供的匹配模式,kratos会在后台注册解析器resolver
		//这里使用kratos包装的Recovery,而不是Go内置的
		grpc.WithDiscovery(dis),
		grpc.WithMiddleware(recovery.Recovery()))
	if err != nil {
		panic(err)
	}
	// 使用服务发现的连接New一个review-service RPC连接
	return v1.NewReviewClient(conn)
}

// 使用服务发现,参数为配置文件
func NewDiscovery(conf *conf.Registry) registry.Discovery {
	// 1.建立consul连接
	c := api.DefaultConfig()
	c.Address = conf.Consul.Address
	c.Scheme = conf.Consul.Scheme
	cli, err := api.NewClient(c)
	if err != nil {
		panic(err)
	}
	// 2.用kratos中的consul包提供的New方法
	dis := consul.New(cli, consul.WithHealthCheck(true)) //设置进行HC健康检查
	return dis                                           //返回这个服务发现对象
}
