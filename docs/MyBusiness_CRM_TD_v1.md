# ТЗ — MyBusiness CRM

> Ревизия исходного ТЗ с учётом уточнений. Изменения отмечены пометкой **[изменено]**. Места, где решение принято как best practice (без явного указания), отмечены **[best practice]** с обоснованием.

## Стек
**Backend:** Go, MongoDB, Redis, Docker, Protobuf/gRPC
**Frontend:** Vue, Nuxt, JavaScript
**Storage:** изображения — локально на диске сервера

## Общие требования

Все сущности поддерживают:
- UUID как ID
- CreatedAt, UpdatedAt
- Soft Delete
- ETag для optimistic locking

**Исключения из общего правила [изменено]:**
- `InventoryMovement` — неизменяемый лог. Нет `UpdatedAt`, `DeletedAt`, `ETag`, нет методов `Update`/`SoftDelete`.
- `ProductImage` — дочерняя сущность, hard delete. Нет `DeletedAt`/`ETag`.

**ETag добавлен туда, где раньше отсутствовал:**
- `Inventory` — критично, т.к. это самая горячая точка гонок записи (несколько продаж одновременно меняют остаток).
- `ProductPrice` — добавлен ETag.

Аудит изменений (кто и когда менял Product/Price/Sale.Status) — **не требуется на данном этапе.**

### Мультиязычность

```
type LocalizedString map[string]string
```

- Обязательный язык: **`sr`** (сербский). Остальные (`en`, `ru`, ...) — опциональны.
- Валидация: при создании/обновлении сущности с `LocalizedString`-полями ключ `sr` обязателен, иначе ошибка валидации.
- Fallback при отсутствии запрошенного языка: **`sr` как дефолтный** (т.к. он гарантированно есть).

**[best practice] Какие поля остаются `LocalizedString`, а какие — `string`:**
Общее правило — локализуются поля, которые являются **описательным/каталожным контентом**, показываемым конечному пользователю на разных языках (названия товаров, категорий, брендов, складов). Поля, которые являются **фактическими данными или служебными заметками** (имя клиента, номер телефона, физический адрес, внутренние примечания сотрудников), не переводятся — перевод имени человека или адреса не имеет смысла.

| Сущность | Поле | Тип | Обоснование |
|---|---|---|---|
| Client | Name | `string` | Имя — факт, не переводится |
| Client | Address | `string` | Адрес не переводится |
| Client | Note | `string` | Внутренняя заметка сотрудника, не клиентский контент |
| Partner | Name | `string` | Имя/название партнёра — факт |
| Partner | Note | `string` | Внутренняя заметка |
| Warehouse | Address | `string` | Адрес не переводится |
| Warehouse | Name, Description | `LocalizedString` | Уже каталожный контент (без изменений) |

Все остальные `Name`/`Description` (Product, Category, Brand) остаются `LocalizedString` без изменений.

---

## 1. Users

```go
type User struct {
    ID       uuid.UUID

    Name     LocalizedString
    LastName LocalizedString

    Phone string // unique
    Email string // unique

    Description LocalizedString

    Role   UserRole
    Status UserStatus

    PasswordHash string

    CreatedAt int64
    UpdatedAt int64
    DeletedAt *int64

    ETag string
}

type UserRole string
const (
    RoleAdmin    UserRole = "admin"
    RoleEmployee UserRole = "employee"
    RoleGuest    UserRole = "guest"
)

type UserStatus string
const (
    UserStatusActive   UserStatus = "active"
    UserStatusInactive UserStatus = "inactive"
)
```

### Methods
Create, Get, List, Update, SoftDelete, Login, ChangePassword

### Logic
- Пароль хранится только как hash, никогда не возвращается наружу.
- `Email` и `Phone` — уникальны глобально (unique index в Mongo).
- **Admin создаётся через конфиг** (bootstrap-администратор), не через обычную регистрацию **[изменено]**.
- Права ролей (`admin`/`employee`/`guest`) **настраиваются через конфиг** — RBAC-таблица "роль → разрешённые разделы/действия", а не жёстко зашитая логика в коде **[изменено]**.
  - Пример конфига на первом этапе: `employee` → доступ к спискам и созданию Products/Categories/Brands/связанного + создание Sales. Остальное (Partners, Users, Warehouses) — только `admin`.
  - Конфиг допускает добавление новых прав в будущем (например, `employee` → создание Partners) без изменения кода.

