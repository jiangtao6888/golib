package trace

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func GrpcClientInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	md, ok := metadata.FromOutgoingContext(ctx)

	if !ok {
		md = metadata.Pairs()
	}

	md.Set(ContextKeyTraceId, GetTraceId(ctx))
	ctx = metadata.NewOutgoingContext(ctx, md)

	return invoker(ctx, method, req, reply, cc, opts...)
}

func GrpcServerInterceptor(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	md, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		md = metadata.Pairs()
	}

	if values := md.Get(ContextKeyTraceId); len(values) > 0 && IsTraceId(values[0]) {
		ctx = WithTraceId(ctx, values[0])
	} else {
		ctx = WithTraceId(ctx, NewTraceId())
	}

	return handler(ctx, req)
}
