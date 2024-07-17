package server

import (
	"business/internal/conf"

	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/google/wire"
	"github.com/hashicorp/consul/api"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(NewGRPCServer, NewHTTPServer, NewRegistrar)

func NewRegistrar(rc *conf.Registry) registry.Registrar {
	c := api.DefaultConfig()
	c.Address = rc.Consul.Address
	c.Scheme = rc.Consul.Scheme

	client, err := api.NewClient(c)
	if err != nil {
		panic(err)
	}
	// 使用kratos中的consul包创建Register
	reg := consul.New(client, consul.WithHealthCheck(true)) //需要传入一个consul连接
	return reg
}