### Auth
- Базовый **постоянный (не истекающий) токен**, выдаваемый при `Login`. Без refresh-механизма — токен действует до логаута/ручной инвалидации **[изменено]**.

---

## 2. Categories

```go
type Category struct {
    ID uuid.UUID

    Name        LocalizedString
    Description LocalizedString

    ParentID *uuid.UUID

    Status CategoryStatus

    CreatedAt int64
    UpdatedAt int64
    DeletedAt *int64

    ETag string
}

type CategoryStatus string
const (
    CategoryStatusActive   CategoryStatus = "active"
    CategoryStatusInactive CategoryStatus = "inactive"
)
```

### Methods
Create, Get, List, Update, SoftDelete

### Logic
- Поддержка дерева категорий через `ParentID`.
- Нельзя удалить категорию, если есть активные товары.
- Товар может относиться **к нескольким категориям одновременно**, категории при этом равноправны (не строго дерево-принадлежность) **[изменено]** — см. изменения в Product.
- Возможность получить только категории с товарами (метод фильтрации в `List`).

---

## 3. Brands

```go
type Brand struct {
    ID uuid.UUID

    Name        LocalizedString
    Description LocalizedString

    Status BrandStatus

    CreatedAt int64
    UpdatedAt int64
    DeletedAt *int64

    ETag string
}

type BrandStatus string
const (
    BrandStatusActive   BrandStatus = "active"
    BrandStatusInactive BrandStatus = "inactive"
)
```

### Methods
Create, Get, List, Update, SoftDelete

### Logic
Нельзя удалить бренд, если есть товары.

---

## 4. Products

```go
type Product struct {
    ID uuid.UUID

    Name        LocalizedString
    Description LocalizedString

    SKU string // unique

    BrandID     uuid.UUID
    CategoryIDs []uuid.UUID // [изменено] many-to-many вместо одного CategoryID

    Details map[string]LocalizedString // [изменено] значения характеристик локализуются

    Status ProductStatus

    CreatedAt int64
    UpdatedAt int64
    DeletedAt *int64

    ETag string
}

type ProductStatus string
const (
    ProductStatusDraft    ProductStatus = "draft"
    ProductStatusActive   ProductStatus = "active"
    ProductStatusInactive ProductStatus = "inactive"
)
```

### Methods
Create, Get, List, Update, SoftDelete

### Logic
- Product не хранит остатки и не хранит цену — только описание товара.
- `SKU` уникален глобально.
- `CategoryIDs` — товар может принадлежать нескольким категориям, категории параллельны друг другу (не нужно указывать "главную").
- `Details` — характеристики товара (материал, размер и т.п.), значения локализуются, т.к. показываются пользователю.

---

## 5. Product Images

```go
type ProductImage struct {
    ID uuid.UUID

    ProductID uuid.UUID

    FileName string
    Path     string
    MimeType string
    Size     int64

    IsMain    bool
    SortOrder int

    CreatedAt int64
}
```

### Methods
Upload, List, Delete, SetMain

### Logic
- Хранение: файл физически на диске (`/storage/products/{product_id}/...`), в Mongo — только metadata.
- **Допустимые форматы: JPG, PNG** (ограничено через конфиг, список расширяем) **[изменено]**.
- **Максимальный размер файла — задаётся через конфиг** (дефолт предлагается 5 MB, изменяемо без пересборки) **[best practice — дефолтное значение]**.

---

## 6. Prices

```go
type ProductPrice struct {
    ID uuid.UUID

    ProductID uuid.UUID

    Price    decimal.Decimal
    Currency Currency

    DiscountPrice *decimal.Decimal

    CreatedAt int64
    UpdatedAt int64

    ETag string // [изменено] добавлено
}

// [изменено] Валюта одна на всю систему, задаётся через конфиг, а не выбирается per-entity
type Currency string
// значение читается из конфига при старте сервиса, напр. Currency = "RSD"
```

