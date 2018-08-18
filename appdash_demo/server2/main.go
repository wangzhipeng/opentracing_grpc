package main

import (
	"flag"
	"fmt"
	"net"
	"time"

	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	opentracing_go "github.com/opentracing/opentracing-go"
	"github.com/wangzhipeng/opentracing_grpc/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"sourcegraph.com/sourcegraph/appdash"
	"sourcegraph.com/sourcegraph/appdash/opentracing"
)

func newServer() *testServer {
	return &testServer{}
}

type testServer struct {
}

var (
	A    = 0
	port = flag.Int("port", 50000, "listening port")
)

func (s *testServer) TestHello(ctx context.Context, request *proto.TestRequest) (*proto.TestResponse, error) {

	A += 1

	time.Sleep(time.Millisecond * 200)
	//fmt.Println(ctx)
	hello(ctx)
	hello1(ctx)

	conn, err := grpc.Dial("127.0.0.1:50001",
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(opentracing_go.GlobalTracer())))
	if err != nil {
		panic(err)
	}

	var re proto.TestRequest
	re.Message = request.Message + " Hello"

	client := proto.NewTest2Client(conn)
	rsp, err := client.TestWord(ctx, &re)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(rsp.GetRetCode())
	return &proto.TestResponse{RetCode: int64(A), Message: rsp.Message}, nil
}

func hello(ctx context.Context) {

	span := opentracing_go.SpanFromContext(ctx) // Start a span using the global, in this case noop, tracer

	sp := opentracing_go.GlobalTracer().StartSpan(
		"/operation_name",
		opentracing_go.ChildOf(span.Context()))

	defer sp.Finish()

	time.Sleep(time.Millisecond * 150)
	fmt.Println("hello")
	//time.Sleep(time.Millisecond*100)
}

func hello1(ctx context.Context) {
	spanctx := opentracing_go.SpanFromContext(ctx)
	clientSpan := opentracing_go.GlobalTracer().StartSpan(
		"operation_db",
		opentracing_go.ChildOf(spanctx.Context()),
	)
	defer clientSpan.Finish()
	time.Sleep(time.Millisecond * 180)
	fmt.Println("hello1")
}

func main() {
	flag.Parse()

	collectorAdd := fmt.Sprintf(":%d", 7777)
	tracer := opentracing.NewTracer(appdash.NewRemoteCollector(collectorAdd))

	opentracing_go.InitGlobalTracer(tracer)

	lis, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", *port))
	if err != nil {
		fmt.Println(err)
		return
	}

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(otgrpc.OpenTracingServerInterceptor(tracer)))
	proto.RegisterTest1Server(grpcServer, newServer())
	grpcServer.Serve(lis)
}
