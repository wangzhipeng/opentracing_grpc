package main

import (
	"flag"
	"fmt"
	"net"
	"time"

	grpc_jaeger "github.com/moxiaomomo/grpc-jaeger"
	opentracing_go "github.com/opentracing/opentracing-go"
	"github.com/wangzhipeng/opentracing_grpc/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type xxx struct {
}

var (
	B = 0
	p = flag.Int("port", 50001, "listening port")
)

func (s *xxx) TestWord(ctx context.Context, request *proto.TestRequest) (*proto.TestResponse, error) {

	time.Sleep(time.Millisecond * 60)

	fmt.Println(ctx)
	B++

	return &proto.TestResponse{RetCode: int64(B), Message: request.Message + " word!"}, nil
}

func main() {
	flag.Parse()
	tracer, _, err := grpc_jaeger.NewJaegerTracer("testSrv", "127.0.0.1:6831")
	if err != nil {
		fmt.Println(err)
		return
	}

	opentracing_go.InitGlobalTracer(tracer)

	lis, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", *p))
	if err != nil {
		fmt.Println(err)
		return
	}

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(grpc_jaeger.ServerInterceptor(tracer)))
	proto.RegisterTest2Server(grpcServer, &xxx{})
	grpcServer.Serve(lis)
}