### Methods
Create, Get, Update, History

### Logic
- Система работает с **одной валютой**, заданной в конфиге. Поле `Currency` в модели сохраняется для истории/отчётности, но не выбирается пользователем при создании цены.
- `decimal.Decimal` при передаче через gRPC/хранении в Mongo сериализуется как **строка** (во избежание потери точности) — фиксируем это как конвенцию проекта.

---

## 7. Inventory

```go
type Inventory struct {
    ID uuid.UUID

    ProductID   uuid.UUID
    WarehouseID uuid.UUID

    Quantity int

    UpdatedAt int64
    ETag      string // [изменено] добавлено — горячая точка гонок записи
}
```

### Methods
Get, List

### Logic
- Используется для отображения наличия и проверки перед продажей.
- **Уникальный индекс на (ProductID, WarehouseID)** — не может быть двух записей остатка для одной пары товар/склад **[изменено, зафиксировано явно]**.
- **Кешируется в Redis** для быстрого чтения (частые запросы наличия). Инвалидация кеша — при каждом `InventoryMovement` **[изменено]**.

---

## 8. Inventory Movements

```go
type InventoryMovement struct {
    ID uuid.UUID

    ProductID   uuid.UUID
    WarehouseID uuid.UUID

    Type MovementType
    // Receipt
    // Sale
    // WriteOff
    // Adjustment
    // Transfer — [изменено] добавлен тип для переноса остатков между складами (нужен при деактивации склада, см. п.13)

    Quantity int
    Comment  string

    CreatedBy uuid.UUID

    CreatedAt int64
}
```

### Methods
Create, List, GetHistory

### Logic
- Неизменяемый лог: нет Update/Delete методов и соответствующих полей (см. "Общие требования").
- При изменении остатка: создать `InventoryMovement` → обновить `Inventory` → инвалидировать Redis-кеш.
- Пример: приход `+100 Receipt`, продажа `-5 Sale`, перенос между складами — пара движений `Transfer` (списание с одного, приход на другой).

---

## 9. Clients

```go
type Client struct {
    ID uuid.UUID

    Name string
    Phone string // unique
    Email string

    Address string
    Note    string

    CreatedAt int64
    UpdatedAt int64
    DeletedAt *int64

    ETag string
}
```

### Methods
Create, Get, List, Update, SoftDelete

### Logic
- `Phone` уникален глобально **[изменено, зафиксировано]**.

---

## 10. Sales

```go
type Sale struct {
    ID uuid.UUID

    ClientID    uuid.UUID
    WarehouseID uuid.UUID   // [изменено] обязателен — продажа списывается с одного склада целиком
    PartnerID   *uuid.UUID  // [изменено] опционален — продажа может быть без партнёра

    Items []SaleItem

    TotalPrice decimal.Decimal

    Status SaleStatus

    CreatedBy uuid.UUID

    CreatedAt int64
    UpdatedAt int64

    DeletedAt *int64 // [изменено] добавлено
    ETag      string // [изменено] добавлено
}

type SaleItem struct {
    ProductID uuid.UUID

    Quantity int
    Price    decimal.Decimal

    Discount decimal.Decimal // [изменено] всегда процент (0-100)
}

type SaleStatus string
const (
    SaleStatusDraft     SaleStatus = "draft"
    SaleStatusPaid      SaleStatus = "paid"
    SaleStatusShipped   SaleStatus = "shipped"
    SaleStatusCompleted SaleStatus = "completed"
    SaleStatusCancelled SaleStatus = "cancelled"
    SaleStatusRefunded  SaleStatus = "refunded"
)
```

### Methods
Create, Get, List, UpdateStatus, Cancel

