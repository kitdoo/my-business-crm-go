// Package rbac provides the gRPC server interceptor that enforces
// internal/rbac.Table (config-driven role -> permission grants) per
// method. It must run after the auth interceptor
// (internal/transports/grpc/interceptors/auth), which is what puts the
// authenticated *entities.User in context — see Dependencies.
package rbac

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	stdGrpc "google.golang.org/grpc"

	atlasauth "github.com/altessa-s/go-atlas/transport/grpc/interceptors/auth"

	"github.com/altessa-s/go-atlas/transport/grpc/interceptors"

	"github.com/kitdoo/my-business-crm-go/internal/rbac"
	grpcauth "github.com/kitdoo/my-business-crm-go/internal/transports/grpc/interceptors/auth"
)

const interceptorName = "rbac"

type interceptor struct {
	table rbac.Table
}

var _ interceptors.ServerInterceptor = (*interceptor)(nil)

// New builds the RBAC interceptor. table is the role -> permission grants
// (from CRMConfig.RBAC); a role with no matching entry (or an empty/nil
// table) is denied every permission-gated method — see
// internal/rbac.Table.Allowed.
func New(table rbac.Table) interceptors.ServerInterceptor {
	return &interceptor{table: table}
}

func (i *interceptor) Name() string { return interceptorName }

// Dependencies declares that this interceptor must run after the auth
// interceptor (Name "auth"), whose ClientAuth hook is what populates the
// context this interceptor reads via grpcauth.UserFromContext.
func (i *interceptor) Dependencies() []string { return []string{atlasauth.Name()} }

func (i *interceptor) ServerUnaryInterceptor() stdGrpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *stdGrpc.UnaryServerInfo, handler stdGrpc.UnaryHandler) (any, error) {
		if err := i.authorize(ctx, info.FullMethod); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

func (i *interceptor) ServerStreamInterceptor() stdGrpc.StreamServerInterceptor {
	return func(srv any, stream stdGrpc.ServerStream, info *stdGrpc.StreamServerInfo, handler stdGrpc.StreamHandler) error {
		if err := i.authorize(stream.Context(), info.FullMethod); err != nil {
			return err
		}
		return handler(srv, stream)
	}
}

// authorize enforces permissions for fullMethod. Methods the auth
// interceptor exempted (Login, plus the gRPC reflection/health probes it
// exempts by default) carry no user in context and pass through
// unchecked — there is nothing to authorize since nobody was
// authenticated. Everything else must be in permissions and the caller's
// role must be granted that permission by table.
func (i *interceptor) authorize(ctx context.Context, fullMethod string) error {
	u, ok := grpcauth.UserFromContext(ctx)
	if !ok {
		return nil
	}

	perm, registered := permissions[fullMethod]
	if !registered {
		return status.Errorf(codes.PermissionDenied, "method %s has no registered RBAC permission", fullMethod)
	}
	if perm == permissionAuthenticatedOnly {
		return nil
	}
	if !i.table.Allowed(u.Role.String(), perm) {
		return status.Errorf(codes.PermissionDenied, "role %q lacks permission %q", u.Role.String(), perm)
	}
	return nil
}
