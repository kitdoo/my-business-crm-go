package image

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	slogx "github.com/altessa-s/go-atlas/observability/slog"
	httpserver "github.com/altessa-s/go-atlas/transport/http/server"
	"github.com/altessa-s/go-atlas/transport/http/server/writer"

	"github.com/kitdoo/my-business-crm-go/internal/entities"
	"github.com/kitdoo/my-business-crm-go/internal/errs"
	"github.com/kitdoo/my-business-crm-go/internal/rbac"
	usersvc "github.com/kitdoo/my-business-crm-go/internal/services/user"
	"github.com/kitdoo/my-business-crm-go/internal/storages/images"
	grpcrbac "github.com/kitdoo/my-business-crm-go/internal/transports/grpc/interceptors/rbac"
)

// DefaultMaxSizeBytes is used when the configured maxSizeBytes is unset.
const DefaultMaxSizeBytes int64 = 10 << 20 // 10 MiB

const pathPrefix = "/images"

// allowedContentTypes are sniffed via http.DetectContentType — the upload
// carries no trusted extension or Content-Type header.
var allowedContentTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

// Handler implements httpserver.Handler for the image upload/serve routes.
// Upload requires the caller's role to hold grpcrbac.ImagesUploadPermission
// (see requirePermission) — checked the same way as any gRPC method, just
// inline instead of via the RBAC interceptor, since this is a plain-HTTP
// endpoint; serving is unauthenticated, same as any other catalog-image
// asset.
type Handler struct {
	dir     string
	maxSize int64
	users   usersvc.Service
	rbac    rbac.Table
	images  images.Storage
	logger  *slog.Logger
}

var _ httpserver.Handler = (*Handler)(nil)

// New builds a Handler, creating dir if it does not exist. maxSize <= 0
// falls back to DefaultMaxSizeBytes.
func New(dir string, maxSize int64, users usersvc.Service, table rbac.Table, imagesStorage images.Storage) (*Handler, error) {
	if maxSize <= 0 {
		maxSize = DefaultMaxSizeBytes
	}
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, fmt.Errorf("create images directory: %w", err)
	}
	return &Handler{
		dir:     dir,
		maxSize: maxSize,
		users:   users,
		rbac:    table,
		images:  imagesStorage,
		logger:  slog.Default().With(slogx.Module("http:image")),
	}, nil
}

// Register attaches the upload/serve routes.
func (h *Handler) Register(r httpserver.RouteRegistrar, _ <-chan struct{}) {
	r.Handle(pathPrefix, h.upload).Methods(http.MethodPost)
	r.Handle(pathPrefix+"/{id}", h.serve).Methods(http.MethodGet)
}

func (h *Handler) upload(rw writer.ReadWriter) {
	ctx := rw.Request().Context()

	if err := h.requirePermission(ctx, rw.Request()); err != nil {
		writeAuthError(rw, err)
		return
	}

	r := rw.Request()
	// Headroom above maxSize absorbs multipart boundary/header overhead;
	// the real size ceiling is enforced below via the LimitReader.
	r.Body = http.MaxBytesReader(rw.ResponseWriter(), r.Body, h.maxSize+(1<<20))
	if err := r.ParseMultipartForm(h.maxSize); err != nil {
		_ = rw.WriteError(errors.New("invalid or too large multipart upload"), http.StatusBadRequest)
		return
	}
	defer func() {
		if r.MultipartForm != nil {
			_ = r.MultipartForm.RemoveAll()
		}
	}()

	file, _, err := r.FormFile("file")
	if err != nil {
		_ = rw.WriteError(errors.New(`missing "file" field`), http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, h.maxSize+1))
	if err != nil {
		_ = rw.WriteError(errors.New("read upload failed"), http.StatusBadRequest)
		return
	}
	if int64(len(data)) > h.maxSize {
		_ = rw.WriteError(errors.New("file too large"), http.StatusRequestEntityTooLarge)
		return
	}

	contentType := http.DetectContentType(data)
	if !allowedContentTypes[contentType] {
		_ = rw.WriteError(fmt.Errorf("unsupported content type %q", contentType), http.StatusUnsupportedMediaType)
		return
	}

	id := uuid.NewString()
	if err := os.WriteFile(filepath.Join(h.dir, id), data, 0o640); err != nil {
		h.logger.ErrorContext(ctx, "write uploaded image failed", slog.String("id", id), slog.String("error", err.Error()))
		_ = rw.WriteError(errors.New("internal error"), http.StatusInternalServerError)
		return
	}

	// Best-effort past this point: the file is already written and usable
	// (serve falls back to sniffing magic bytes if this record is missing),
	// so a metadata-store hiccup is logged loudly rather than failing an
	// upload whose bytes are already safely on disk.
	if err := h.images.Insert(ctx, &entities.Image{
		ID:          id,
		ContentType: contentType,
		SizeBytes:   int64(len(data)),
		CreatedAt:   time.Now().UTC(),
	}); err != nil {
		h.logger.ErrorContext(ctx, "insert image metadata failed", slog.String("id", id), slog.String("error", err.Error()))
	}

	_ = rw.Write(map[string]string{"id": id})
}

func (h *Handler) serve(rw writer.ReadWriter) {
	id := mux.Vars(rw.Request())["id"]
	// uuid.NewString() never produces "/" or "\\"; reject anything that
	// does rather than letting filepath.Join escape h.dir.
	if id == "" || strings.ContainsAny(id, `/\`) {
		http.NotFound(rw.ResponseWriter(), rw.Request())
		return
	}

	data, err := os.ReadFile(filepath.Join(h.dir, id))
	if err != nil {
		http.NotFound(rw.ResponseWriter(), rw.Request())
		return
	}

	// The metadata record is the source of truth for Content-Type (set at
	// upload time from the same sniff that validated the file); falling
	// back to re-sniffing keeps files uploaded before this record existed
	// working the same as always.
	contentType := ""
	if meta, err := h.images.Get(rw.Request().Context(), id); err == nil {
		contentType = meta.ContentType
	}
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}

	w := rw.ResponseWriter()
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// requirePermission extracts the bearer token and confirms the caller's
// role holds grpcrbac.ImagesUploadPermission — the same RBAC table the gRPC
// interceptor checks (internal/transports/grpc/interceptors/rbac), just
// enforced inline since this is a plain-HTTP endpoint the interceptor never
// sees.
func (h *Handler) requirePermission(ctx context.Context, r *http.Request) error {
	token := bearerToken(r)
	if token == "" {
		return errs.ErrUnauthenticated
	}
	u, err := h.users.Authenticate(ctx, token)
	if err != nil {
		return err
	}
	if !h.rbac.Allowed(u.Role.String(), grpcrbac.ImagesUploadPermission) {
		return errs.ErrForbidden
	}
	return nil
}

func bearerToken(r *http.Request) string {
	const prefix = "Bearer "
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(h, prefix))
}

func writeAuthError(rw writer.ReadWriter, err error) {
	if errors.Is(err, errs.ErrForbidden) {
		_ = rw.WriteError(err, http.StatusForbidden)
		return
	}
	_ = rw.WriteError(err, http.StatusUnauthorized)
}
