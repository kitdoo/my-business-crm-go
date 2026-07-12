// Package clientkey provides a gRPC server interceptor that gates a
// configured set of methods behind a static per-caller API key, instead
// of the user bearer-token auth in internal/transports/grpc/interceptors/
// auth. It exists for RPCs anonymous website visitors must reach (so a
// user session token isn't available) but that must still be restricted
// to approved frontends rather than any caller that can reach the gRPC
// port — currently just NotificationsService.Send.
package clientkey

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/altessa-s/go-atlas/transport/grpc/interceptors"
)

const (
	interceptorName = "clientkey"

	// HeaderName is the gRPC metadata key callers must set to their
	// approved key.
	HeaderName = "x-client-key"
)

type interceptor struct {
	// keys maps an approved API key to the client name it belongs to
	// (CRMConfig.NotificationClients, inverted) — the name isn't used for
	// authorization, only available to a handler that wants to log who
	// called.
	keys      map[string]string
	protected map[string]struct{}
}

var _ interceptors.ServerInterceptor = (*interceptor)(nil)

// New builds the client-key interceptor. keys maps an approved API key to
// its client name (CRMConfig.NotificationClients, inverted — see
// internal/fx). protectedMethods are the gRPC full method names this
// interceptor checks; every other method passes through untouched (it is
// either public or guarded by the user auth/RBAC interceptors instead).
// An empty keys map denies every call to a protected method — fail
// closed, same as an empty RBAC table.
func New(keys map[string]string, protectedMethods ...string) interceptors.ServerInterceptor {
	protected := make(map[string]struct{}, len(protectedMethods))
	for _, m := range protectedMethods {
		protected[m] = struct{}{}
	}
	return &interceptor{keys: keys, protected: protected}
}

func (i *interceptor) Name() string { return interceptorName }

// Dependencies is empty: this interceptor only reads request metadata, so
// it has no ordering requirement against the user auth/RBAC interceptors.
func (i *interceptor) Dependencies() []string { return nil }

func (i *interceptor) ServerUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if err := i.authorize(ctx, info.FullMethod); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

func (i *interceptor) ServerStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if err := i.authorize(stream.Context(), info.FullMethod); err != nil {
			return err
		}
		return handler(srv, stream)
	}
}

// authorize is a no-op for any method not in i.protected — see New.
func (i *interceptor) authorize(ctx context.Context, fullMethod string) error {
	if _, ok := i.protected[fullMethod]; !ok {
		return nil
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "missing client key")
	}
	vals := md.Get(HeaderName)
	if len(vals) == 0 || vals[0] == "" {
		return status.Error(codes.Unauthenticated, "missing client key")
	}
	if _, known := i.keys[vals[0]]; !known {
		return status.Error(codes.Unauthenticated, "invalid client key")
	}
	return nil
}
