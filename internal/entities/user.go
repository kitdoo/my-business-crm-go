package entities

import (
	"time"

	"github.com/google/uuid"

	"github.com/altessa-s/go-atlas/domain/converter"
)

// UserRole is int32-aligned with crm.types.user.UserRole so
// converter.Convert maps it both ways as a plain scalar.
type UserRole int32

const (
	UserRoleUnspecified UserRole = iota
	UserRoleAdmin
	UserRoleEmployee
	UserRoleGuest
)

// UserStatus is int32-aligned with crm.types.user.UserStatus so
// converter.Convert maps it both ways as a plain scalar.
type UserStatus int32

const (
	UserStatusUnspecified UserStatus = iota
	UserStatusActive
	UserStatusInactive
)

// User is a staff account. PasswordHash and TokenHash have no proto
// counterpart on crm.types.user.User and are never serialized to the wire;
// converter.Convert skips fields absent from the destination type.
type User struct {
	ID          string
	Name        LocalizedString
	LastName    LocalizedString
	Phone       string // unique
	Email       string // unique
	Description LocalizedString
	Role        UserRole
	Status      UserStatus
	// PasswordHash is a bcrypt hash; never logged or returned.
	PasswordHash string
	// TokenHash is the SHA-256 hex digest of the current session token
	// (see services/user.Service.Login); nil when the user has no active
	// session. The raw token itself is never stored.
	TokenHash *string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time // nil = active
	Etag      string     // OCC token; rolled on every write
}

// UserNew creates a User with a fresh ID, timestamps, and etag.
func UserNew(init ...func(*User)) *User {
	u := &User{ID: uuid.NewString()}
	if len(init) > 0 {
		init[0](u)
	}
	u.UpdateTimestamps()
	u.UpdateEtag()
	return u
}

func (u *User) UpdateTimestamps() {
	now := time.Now().UTC()
	u.UpdatedAt = now
	if u.CreatedAt.IsZero() {
		u.CreatedAt = now
	}
}

func (u *User) UpdateEtag() { u.Etag = uuid.NewString() }

func (u *User) BeforeUpdate() {
	u.UpdateTimestamps()
	u.UpdateEtag()
}

// UserCreate is the Create input. Merge applies it onto a freshly
// constructed User via the converter. PasswordHash is set by the service
// (bcrypt of the plaintext Password) before Merge runs, since it is not a
// field the caller sets directly.
type UserCreate struct {
	Name         LocalizedString
	LastName     LocalizedString
	Phone        string `normalize:"trim"`
	Email        string `normalize:"trim"`
	Description  LocalizedString
	Role         UserRole
	Password     string `normalize:"trim"`
	PasswordHash string
}

func (c *UserCreate) Merge(dst *User) *User {
	if c == nil || dst == nil {
		return dst
	}
	converter.Convert(c, dst, converter.WithIgnoreNilValues())
	return dst
}

// UserUpdate is the Update input. Nil fields mean "leave unchanged".
type UserUpdate struct {
	ID          string `normalize:"trim"`
	Name        LocalizedString
	LastName    LocalizedString
	Phone       *string `normalize:"trim,nil_on_empty"`
	Email       *string `normalize:"trim,nil_on_empty"`
	Description LocalizedString
	Role        *UserRole
	Status      *UserStatus
	Etag        *string `normalize:"trim,nil_on_empty"` // client OCC precondition
}

func (u *UserUpdate) Merge(dst *User) *User {
	if u == nil || dst == nil {
		return dst
	}
	converter.Convert(u, dst, converter.WithIgnoreNilValues())
	return dst
}

// UserDelete is the Delete input.
type UserDelete struct {
	ID   string  `normalize:"trim"`
	Etag *string `normalize:"trim,nil_on_empty"`
}

// UserLogin is the Login input. Login is a phone or email.
type UserLogin struct {
	Login    string `normalize:"trim"`
	Password string `normalize:"trim"`
}

// UserChangePassword is the ChangePassword input.
type UserChangePassword struct {
	ID              string `normalize:"trim"`
	CurrentPassword string `normalize:"trim"`
	NewPassword     string `normalize:"trim"`
}

// UsersListSortField is int32-aligned with UsersListRequest.Sort.Field.
type UsersListSortField int32

const (
	UsersListSortFieldCreatedAt UsersListSortField = iota
	UsersListSortFieldName
)

type UsersListSort struct {
	Field     UsersListSortField
	Direction SortDirection
}

// UsersList is the single List input; scope/filters/sort/pagination all
// live inside it, per the List(ctx, in *XxxList) convention.
type UsersList struct {
	Statuses          []UserStatus
	Roles             []UserRole
	CreatedAt         *PeriodFilter
	Sort              UsersListSort
	Pagination        ListPagination
	IncludeTotalCount bool
}