### Logic
- Продажа списывается **с одного склада целиком** (не по позициям) **[изменено, зафиксировано]**.
- Продажа может быть без партнёра (обычная розница) или с партнёром (дилерская) **[изменено, зафиксировано]**.
- Если указан `PartnerID` — при завершении продажи партнёру начисляется **процент от суммы продажи** согласно `Partner.Percentage` (комиссия, а не скидка клиенту) **[изменено, зафиксировано]**.
- `Discount` в `SaleItem` — всегда в процентах, а не фиксированной суммой.
- При создании продажи: проверить наличие на `WarehouseID` → создать Sale → создать `InventoryMovement` (Sale) для каждой позиции → уменьшить `Inventory` → инвалидировать Redis-кеш.

---

## 11. Partners

```go
type Partner struct {
    ID uuid.UUID

    Name string
    Phone string // unique
    Email string

    Percentage int64 // [уточнено] процент комиссии, начисляемый партнёру с суммы продажи

    Note string

    Status PartnerStatus

    CreatedAt int64
    UpdatedAt int64
    DeletedAt *int64

    ETag string
}

type PartnerStatus string
const (
    PartnerStatusActive   PartnerStatus = "active"
    PartnerStatusInactive PartnerStatus = "inactive"
)
```

### Methods
Create, Get, List, Update, SoftDelete

### Logic
- `Phone` уникален глобально **[изменено, зафиксировано]**.
- `Percentage` — фиксированный процент комиссии партнёра, применяется к `Sale.TotalPrice` при продажах с этим партнёром.

---

## 12. Reports (computed)

Отдельных коллекций нет. Данные собираются из `Sales`, `SaleItems`, `InventoryMovements`.

- Отчёты **вычисляются "на лету" из MongoDB** (агрегационные пайплайны), не кешируются в Redis — в отличие от `Inventory`, где важна скорость чтения **[изменено, зафиксировано]**.

Отчёты:
- продажи за период;
- продажи по сотрудникам, партнёрам;
- популярные товары;
- оборот;
- остатки.

---

## 13. Warehouses

```go
type Warehouse struct {
    ID uuid.UUID

    Name        LocalizedString
    Description LocalizedString

    Address string // [изменено] не локализуется, см. раздел "Мультиязычность"

    Status WarehouseStatus
    // active
    // inactive

    CreatedAt int64
    UpdatedAt int64
    DeletedAt *int64

    ETag string
}

type WarehouseStatus string
const (
    WarehouseStatusActive   WarehouseStatus = "active"
    WarehouseStatusInactive WarehouseStatus = "inactive"
)
```

### Methods
Create, Get, List, Update, SoftDelete

### Logic
- Нельзя удалить склад, если на нём есть остатки.
- Неактивный склад нельзя использовать для новых операций.
- **При деактивации склада с ненулевыми остатками — обязателен перенос остатков на другой склад** (`InventoryMovement` типа `Transfer`) перед деактивацией; система блокирует деактивацию, пока `Inventory.Quantity` по складу не станет 0 **[изменено, зафиксировано]**.
- Пользователь может иметь доступ только к определённым складам (опционально, в будущем).

---

## 14. Списки и пагинация [изменено, новый раздел]

Все методы `List` во всех сущностях поддерживают:
- пагинацию (формат — limit/offset или cursor — будет предоставлен отдельно заказчиком);
- фильтрацию по основным полям (Status, даты создания и т.п.);
- сортировку.

Точный контракт (proto-схема параметров) фиксируется отдельным документом, здесь только требование о наличии.

---

## 15. Redis — назначение [изменено, новый раздел]

Redis используется **только для кеша `Inventory`** (быстрое чтение остатков). Отчёты и прочие сущности в Redis не кешируются — источник истины для них MongoDB.

Инвалидация кеша Inventory — синхронно при каждой записи `InventoryMovement`.

---

## Основные связи (обновлено)

```
Category *---* Product
Brand     1---* Product
Product   1---* ProductImage
Product   1---* ProductPrice
Product   1---* Inventory
Warehouse 1---* Inventory
Warehouse 1---* InventoryMovement
Client    1---* Sale
Partner   0..1---* Sale   [изменено — партнёр опционален]
Warehouse 1---* Sale      [изменено — добавлена связь]
Sale      1---* SaleItem
SaleItem  *---1 Product
User      1---* Sale (CreatedBy)
User      1---* InventoryMovement (CreatedBy)
```
