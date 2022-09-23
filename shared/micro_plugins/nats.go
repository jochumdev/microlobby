package micro_plugins

import (
	_ "github.com/go-micro/plugins/v4/broker/nats"
	_ "github.com/go-micro/plugins/v4/registry/nats"
	_ "github.com/go-micro/plugins/v4/transport/grpc"
	_ "github.com/go-micro/plugins/v4/transport/nats"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/server"
)

func Init() {
	client.DefaultContentType = "application/grpc+proto"
	server.DefaultContentType = "application/grpc+proto"
}
