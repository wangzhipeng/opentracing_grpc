package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	opentracing_go "github.com/opentracing/opentracing-go"
	"github.com/wangzhipeng/opentracing_grpc/proto"
	"google.golang.org/grpc"
	"sourcegraph.com/sourcegraph/appdash"
	"sourcegraph.com/sourcegraph/appdash/opentracing"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	var ctx opentracing_go.SpanContext

	clientSpan := opentracing_go.GlobalTracer().StartSpan(
		"serverHand",
		opentracing_go.ChildOf(ctx),
	)
	defer clientSpan.Finish()

	time.Sleep(time.Millisecond * 100)
	conn, err := grpc.Dial("127.0.0.1:50000",
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(opentracing_go.GlobalTracer())))
	if err != nil {
		panic(err)
	}

	var request proto.TestRequest
	request.Message = "william "

	client := proto.NewTest1Client(conn)
	chictx := opentracing_go.ContextWithSpan(context.Background(), clientSpan)
	rsp, err := client.TestHello(chictx, &request)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(rsp.GetRetCode())

	w.Write([]byte(fmt.Sprintf("code:%d,message:%s", rsp.GetRetCode(), rsp.Message)))

	// here would be the actual call to a DB.
}

func main() {
	collectorAdd := fmt.Sprintf(":%d", 7777)
	tracer := opentracing.NewTracer(appdash.NewRemoteCollector(collectorAdd))

	opentracing_go.InitGlobalTracer(tracer)
	port := 8080
	addr := fmt.Sprintf(":%d", port)
	mux := http.NewServeMux()
	mux.HandleFunc("/server1", Handler)
	fmt.Printf("Go to http://localhost:%d/server1 to start a request!\n", port)
	http.ListenAndServe(addr, mux)
}
