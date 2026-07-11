# Service Development Standard

> **Version**: 1.0
> **Status**: Active

This standard governs the Go service code in this repository. It is the
implementation-side counterpart to [`PROTO_DEVELOPMENT_STANDARD.md`](./PROTO_DEVELOPMENT_STANDARD.md):
the proto standard defines the wire contract, this one defines how the
service maps that contract onto entities, services, storages, and transports.

It is built **on top of** [`go-atlas`](https://github.com/altessa-s/go-atlas).
The single most important rule: **reuse go-atlas; do not hand-roll what it
already provides.** Hand-written boilerplate (struct-to-struct mapping, manual
errgroups, ad-hoc retries, string error matching) is a defect, not a style
choice.

Every example below uses a generic `Product` entity. The patterns apply to
any aggregate — substitute your entity's name wherever `Product` appears.

---

## Table of Contents

1. [General Principles](#general-principles)
2. [Critical Rules](#critical-rules)
3. [Layered Architecture](#layered-architecture)
4. [Entities Layer](#entities-layer)
5. [Storages Layer](#storages-layer)
6. [Services Layer](#services-layer)
7. [Transport Layer (gRPC handlers)](#transport-layer-grpc-handlers)
8. [Errors](#errors)
9. [Options (optgen)](#options-optgen)
10. [Optional and Result Types](#optional-and-result-types)
11. [go-atlas Reuse Map](#go-atlas-reuse-map)
12. [Documentation Standard](#documentation-standard)
13. [Templates](#templates)
14. [Review Checklist](#review-checklist)
15. [Conclusion](#conclusion)

---

## General Principles

### 1. Reuse before writing

Before writing any helper, check go-atlas `core/*` and `data/*` first. Manual
implementations of things go-atlas provides are rejected in review. The full
catalog is in [go-atlas Reuse Map](#go-atlas-reuse-map). The headline rules:

- **Struct mapping** → `domain/converter`, never hand-written `entityToModel` /
  `modelToEntity` functions.
- **Mongo documents** → `mongo.ConvertToNewDocument` / `ConvertToUpdateDocument`,
  never a hand-built `bson.M{...}` mirror of a struct.
- **Concurrency** → `core/runtime/concurrency.Process`, never raw
  `errgroup` + `semaphore`.
- **Retry** → `core/retry.Do`, never `for`-loops with `time.Sleep`.
- **Errors** → `core/errors.WrapOperation`, never `fmt.Errorf("...: %w")` in new code.
- **Error classification** → `coreerrs.Is*`, never `strings.Contains(err.Error(), ...)`.

### 2. Layer separation is absolute

```
transport ─► service ─► storage ─► MongoDB
   (proto)    (entity)   (entity)    (bson model)
```

- Transports speak proto and entities. They never touch bson or storage internals.
- Services speak entities only. They never import a proto package or a bson model.
- Storages speak entities (in/out) and bson models (internal). They never import
  a proto package.
- The bson `model` struct is **private** to the storage's `mongo` package.

### 3. Interface in parent, implementation in subpackage

Every service and storage exposes its interface in the parent package and its
implementation one directory down:

```
services/product/interface.go          # type Service interface
services/product/product/service.go    # type Service struct (impl)
storages/products/interface.go         # type Storage interface
storages/products/mongo/storage.go     # type Storage struct (impl)
```

The implementation always asserts the interface at compile time:

```go
var _ product.Service = (*Service)(nil)
var _ products.Storage = (*Storage)(nil)
```

### 4. Request-scope data is non-optional where the domain needs it

If an aggregate is scoped by a request-context identifier (an owning
organization, warehouse, or similar boundary), every storage method that
reads or mutates it takes that identifier and filters on it — see
[`PROTO_DEVELOPMENT_STANDARD.md` § Request Context vs Filters](./PROTO_DEVELOPMENT_STANDARD.md)
for how such fields are shaped on the wire. Aggregates with no such boundary
skip this entirely; do not add scoping fields the domain doesn't need.

### 5. The skeleton is a contract

Unimplemented methods return `errs.ErrNotImplemented`. Implement a feature by
filling the existing method body. Do not widen interfaces ad-hoc, do not split
into parallel packages for a new API version, do not add top-level packages
without updating the project's structure documentation.

---

## Critical Rules

### 🔁 Always use `converter`, never hand-write mapping

**This is the rule most often violated.** Struct-to-struct copying MUST go
through `domain/converter`. Hand-written `entityToModel` / `modelToEntity`
functions are forbidden in new code and migrated on touch.

```go
// ❌ WRONG — hand-written field-by-field mapping:
func entityToModel(p *entities.Product) model {
    return model{
        ID:        p.ID,
        SKU:       p.SKU,
        Name:      p.Name,
        CreatedAt: p.CreatedAt,
        Etag:      p.Etag,
        // ...20 more lines that rot every time a field is added...
    }
}

// ✅ CORRECT — converter does the field matching by name:
m := converter.Convert(p, &model{}, converter.WithHandleEmbeddedStructs(true))
m.CursorId = bson.NewObjectID()
```

### 🗄️ Always build Mongo documents with `ConvertToNewDocument` / `ConvertToUpdateDocument`

Inserts and updates go through the go-atlas Mongo document converters, which
apply the `bson` struct tags (`omitonupdate`, `omitempty`) and codecs (e.g.
`time.Time` handling) consistently.

```go
// ❌ WRONG — hand-built update document:
update := bson.M{"$set": bson.M{"name": p.Name, "updated_at": p.UpdatedAt}}

// ✅ CORRECT — for inserts:
m := converter.Convert(p, &model{}, converter.WithHandleEmbeddedStructs(true))
m.CursorId = bson.NewObjectID()
doc, err := s.db.ConvertToNewDocument(ctx, m)
if err != nil {
    return coreerrs.WrapOperation(err, "convert product document")
}
// InsertOne(ctx, doc)

// ✅ CORRECT — for updates (respects `omitonupdate` tags):
update, err := s.db.ConvertToUpdateDocument(ctx, converter.Convert(p, &model{}))
if err != nil {
    return coreerrs.WrapOperation(err, "build update document")
}
```

**Exception** — targeted partial writes that must NOT touch sibling fields
(e.g. a `SoftDelete` `$set`, or an append-only sub-array update) build a
minimal `bson.M{"$set": ...}` by hand. Document *why* the converter is
bypassed in one line.

### 🏷️ Etag (optimistic concurrency) is generated in the entity, never in storage

The single consistent rule across **all** aggregates:

| Operation | Where the etag/timestamp is produced |
|-----------|--------------------------------------|
| Create    | entity constructor `XxxNew` (calls `UpdateEtag` + `UpdateTimestamps`) |
| Update    | service calls `entity.BeforeUpdate()` |
| Delete    | service calls `entity.BeforeUpdate()`, passes values via `entities.SoftDelete` |

Storage methods **never** call `etag.Generate()` or `time.Now()` for the OCC
token or timestamps — they receive these as data and write them.

### 🗑️ All deletes are soft — no exceptions

No aggregate is ever hard-deleted. Soft delete stamps `deleted_at`, the read
paths filter it out (`activeOnly`), and the unique indexes are **partial** on
`{deleted_at: null}` so the natural key (SKU, slug, code, …) frees on delete.
This applies uniformly across every aggregate in the domain — an entity does
not get a hand-rolled hard-delete path because it "feels" like a log or a
child record.

> **MongoDB gotcha:** a partial index filter MUST NOT use `$exists: false`
> (it desugars to `$not`, which partial indexes reject:
> *"Expression not supported in partial index"*). Instead, store `deleted_at`
> as an explicit `null` on active rows (no `omitempty`) and filter on
> `{deleted_at: null}` — an equality expression, which partial indexes accept.

### 🔇 Documentation is minimal by default

Write the least documentation that makes intent clear. See
[Documentation Standard](#documentation-standard). Godoc is 1–3 lines, no error
enumeration, no restating the signature.

---

## Layered Architecture

```
internal/
├── entities/                 # pure domain structs + OCC helpers + patch/list types
├── errs/                     # sentinel errors
├── services/<name>/          # Service interface
│   └── <name>/                # Service impl (+ metrics.go)
├── storages/<resource>/      # Storage interface
│   └── mongo/                 # Mongo impl: storage.go, entity.go (private model), <ts>_*_indexes.go, doc.go
├── transports/grpc/handlers/ # gRPC handlers: handler.go (+ MapError), doc.go
└── fx/                        # fx modules: infrastructure.go, transport.go, services.go
```

Each subsystem exports `Module fx.Option = fx.Options(...)` and is wired in
`internal/fx/`.

---

## Entities Layer

`internal/entities/` holds **plain domain structs with no database or transport
annotations** (except `normalize` tags on input/patch types). Absence is `nil`
(pointers), mirroring storage models and proto messages so `converter.Convert`
maps them directly.

### 1. Resource entity + OCC helpers

Every mutable aggregate carries an `Etag string` OCC token and a
`DeletedAt *time.Time` soft-delete marker, plus the standard helper set
(constructor, `UpdateTimestamps`, `UpdateEtag`, `BeforeUpdate`). This set is
**identical** across aggregates — copy it verbatim.

```go
type Product struct {
    ID          string
    SKU         string
    Name        string
    Description string
    BrandID     string
    CreatedAt   time.Time
    UpdatedAt   time.Time
    DeletedAt   *time.Time // nil = active
    Etag        string     // OCC token; rolled on every write
}

// ProductNew creates a Product with a fresh ID, timestamps, and etag.
func ProductNew(init ...func(*Product)) *Product {
    p := &Product{ID: uuid.NewString()}
    if len(init) > 0 {
        init[0](p)
    }
    p.UpdateTimestamps()
    p.UpdateEtag()
    return p
}

func (p *Product) UpdateTimestamps() {
    now := time.Now().UTC()
    p.UpdatedAt = now
    if p.CreatedAt.IsZero() {
        p.CreatedAt = now
    }
}

func (p *Product) UpdateEtag()   { p.Etag = etag.Generate() }
func (p *Product) BeforeUpdate() { p.UpdateTimestamps(); p.UpdateEtag() }
```

Rules:
- `XxxNew` is the only place a create-time etag is produced.
- `BeforeUpdate` is the only place an update/delete-time etag is produced.
- IDs are UUID strings (`uuid.NewString()`).
- UTC always (`time.Now().UTC()`).

### 2. Patch types with `normalize` tags and `Merge()`

Create / Update / Delete inputs are dedicated structs. Mutable optional fields
are pointers (`nil` = "leave unchanged"). String fields carry `normalize` tags;
the service runs `normalizer.Normalize(in)` before use. Each patch type owns a
`Merge(dst)` that applies its non-nil fields onto the entity **via the
converter**.

```go
type ProductCreate struct {
    SKU     string `normalize:"trim"`
    Name    string `normalize:"trim"`
    BrandID string `normalize:"trim"`
}

func (c *ProductCreate) Merge(dst *Product) *Product {
    if c == nil || dst == nil {
        return dst
    }
    converter.Convert(c, dst, converter.WithIgnoreNilValues())
    return dst
}

type ProductUpdate struct {
    ID          string  `normalize:"trim"`
    Name        *string `normalize:"trim,nil_on_empty"`
    Description *string `normalize:"trim,nil_on_empty"`
    Etag        *string `normalize:"trim,nil_on_empty"`
}

func (u *ProductUpdate) Merge(dst *Product) *Product {
    if u == nil || dst == nil {
        return dst
    }
    converter.Convert(u, dst, converter.WithIgnoreNilValues())
    return dst
}
```

Rules:
- `Merge` MUST use `converter.Convert(..., converter.WithIgnoreNilValues())` —
  no hand-written `if p != nil { dst.X = *p }` ladders.
- `nil_on_empty` collapses `""` to `nil` so an empty patch field means "unset".
- The patch's `Etag *string` is the **client-supplied** OCC precondition,
  checked in the service; it is distinct from the entity's current `Etag`.

### 3. List input types — scope inside, single argument

List inputs are a single struct carrying sort, pagination, filters, and any
request-scope identifier the domain needs. The signature everywhere is
`List(ctx, in *XxxList)` — the scope lives **inside** `in`, never as a
separate positional argument.

```go
type ProductsListSort struct {
    Field     ProductsListSortField
    Direction SortDirection
}

type ProductsList struct {
    BrandID           *string `normalize:"trim,nil_on_empty"`
    Sort              ProductsListSort
    Pagination        ListPagination
    IncludeTotalCount bool
}
```

Shared list primitives (`ListPagination`, `List[T]`, `SortDirection`) live in
`entities/list.go` and are reused by every aggregate.

### 4. Soft-delete request types

`entities.SoftDelete` (single row) and `entities.BulkSoftDelete` (cascade)
carry the values the service rolled, so storage generates nothing:

```go
type SoftDelete struct {
    ID           string
    Etag         string    // OCC filter; "" skips the check
    NewUpdatedAt time.Time // stamped onto deleted_at AND updated_at
    NewEtag      string
}

type BulkSoftDelete struct {
    NewUpdatedAt time.Time
    NewEtag      string
}
```

### 5. Enums — int-typed and value-aligned with the proto

A domain enum that has a proto counterpart MUST be an **`int32`-based type whose
values match the proto enum integers** (Unspecified = 0, …). Then
`converter.Convert` maps it both ways as a plain scalar — no `statusToProto` /
`statusFromProto` switch, no codec.

```go
// ✅ CORRECT — converter maps ProductStatus ⇄ proto ProductStatus directly:
type ProductStatus int32

const (
    ProductStatusUnspecified ProductStatus = iota // 0, matches proto
    ProductStatusDraft                             // 1
    ProductStatusActive                            // 2
    ProductStatusInactive                          // 3
)

// ❌ WRONG — string enum can't bridge to an int32 proto enum, forcing a
// hand-written switch in the handler:
type ProductStatus string
const ProductStatusDraft ProductStatus = "draft"
```

In storage, persist the enum as `int32` (mirroring the proto wire value), not a
string. In the handler, map with a plain cast (`entities.ProductStatus(in.GetStatus())`)
or let `converter.Convert` carry it as part of the whole message — never a switch.

### 6. Timestamps are Unix seconds on the wire, `time.Time` in the entity

Per [`PROTO_DEVELOPMENT_STANDARD.md`](./PROTO_DEVELOPMENT_STANDARD.md), proto
timestamp fields (`createdAt`, `updatedAt`, `deletedAt`, …) are `int64` Unix
seconds — **never** `google.protobuf.Timestamp`. Entities still use
`time.Time` internally; the boundary conversion is a codec, not a manual
`.Unix()` / `time.Unix(...)` call at every call site (see
[Transport Layer § proto ⇄ entity via converter](#2-proto--entity-via-converter)).

---

## Storages Layer

`internal/storages/<resource>/` — interface in parent, MongoDB implementation in
`mongo/`. The bson `model` is private to the `mongo` package.

### 1. The bson model

```go
const collectionName = "products"

const (
    FieldID        = "_id"
    FieldSKU       = "sku"
    FieldDeletedAt = "deleted_at"
    FieldEtag      = "etag"
    FieldCursorId  = "cursor_id"
)

type model struct {
    ID          string        `bson:"_id"`
    SKU         string        `bson:"sku,omitonupdate"`
    Name        string        `bson:"name"`
    Description string        `bson:"description"`
    BrandID     string        `bson:"brand_id"`
    CreatedAt   time.Time     `bson:"created_at,omitonupdate"`
    UpdatedAt   time.Time     `bson:"updated_at"`
    DeletedAt   *time.Time    `bson:"deleted_at,omitonupdate"`
    Etag        string        `bson:"etag"`
    CursorId    bson.ObjectID `bson:"cursor_id,omitonupdate"`
}
```

Rules:
- Immutable fields (`_id`, natural keys like `sku`, `created_at`, `cursor_id`,
  `deleted_at`) carry `omitonupdate` so `ConvertToUpdateDocument` never rewrites them.
- `deleted_at` is stored as explicit `null` on active rows (no `omitempty`) so the
  partial unique index `{deleted_at: null}` can match them; a soft-deleted row gets a
  timestamp and leaves the index, freeing the natural key.
- Field-name constants (`FieldXxx`) are mandatory; never inline a bson key string in a query.
- `cursor_id` is a separate `bson.ObjectID` keyset tie-breaker for cursor pagination
  (entity IDs are UUIDs, which `ListCursor` cannot use).

### 2. Mapping is converter-only

Storages convert with `domain/converter`, never hand-written functions. Slice
elements use `coreslices.To`.

```go
// entity → model (insert path)
m := converter.Convert(p, &model{}, converter.WithHandleEmbeddedStructs(true))
m.CursorId = bson.NewObjectID()
doc, err := s.db.ConvertToNewDocument(ctx, m)

// model → entity (read path)
return converter.Convert(&m, &entities.Product{})

// slices
Items: coreslices.To(result.Items, func(m model) *entities.Product {
    return converter.Convert(&m, &entities.Product{})
})
```

### 3. Standard CRUD shape

```go
func (s *Storage) Insert(ctx context.Context, p *entities.Product) error {
    ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)
    defer cancel()

    m := converter.Convert(p, &model{})
    m.CursorId = bson.NewObjectID()
    if _, err := s.collection.InsertOne(ctx, m); err != nil {
        if dupErr := classifyDuplicate(err); dupErr != nil {
            return dupErr
        }
        return coreerrs.WrapOperation(err, "insert product")
    }
    return nil
}

func (s *Storage) Update(ctx context.Context, p *entities.Product, oldEtag string) error {
    ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)
    defer cancel()

    update, err := s.db.ConvertToUpdateDocument(ctx, converter.Convert(p, &model{}))
    if err != nil {
        return coreerrs.WrapOperation(err, "build update document")
    }
    filter := activeOnly(bson.M{FieldID: p.ID})
    if oldEtag != "" {
        filter[FieldEtag] = oldEtag
    }
    res, err := s.collection.UpdateOne(ctx, filter, update)
    if err != nil {
        if dupErr := classifyDuplicate(err); dupErr != nil {
            return dupErr
        }
        return coreerrs.WrapOperation(err, "update product")
    }
    if res.MatchedCount == 0 {
        return errs.ErrStaleEntity
    }
    return nil
}
```

Rules:
- Every method applies a query timeout: `ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)`.
- Every error is wrapped with `coreerrs.WrapOperation(err, "<lower-case verb phrase>")`.
- `not found` reads return the domain sentinel (`errs.ErrProductNotFound`), mapping
  `mongo.ErrNoDocuments`.
- A zero `MatchedCount` on a guarded write is `errs.ErrStaleEntity`.

### 4. Soft delete

```go
func (s *Storage) SoftDelete(ctx context.Context, in *entities.SoftDelete) error {
    ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)
    defer cancel()

    filter := activeOnly(bson.M{FieldID: in.ID})
    if in.Etag != "" {
        filter[FieldEtag] = in.Etag
    }
    update := bson.M{"$set": bson.M{
        FieldDeletedAt: in.NewUpdatedAt,
        FieldUpdatedAt: in.NewUpdatedAt,
        FieldEtag:      in.NewEtag,
    }}
    res, err := s.collection.UpdateOne(ctx, filter, update)
    if err != nil {
        return coreerrs.WrapOperation(err, "soft delete product")
    }
    if res.MatchedCount == 0 {
        if in.Etag != "" {
            return errs.ErrStaleEntity
        }
        return errs.ErrProductNotFound
    }
    return nil
}
```

`activeOnly` is the shared filter helper, one per `mongo` package:

```go
func activeOnly(filter bson.M) bson.M {
    filter[FieldDeletedAt] = nil // matches active rows (deleted_at null); aligns with the partial index
    return filter
}
```

### 5. Duplicate-key classification

Map Mongo duplicate-key errors to domain sentinels by the conflicting field —
never by string matching. Use `datamongo.IsErrorDuplicate`.

```go
func classifyDuplicate(err error) error {
    isDup, fields := datamongo.IsErrorDuplicate(err)
    if !isDup || fields == nil {
        return nil
    }
    switch {
    case fields.Contains(FieldSKU):
        return errs.ErrProductSKUConflict
    default:
        return nil
    }
}
```

### 6. Cursor pagination

Listings use `datamongo.ListCursor[model]` with `activeOnly` filters, a `-`
prefixed sort field for descending order, and translate cursor errors to
`errs.ErrInvalidListCursor`.

```go
func (s *Storage) List(ctx context.Context, in *entities.ProductsList) (*entities.List[entities.Product], error) {
    ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)
    defer cancel()

    limit := in.Pagination.Limit
    if limit <= 0 {
        limit = defaultListLimit
    }
    filter := bson.M{}
    if in.BrandID != nil {
        filter[FieldBrandID] = *in.BrandID
    }
    opts := []datamongo.ListCursorOption{
        datamongo.WithListCursorFilter(activeOnly(filter)),
        datamongo.WithListCursorSort(listSortField(in.Sort)),
        datamongo.WithListCursorLimit(limit),
        datamongo.WithListCursorTotal(in.IncludeTotalCount),
    }
    if in.Pagination.Cursor != "" {
        opts = append(opts, datamongo.WithListCursorCursor(in.Pagination.Cursor))
    }
    result, err := datamongo.ListCursor[model](ctx, s.collection, opts...)
    if err != nil {
        if errors.Is(err, datamongo.ErrInvalidCursor) ||
            errors.Is(err, datamongo.ErrCursorChecksumMismatch) ||
            errors.Is(err, datamongo.ErrCursorFilterMismatch) {
            return nil, errs.ErrInvalidListCursor
        }
        return nil, coreerrs.WrapOperation(err, "list products")
    }
    // ...map result.Items via converter; copy NextCursor/Total...
}
```

### 7. Index migrations

One migration file per concern, named `<unix-ts>_<resource>_<concern>.go`,
registering via `migrate.Register` in `init()`. Unique indexes that must free
their key on soft delete are **partial** on `{deleted_at: null}` (never
`$exists: false` — partial indexes reject it). A unique `cursor_id` index is
paired with an idempotent backfill that runs **before** index creation.

```go
Options: mongooptions.Index().
    SetName("idx_product_sku").
    SetUnique(true).
    SetPartialFilterExpression(bson.M{FieldDeletedAt: nil}),
```

> ⚠️ Migrations run once. Changing an already-applied index in place will not
> re-apply — cut a **new** migration file to alter an existing index.

---

## Services Layer

`internal/services/<name>/` — interface in parent, impl in `<name>/`. The
service is the only layer that **rolls the etag** and orchestrates dependencies.

### 1. A service controls only its own storage

A service holds exactly one `Storage` — its own. Any interaction with
another entity's data (an existence check, a lookup, a mutation) goes
through **that entity's Service**, never through its `Storage` directly.
Reaching into a sibling entity's storage skips that entity's own
validation, error mapping, and any future logic added there — the whole
point of the service layer is to be the one place that logic lives.

```go
// Wrong — product depends on brands.Storage directly.
type Service struct {
    storage products.Storage
    brands  brands.Storage
}

// Right — product depends on brand.Service.
type Service struct {
    storage products.Storage
    brands  brandsvc.Service
}
```

Since the dependency is a real capability, not a raw data-access shape,
declare the **narrowest interface the consumer actually needs** in the
consumer's own `interface.go` (not the foreign service's own `Service`
interface) — e.g. a `ProductGetter interface { Get(ctx, id string)
(*entities.Product, error) }` rather than the full `product.Service`.
This is the same narrow-interface pattern already used for
`brand.ProductsExistenceChecker` / `warehouse.InventoryChecker`: it keeps
the dependency to exactly the method(s) used, and — critically — it lets
the foreign service satisfy the interface **structurally**, with no
import of the foreign service's package required. That structural typing
is what makes the next rule possible.

**Exception — genuine circular dependencies.** Two services can each
legitimately need to ask the other something (e.g. `Product.Create`
needs `Brand.Get`/`Category.Get` to validate the FK it's about to write;
`Brand.Delete`/`Category.Delete` need to know whether any `Product` still
references them). Wiring **both** directions through the Service layer is
not possible — the DI container cannot construct two services that each
require the other to exist first. When this happens:
- Pick **one** direction to go through the Service (normally the
  direction doing FK validation on write — the more security/correctness
  -sensitive one).
- The other direction (normally a read-only existence check gating a
  Delete) is the **one sanctioned case** where depending on the foreign
  entity's `Storage` directly is acceptable — but only for a narrow,
  named, read-only capability (an `ExistsForX(ctx, id) (bool, error)`
  shape, never a general `Get`/mutation), and the constructor doc comment
  must say *why* in one sentence ("would otherwise cycle with
  `product.Service`, which itself depends on this service").
- This is intentionally rare. If you find yourself reaching for it for a
  third or fourth dependency on the same service, the FK-validation
  direction is probably wired backwards — reconsider which side should
  own the check before adding another exception.

### 2. Struct, constructor, logger, metrics

```go
type Service struct {
    storage   products.Storage
    brands    brandsvc.Service
    collector metrics.Collector
    logger    *slog.Logger
}

func New(storage products.Storage, brands brandsvc.Service, collector metrics.Collector) *Service {
    return &Service{
        storage:   storage,
        brands:    brands,
        collector: collector,
        logger:    slog.Default().With(slogx.Module("service:product")),
    }
}
```

Rules:
- Mandatory dependencies are positional constructor args, in dependency-then-collector order.
- The logger is `slog.Default().With(slogx.Module("service:<name>"))`.
- Accept interfaces (`products.Storage`, `brandsvc.Service` or a narrower
  consumer-defined interface — see §1), store interfaces, never concrete
  `*mongo.Storage` or a foreign service's concrete implementation type.

### 3. Create / Update / Delete flow

```go
func (s *Service) Create(ctx context.Context, in *entities.ProductCreate) (*entities.Product, error) {
    _ = normalizer.Normalize(in) //nolint:errcheck

    if _, err := s.brands.Get(ctx, in.BrandID); err != nil { // dependency gate
        return nil, err
    }
    p := entities.ProductNew()
    in.Merge(p)

    if err := s.storage.Insert(ctx, p); err != nil {
        s.logger.DebugContext(ctx, "insert product failed",
            slog.String("sku", p.SKU), slogx.Error(err))
        return nil, err
    }
    return p, nil
}

func (s *Service) Update(ctx context.Context, in *entities.ProductUpdate) (*entities.Product, error) {
    _ = normalizer.Normalize(in) //nolint:errcheck

    p, err := s.storage.Get(ctx, in.ID)
    if err != nil {
        return nil, err
    }
    if in.Etag != nil && *in.Etag != p.Etag { // client OCC precondition
        return nil, errs.ErrStaleEntity
    }
    oldEtag := p.Etag
    in.Merge(p)
    p.BeforeUpdate() // <-- the only place the new etag is produced
    if err := s.storage.Update(ctx, p, oldEtag); err != nil {
        return nil, err
    }
    return p, nil
}

func (s *Service) Delete(ctx context.Context, in *entities.ProductDelete) error {
    _ = normalizer.Normalize(in) //nolint:errcheck

    p, err := s.storage.Get(ctx, in.ID)
    if err != nil {
        return err
    }
    if in.Etag != nil && *in.Etag != p.Etag {
        return errs.ErrStaleEntity
    }
    oldEtag := p.Etag
    now := time.Now().UTC()
    p.DeletedAt = &now
    p.BeforeUpdate()
    return s.storage.SoftDelete(ctx, &entities.SoftDelete{
        ID: p.ID, Etag: oldEtag,
        NewUpdatedAt: *p.DeletedAt, NewEtag: p.Etag,
    })
}
```

Rules:
- `normalizer.Normalize(in)` first, on every mutating input.
- Validate cross-aggregate dependencies (e.g. brand existence) before writing.
- Build new entities with `XxxNew`; apply patches with `in.Merge(entity)`.
- Roll the etag with `entity.BeforeUpdate()` immediately before the storage call.
- Capture `oldEtag := entity.Etag` **before** `Merge`/`BeforeUpdate` for the OCC filter.
- Log failures at `Debug` with `slog.String` context + `slogx.Error(err)`; return the
  error unchanged (the handler maps it).

### 4. Transport-only concerns stay out of the service

Read masks, update masks, and `includeTotalCount` are resolved in the
handler. The service sees only business inputs.

---

## Transport Layer (gRPC handlers)

`internal/transports/grpc/handlers/<name>/handler.go`. Handlers translate proto
⇄ entity with `converter`, resolve masks, delegate to the service, and map
errors.

### 1. Handler shell

```go
// unixtime.New() bridges int64 Unix-second proto fields (createdAt, updatedAt, …)
// with time.Time entity fields — see PROTO_DEVELOPMENT_STANDARD.md § Timestamps.
var withUnixTimeCodec = converter.WithCodecs(unixtime.New())

type Handler struct {
    productpb.UnimplementedProductsServiceServer
    svc product.Service
}

func New(svc product.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Register(gs *grpc.Server, _ <-chan struct{}) {
    productpb.RegisterProductsServiceServer(gs, h)
}
```

### 2. proto ⇄ entity via converter

```go
func (h *Handler) Create(ctx context.Context, in *productpb.ProductCreateRequest) (*productpb.Product, error) {
    p, err := h.svc.Create(ctx, converter.Convert(in, &entities.ProductCreate{}))
    if err != nil {
        return nil, MapError(err)
    }
    out := converter.Convert(p, &productpb.Product{}, withUnixTimeCodec)
    applyReadMask(in.GetOptions().GetReadMask(), out)
    return out, nil
}
```

Rules:
- proto→entity and entity→proto are `converter.Convert` calls — never field-by-field.
- **No hand-written `<entity>ToProto` / `protoTo<entity>` / `statusToProto` helpers.**
  Enums map via int-alignment (see [Entities § Enums](#5-enums--int-typed-and-value-aligned-with-the-proto)),
  timestamps via the `unixtime` codec, nested slices automatically. A `convert.go`
  full of switches is the signal this rule was missed.
- Timestamps require `withUnixTimeCodec` — the codec handles nested timestamps
  too (e.g. a nested message's `updatedAt` inside a slice). List responses also
  use `converter.WithIgnoreZeroValues()` so an empty `nextCursor` serializes as an
  absent field.
- Pull any request-scope identifier from the request per
  [`PROTO_DEVELOPMENT_STANDARD.md`](./PROTO_DEVELOPMENT_STANDARD.md); for list, set it
  into the entity list struct.

### 3. Field masks

**Read mask (response narrowing)** — `Filter` keeps only the requested paths:

```go
import "github.com/altessa-s/go-atlas/domain/proto/fieldmask"

func applyReadMask(mask *fieldmaskpb.FieldMask, msg *productpb.Product) {
    if msg == nil || mask == nil || len(mask.GetPaths()) == 0 {
        return
    }
    fieldmask.FromProtoFieldMask(mask).Filter(msg)
}
```

**Update mask (request narrowing)** — `ApplyUpdateMask` clears every field NOT in
the mask on the request message, then the converter maps the surviving fields
onto the patch. **Never hand-write a `mask.Contains(path)` ladder with `pathXxx`
string constants.** The handler reads the etag *before* applying the mask
(masking clears it), and unions `"id"` so the identifier survives (only add
`google.api.field_behavior` annotations per `PROTO_DEVELOPMENT_STANDARD.md`
when a field needs `REQUIRED`/`IMMUTABLE`/`OUTPUT_ONLY` semantics beyond this):

```go
func (h *Handler) Update(ctx context.Context, in *productpb.ProductUpdateRequest) (*productpb.Product, error) {
    update := &entities.ProductUpdate{Etag: optionalString(in.GetOptions().GetEtag())}

    if pbMask := in.GetOptions().GetUpdateMask(); pbMask != nil {
        fm := fieldmask.FromProtoFieldMask(pbMask).Union(fieldmask.FromPaths("id"))
        if err := fm.ApplyUpdateMask(in); err != nil {
            return nil, status.Error(codes.InvalidArgument, err.Error())
        }
    }
    converter.Convert(in, update, converter.WithIgnoreNilValues())

    p, err := h.svc.Update(ctx, update)
    if err != nil {
        return nil, MapError(err)
    }
    return converter.Convert(p, &productpb.Product{}, withUnixTimeCodec), nil
}
```

No mask → no `ApplyUpdateMask` call; the converter copies every present (non-nil)
optional field, which is the "apply everything sent" contract. The proto's
`optional` fields and the patch's `*T` fields line up, so `WithIgnoreNilValues`
leaves unsent fields as `nil` ("unchanged"). This is why enums must be
int-aligned ([Entities § Enums](#5-enums--int-typed-and-value-aligned-with-the-proto)) —
the converter carries `*proto.Status → *entities.Status` with no switch.

### 4. Error mapping

A single `MapError` per handler package maps domain sentinels to gRPC status
codes with `errors.Is`. Default is `codes.Internal` with an opaque message.

```go
func MapError(err error) error {
    switch {
    case err == nil:
        return nil
    case errors.Is(err, errs.ErrProductNotFound):
        return status.Error(codes.NotFound, errs.ErrProductNotFound.Error())
    case errors.Is(err, errs.ErrProductSKUConflict):
        return status.Error(codes.AlreadyExists, errs.ErrProductSKUConflict.Error())
    case errors.Is(err, errs.ErrStaleEntity):
        return status.Error(codes.Aborted, errs.ErrStaleEntity.Error())
    case errors.Is(err, errs.ErrInvalidArgument):
        return status.Error(codes.InvalidArgument, errs.ErrInvalidArgument.Error())
    case errors.Is(err, errs.ErrNotImplemented):
        return status.Error(codes.Unimplemented, errs.ErrNotImplemented.Error())
    default:
        return status.Error(codes.Internal, "internal error")
    }
}
```

---

## Errors

- Sentinels live in `internal/errs/errors.go` as `errors.New("lower case message")`.
- Wrap operational errors with `coreerrs.WrapOperation(err, "verb phrase")`.
- Attach a sentinel + cause with `fmt.Errorf("%w: %w", errs.ErrSentinel, cause)`,
  or a sentinel + static context with `fmt.Errorf("%w: %s", errs.ErrSentinel, value)`.
- Classify for control flow with `coreerrs.Is*` (`IsNetworkError`,
  `IsContextDeadlineExceeded`, …) — **never** `strings.Contains(err.Error(), ...)`.
- Compare with `errors.Is` / `errors.As`, never string equality.

---

## Options (optgen)

Configurable packages use generated functional options. Never expose a public
`Config` struct, never hand-write `WithXxx`.

```go
//go:generate go run github.com/altessa-s/go-atlas/cmd/optgen generate

type options struct {
    timeout time.Duration `optgen:"default=DefaultTimeout"`
    retries int           `optgen:"default=DefaultRetries"`
}

const (
    DefaultTimeout = 30 * time.Second
    DefaultRetries = 3
)
```

- Mandatory args stay positional: `New(logger, deps, opts...)`.
- Override the option name with `opt:"Name"` (e.g. `ttl` → `WithTTL`).
- Run `go generate ./...` and commit `options.go` + `options_gen.go` together.
- Migrate hand-written constructors to optgen on touch.

---

## Optional and Result Types

Use `core/types/optional` and `core/types/result`:

- `optional.Optional[T]` for struct fields that may be unset and storage-lookup
  semantics — not for function parameters or returns (keep idiomatic
  `(T, bool)` / `(T, error)`), and not for structs >64 bytes.
- `result.Result[T]` only where tuples are awkward — channel elements,
  per-item batch results.
- Convert at the storage boundary with `optional.ToPtr` / `optional.FromPtr`.
- Never compare `opt == optional.None[T]()`; use `opt.IsNone()`.

---

## go-atlas Reuse Map

| Need | Use | Never |
|------|-----|-------|
| Struct → struct mapping | `domain/converter.Convert` | hand-written `entityToModel` |
| Mongo insert document | `mongo.ConvertToNewDocument` | hand-built `bson.M` |
| Mongo update document | `mongo.ConvertToUpdateDocument` | hand-built `$set` mirror |
| Unix-second timestamp codec | `domain/converter/codec/unixtime` | manual `.Unix()` / `time.Unix(...)` conversions |
| Field masks | `domain/proto/fieldmask` | manual path walking |
| Cursor pagination | `data/mongo.ListCursor` | manual skip/limit + cursor encode |
| Duplicate-key detection | `data/mongo.IsErrorDuplicate` | `strings.Contains(err, "E11000")` |
| Error wrapping | `core/errors.WrapOperation` | `fmt.Errorf("...: %w")` (new code) |
| Error classification | `core/errors.Is*` | `strings.Contains(err.Error(), ...)` |
| Retry w/ backoff | `core/retry.Do` | `for` + `time.Sleep` |
| Fan-out over slice | `core/runtime/concurrency.Process` / `ProcessCollect` | `errgroup` + `semaphore` |
| Panic recovery in goroutines | `core/runtime/panics.Handle` | bare `recover()` |
| Timeout management | `core/context.ApplyTimeout` / `datamongo.DefaultQueryTimeout` | ad-hoc `context.WithTimeout` constants |
| Slice/map helpers | `core/collections/slices` (`coreslices.To`), `core/collections/maps` | manual loops |
| Pointer helpers | `core/types/ptr` | inline `&x` temporaries |
| Nullable fields | `core/types/optional` | bare `*T` in structs |
| Logger | `observability/slog` (`slogx.Module`) | `log` / bare `slog.Default()` |
| HTTP response cleanup | `transport/http/client.CloseBody` | `resp.Body.Close()` w/o drain |
| Normalization | `domain/normalizer.Normalize` | manual `strings.TrimSpace` |

If a helper is missing from go-atlas, raise it — do not silently hand-roll a
local copy.

---

## Documentation Standard

**Minimal necessary documentation.** Comments earn their place by explaining
intent or a non-obvious invariant; they are not a transcript of the code.

### Godoc

- Every exported symbol has a doc comment starting with its name.
- Keep it to **1–3 lines.** State what it does and any caller obligation
  (e.g. "The caller rolls p via BeforeUpdate; oldEtag gates the write.").
- **Do not enumerate errors** in godoc — the sentinel name in a single clause
  is enough ("Returns [errs.ErrProductSKUConflict] when the SKU exists.").
- Do not restate the signature ("Takes a context and an id and returns a
  product" adds nothing).

```go
// ✅ CORRECT
// SoftDelete hides a product by stamping deleted_at; the caller supplies the
// new timestamp and etag, so this method generates nothing.

// ❌ WRONG — restates the signature, enumerates the obvious
// SoftDelete is a method on Storage that takes a context.Context and a pointer
// to an entities.SoftDelete and returns an error. It deletes a product. It may
// return ErrStaleEntity, ErrProductNotFound, or a wrapped storage error, or nil
// on success...
```

### Inline comments

- Add an inline comment only for a non-obvious decision: why the converter is
  bypassed, why a retry exists, a concurrency subtlety, an index invariant.
- Delete comments that narrate the next line (`// loop over items`).
- No commented-out code. No `// TODO` without a tracked issue reference.

### Language

- US English only in the source tree: identifiers, comments, godoc, log
  strings, user-facing errors.
- Lower-case, no trailing period, in error sentinel messages and
  `WrapOperation` phrases.

---

## Templates

### 1. Storage Insert/Update (converter-driven)

```go
func (s *Storage) Insert(ctx context.Context, e *entities.Resource) error {
    ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)
    defer cancel()

    m := converter.Convert(e, &model{}, converter.WithHandleEmbeddedStructs(true))
    m.CursorId = bson.NewObjectID()
    doc, err := s.db.ConvertToNewDocument(ctx, m)
    if err != nil {
        return coreerrs.WrapOperation(err, "convert resource document")
    }
    if _, err := s.collection.InsertOne(ctx, doc); err != nil {
        if dupErr := classifyDuplicate(err); dupErr != nil {
            return dupErr
        }
        return coreerrs.WrapOperation(err, "insert resource")
    }
    return nil
}

func (s *Storage) Update(ctx context.Context, e *entities.Resource, oldEtag string) error {
    ctx, cancel := context.WithTimeout(ctx, datamongo.DefaultQueryTimeout)
    defer cancel()

    update, err := s.db.ConvertToUpdateDocument(ctx, converter.Convert(e, &model{}))
    if err != nil {
        return coreerrs.WrapOperation(err, "build update document")
    }
    filter := activeOnly(bson.M{FieldID: e.ID})
    if oldEtag != "" {
        filter[FieldEtag] = oldEtag
    }
    res, err := s.collection.UpdateOne(ctx, filter, update)
    if err != nil {
        if dupErr := classifyDuplicate(err); dupErr != nil {
            return dupErr
        }
        return coreerrs.WrapOperation(err, "update resource")
    }
    if res.MatchedCount == 0 {
        return errs.ErrStaleEntity
    }
    return nil
}
```

### 2. Service mutate

```go
func (s *Service) Update(ctx context.Context, in *entities.ResourceUpdate) (*entities.Resource, error) {
    _ = normalizer.Normalize(in) //nolint:errcheck

    e, err := s.storage.Get(ctx, in.ID)
    if err != nil {
        return nil, err
    }
    if in.Etag != nil && *in.Etag != e.Etag {
        return nil, errs.ErrStaleEntity
    }
    oldEtag := e.Etag
    in.Merge(e)
    e.BeforeUpdate()
    if err := s.storage.Update(ctx, e, oldEtag); err != nil {
        return nil, err
    }
    return e, nil
}
```

### 3. Handler mutate

```go
func (h *Handler) Update(ctx context.Context, in *resourcepb.ResourceUpdateRequest) (*resourcepb.Resource, error) {
    update := &entities.ResourceUpdate{Etag: optionalString(in.GetOptions().GetEtag())}

    if pbMask := in.GetOptions().GetUpdateMask(); pbMask != nil {
        fm := fieldmask.FromProtoFieldMask(pbMask).Union(fieldmask.FromPaths("id"))
        if err := fm.ApplyUpdateMask(in); err != nil {
            return nil, status.Error(codes.InvalidArgument, err.Error())
        }
    }
    converter.Convert(in, update, converter.WithIgnoreNilValues())

    e, err := h.svc.Update(ctx, update)
    if err != nil {
        return nil, MapError(err)
    }
    return converter.Convert(e, &resourcepb.Resource{}, withUnixTimeCodec), nil
}
```

---

## Review Checklist

### ✅ Architecture
- [ ] Interface in parent package, impl in subpackage; `var _ Iface = (*Impl)(nil)`.
- [ ] No layer skips: transport↔service↔storage; no proto in service/storage, no bson in service/transport.
- [ ] bson `model` is private to the `mongo` package.
- [ ] Every request-scoped read/write is scoped, where the domain has a scope.

### ✅ Converters & go-atlas (the headline rules)
- [ ] No hand-written `entityToModel` / `modelToEntity` — `converter.Convert` everywhere.
- [ ] Inserts use `ConvertToNewDocument`; updates use `ConvertToUpdateDocument`.
- [ ] Any hand-built `$set` is justified in one comment (partial/embedded-array write).
- [ ] No manual `errgroup`/`semaphore`, retry loops, or `strings.Contains(err...)`.
- [ ] Slice mapping uses `coreslices.To`; pointers use `core/types/ptr`.

### ✅ Entities
- [ ] Mutable aggregate has `Etag string` + `DeletedAt *time.Time`.
- [ ] Standard helper set present and identical: `XxxNew`, `UpdateTimestamps`, `UpdateEtag`, `BeforeUpdate`.
- [ ] Patch types carry `normalize` tags and a converter-based `Merge`.
- [ ] List input is one struct with the scope inside; `List(ctx, in *XxxList)`.

### ✅ Etag / OCC / soft delete
- [ ] etag produced only in `XxxNew` (create) and `BeforeUpdate` (update/delete).
- [ ] Storage never calls `etag.Generate()` or `time.Now()` for OCC/timestamps.
- [ ] All deletes are soft (`SoftDelete` / `BulkSoftDelete`), reads use `activeOnly`. No hard-delete path exists for any aggregate.
- [ ] Unique indexes that free a key on delete are partial on `deleted_at`.

### ✅ Storage
- [ ] Every method applies `datamongo.DefaultQueryTimeout`.
- [ ] Field-name constants used; no inline bson key strings.
- [ ] `mongo.ErrNoDocuments` → domain not-found sentinel.
- [ ] Duplicate keys classified via `IsErrorDuplicate`, not string matching.
- [ ] Cursor errors → `errs.ErrInvalidListCursor`.
- [ ] Immutable fields tagged `omitonupdate`; `deleted_at` stored as explicit null (no `omitempty`).
- [ ] Partial indexes use `{deleted_at: null}`, never `$exists: false`.

### ✅ Service
- [ ] Service holds only its own `Storage`; foreign-entity access goes through that entity's `Service` (or a narrow consumer-owned interface it satisfies) — no unjustified foreign `Storage` dependency (see Services Layer §1).
- [ ] `normalizer.Normalize(in)` first on mutating inputs.
- [ ] Dependencies validated before write (e.g. referenced aggregate existence).
- [ ] `oldEtag` captured before `Merge`/`BeforeUpdate`.
- [ ] Logger is `slogx.Module("service:<name>")`; failures logged with `slogx.Error`.
- [ ] No transport concerns (masks, includeTotalCount) in the service.

### ✅ Transport
- [ ] proto⇄entity via `converter.Convert`; `unixtime` codec for timestamps (per `PROTO_DEVELOPMENT_STANDARD.md`, never `google.protobuf.Timestamp`).
- [ ] List responses use `WithIgnoreZeroValues`.
- [ ] Read mask via `fieldmask.Filter`; update mask via `fieldmask.ApplyUpdateMask` + converter
      (union `"id"`, read etag before masking) — never a `mask.Contains(path)` ladder or `pathXxx` constants.
- [ ] One `MapError`; sentinels via `errors.Is`; default `codes.Internal` opaque.

### ✅ Errors / Options
- [ ] Sentinels in `internal/errs`; wrapped with `coreerrs.WrapOperation`.
- [ ] Options via optgen; no public `Config`, no hand-written `WithXxx`; `options_gen.go` committed.

### ✅ Documentation
- [ ] Godoc 1–3 lines, no error enumeration, no signature restatement.
- [ ] Inline comments only for non-obvious decisions; no narration, no commented-out code.
- [ ] US English throughout the source tree.

### ✅ Verification
- [ ] `go generate ./...` after touching `options.go`.
- [ ] `make build`, `make lint` (after `make fmt`), `go vet ./...`, `go test ./... -race -count=1`.
- [ ] Pre-existing failures reported separately from regressions.

---

## Conclusion

This standard ensures:

1. **Reuse over reinvention** — go-atlas (`converter`, `ConvertTo*Document`,
   `ListCursor`, `concurrency`, `retry`, `errors`) is mandatory; hand-rolled
   equivalents are defects.
2. **One consistent OCC model** — etag is produced in the entity layer
   (`XxxNew` / `BeforeUpdate`); storage never generates it.
3. **Uniform soft delete** — every aggregate is hidden, never removed; partial
   indexes free keys on delete.
4. **Clean layering** — proto in transport, entities in service, bson in
   storage; converters bridge them without hand-written mapping.
5. **Minimal, intentional documentation** — comments explain why, not what.
6. **Wire-contract alignment** — timestamps, field masks, and options follow
   `PROTO_DEVELOPMENT_STANDARD.md` exactly, so the converter can bridge proto
   and entity without per-field special cases.

When writing or reviewing service code, follow this standard strictly. Changes
to the standard require review and approval.
