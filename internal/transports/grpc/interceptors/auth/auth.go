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

	categorysvcpb "github.com/kitdoo/my-business-crm-go/proto/gen/go/services/grpc/category/v1"
	notificationsvcpb "github.com/kitdoo/my-business-crm-go/proto/gen/go/services/grpc/notification/v1"
	partnersvcpb "github.com/kitdoo/my-business-crm-go/proto/gen/go/services/grpc/partner/v1"
	pricesvcpb "github.com/kitdoo/my-business-crm-go/proto/gen/go/services/grpc/price/v1"
	productsvcpb "github.com/kitdoo/my-business-crm-go/proto/gen/go/services/grpc/product/v1"
	usersvcpb "github.com/kitdoo/my-business-crm-go/proto/gen/go/services/grpc/user/v1"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	"github.com/kitdoo/my-business-crm-go/internal/errs"
	usersvc "github.com/kitdoo/my-business-crm-go/internal/services/user"
)

type contextKey struct{}

// UserFromContext returns the caller authenticated by this interceptor, or
// (nil, false) for methods exempted from auth (Login, the public-catalog
// read methods, plus the gRPC reflection/health probes go-atlas exempts by
// default).
func UserFromContext(ctx context.Context) (*entities.User, bool) {
	u, ok := ctx.Value(contextKey{}).(*entities.User)
	return u, ok
}

// New builds the gRPC auth interceptor. Login is exempt (it is how a token
// is obtained in the first place). ProductsService.List, PricesService.Get
// and CategoriesService.List are also exempt — they back the public
// website's catalog (web-public/), which has anonymous visitors and no
// login of its own; see web-public/README.md. NotificationsService.Send is
// exempt for the same reason (anonymous visitors submit web-public/'s
// contact/dealer forms through it) but is not left open to anyone who can
// reach the gRPC port: internal/transports/grpc/interceptors/clientkey
// runs independently of this interceptor and requires a valid
// "x-client-key" header on that one method — see CRMConfig.NotificationClients.
// Exempting the others is safe only
// because the gRPC port is never reachable from outside the two Nitro BFFs
// (web/, web-public/) — it is not exposed to the public internet. An
// exempted method also skips the RBAC interceptor (see
// internal/transports/grpc/interceptors/rbac.authorize), so it must not
// leak anything beyond what an anonymous website visitor should see; the
// public catalog server route is responsible for always forcing a
// statuses:[ACTIVE] filter server-side (see
// web-public/server/utils/catalogClient.js) since nothing here restricts
// what an anonymous caller passes as filter. Every other RPC requires a
// valid bearer token resolving to an active user.
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
		atlasauth.WithIgnoreMethods(
			usersvcpb.UsersService_Login_FullMethodName,
			productsvcpb.ProductsService_List_FullMethodName,
			pricesvcpb.PricesService_Get_FullMethodName,
			categorysvcpb.CategoriesService_List_FullMethodName,
			partnersvcpb.PartnersService_ListPublic_FullMethodName,
			notificationsvcpb.NotificationsService_Send_FullMethodName,
		),
	)
}
