package main

import (
	"fmt"
	"net/http"

	"github.com/wangzhipeng/opentracing_grpc/proto"
	"google.golang.org/grpc"

	"context"
	"time"

	grpc_jaeger "github.com/moxiaomomo/grpc-jaeger"
	opentracing_go "github.com/opentracing/opentracing-go"
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
		grpc.WithUnaryInterceptor(grpc_jaeger.ClientInterceptor(opentracing_go.GlobalTracer())))
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
	tracer, _, err := grpc_jaeger.NewJaegerTracer("testSrv", "127.0.0.1:6831")
	if err != nil {
		fmt.Println(err)
	}

	opentracing_go.InitGlobalTracer(tracer)
	port := 8080
	addr := fmt.Sprintf(":%d", port)
	mux := http.NewServeMux()
	mux.HandleFunc("/server1", Handler)
	fmt.Printf("Go to http://localhost:%d/server1 to start a request!\n", port)
	http.ListenAndServe(addr, mux)
}
