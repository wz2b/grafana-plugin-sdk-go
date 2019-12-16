package backend

import (
	"context"
	"fmt"
	"strconv"

	"github.com/grafana/grafana-plugin-sdk-go/genproto/pluginv2"
	plugin "github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type TransformGRPCServer struct {
	broker *plugin.GRPCBroker
	Impl   transformWrapper
}

func (t *TransformGRPCServer) DataQuery(ctx context.Context, req *pluginv2.DataQueryRequest) (*pluginv2.DataQueryResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("transform request is missing metadata")
	}
	rawReqIDValues := md.Get("broker_requestId") // TODO const
	if len(rawReqIDValues) != 1 {
		return nil, fmt.Errorf("transform request metadta is missing broker_requestId")
	}
	id64, err := strconv.ParseUint(rawReqIDValues[0], 10, 32)
	if err != nil {
		return nil, err
	}
	conn, err := t.broker.Dial(uint32(id64))
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	api := &TransformCallBackGrpcClient{pluginv2.NewTransformCallBackClient(conn)}
	return t.Impl.DataQuery(ctx, req, api)
}

type TransformGRPCClient struct {
	broker *plugin.GRPCBroker
	client pluginv2.TransformClient
}

func (t *TransformGRPCClient) DataQuery(ctx context.Context, req *pluginv2.DataQueryRequest, callBack TransformCallBack) (*pluginv2.DataQueryResponse, error) {
	callBackServer := &TransformCallBackGrpcServer{Impl: callBack}

	var s *grpc.Server
	serverFunc := func(opts []grpc.ServerOption) *grpc.Server {
		s = grpc.NewServer(opts...)
		pluginv2.RegisterTransformCallBackServer(s, callBackServer)

		return s
	}
	brokerID := t.broker.NextId()
	go t.broker.AcceptAndServe(brokerID, serverFunc)
	metadata.AppendToOutgoingContext(ctx, "broker_requestId", string(brokerID))
	res, err := t.client.DataQuery(ctx, req)
	s.Stop()
	return res, err
}

// Callback

type TransformCallBackGrpcClient struct {
	client pluginv2.TransformCallBackClient
}

func (t *TransformCallBackGrpcClient) DataQuery(ctx context.Context, req *pluginv2.DataQueryRequest) (*pluginv2.DataQueryResponse, error) {
	return t.client.DataQuery(ctx, req)
}

type TransformCallBackGrpcServer struct {
	Impl TransformCallBack
}

func (g *TransformCallBackGrpcServer) DataQuery(ctx context.Context, req *pluginv2.DataQueryRequest) (*pluginv2.DataQueryResponse, error) {
	return g.Impl.DataQuery(ctx, req)
}