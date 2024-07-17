package server

import (
	"review-service/internal/conf"

	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/google/wire"
	"github.com/hashicorp/consul/api"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(NewGRPCServer, NewHTTPServer, NewRegistrar)

// 这里提供registrar
func NewRegistrar(rc *conf.Registry) registry.Registrar {
	// 1.创建一个consul连接
	c := api.DefaultConfig()
	c.Address = rc.Consul.Address
	c.Scheme = rc.Consul.Scheme
	client, err := api.NewClient(c)
	if err != nil {
		panic(err)
	}
	// 2.把这个连接注册到kratos的consul包
	reg := consul.New(client)
	return reg
}
