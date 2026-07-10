// Package auth provides the gRPC server interceptor that authenticates
// requests using the same session-token mechanism as the HTTP transport
// (usersvc.Service.Authenticate) — see
// internal/transports/http/handlers/image for the non-gRPC counterpart.
//
// This interceptor only establishes "who is calling": it resolves the
// bearer token to an *entities.User and stores it in context for handlers
// to read via UserFromContext. Role/permission enforcement per method (see
// internal/rbac) is left to individual handlers, or a future rbac-aware
// interceptor — not this one.
package auth

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	atlasauth "github.com/altessa-s/go-atlas/transport/grpc/interceptors/auth"

	"github.com/altessa-s/go-atlas/transport/grpc/interceptors"

	usersvcpb "github.com/kitdoo/my-business-crm-go/proto/gen/go/services/grpc/user/v1"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	"github.com/kitdoo/my-business-crm-go/internal/errs"
	usersvc "github.com/kitdoo/my-business-crm-go/internal/services/user"
)

type contextKey struct{}

// UserFromContext returns the caller authenticated by this interceptor, or
// (nil, false) for methods exempted from auth (Login, plus the gRPC
// reflection/health probes go-atlas exempts by default).
func UserFromContext(ctx context.Context) (*entities.User, bool) {
	u, ok := ctx.Value(contextKey{}).(*entities.User)
	return u, ok
}

// New builds the gRPC auth interceptor. Login is exempt (it is how a token
// is obtained in the first place); every other RPC requires a valid bearer
// token resolving to an active user.
func New(users usersvc.Service) interceptors.ServerInterceptor {
	return atlasauth.ServerInterceptor(
		atlasauth.WithAuthFn(atlasauth.AuthFunc(func(ctx context.Context, req atlasauth.Request) (any, error) {
			tc, ok := req.TokenCredentials()
			if !ok {
				return nil, status.Error(codes.Unauthenticated, "missing bearer token")
			}
			token := tc.Token.Expose()
			if token == "" {
				return nil, status.Error(codes.Unauthenticated, "missing bearer token")
			}

			u, err := users.Authenticate(ctx, token)
			if err != nil {
				if errors.Is(err, errs.ErrUnauthenticated) {
					return nil, status.Error(codes.Unauthenticated, "unauthenticated")
				}
				return nil, status.Error(codes.Internal, "internal error")
			}
			return u, nil
		})),
		atlasauth.WithClientAuth(atlasauth.ClientAuthFunc(func(ctx context.Context, creds atlasauth.Credentials) (context.Context, error) {
			u, ok := creds.Data.(*entities.User)
			if !ok || u == nil {
				return ctx, status.Error(codes.Unauthenticated, "unauthenticated")
			}
			return context.WithValue(ctx, contextKey{}, u), nil
		})),
		atlasauth.WithIgnoreMethods(usersvcpb.UsersService_Login_FullMethodName),
	)
}
