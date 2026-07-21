# ТЗ — MyBusiness CRM

> Ревизия исходного ТЗ с учётом уточнений. Изменения отмечены пометкой **[изменено]**. Места, где решение принято как best practice (без явного указания), отмечены **[best practice]** с обоснованием.
>
> **Изменения в 1.1** (по факту реализации, см. `internal/entities/`): каталог товаров получил трёхуровневую иерархию Product → ProductVariant → ProductSKU (раздел 4/4a/4b), справочник характеристик `ProductAttributeDefinition` (раздел 4c), Price/Inventory/InventoryMovement теперь ключуются по `SKUID`, а не `ProductID` (разделы 5-7), `Sale.WarehouseID` перенесён с уровня продажи на уровень позиции (раздел 9), Partner получил раздельные `CommissionPercentage`/`DiscountPercentage` вместо единого `Percentage`, плюс публичные поля `Address/Website/IsPublic` (раздел 10), и добавлен сервис уведомлений (раздел 15). Разделы ниже обновлены под текущее состояние кода; там, где решение отличается от исходной версии 1.0, это отмечено **[v1.1]**.

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

## 4. Products **[v1.1 — см. также 4a/4b/4c]**

Каталог — не плоский `Product`, а трёхуровневая иерархия
`Product → ProductVariant → ProductSKU`. Причина: одному товару
соответствует несколько визуальных исполнений (цвет/фактура — `ProductVariant`,
со своим набором фото), а каждое исполнение продаётся в нескольких
покупаемых вариантах (размер/фасовка — `ProductSKU`, со своей ценой и
остатком). Исходная плоская модель (`Product.SKU`, `Product.ImageIDs`)
из ревизии 1.0 **не реализована** — вместо неё:

```go
type Product struct {
    ID uuid.UUID

    Name        LocalizedString
    Description LocalizedString

    BrandID     uuid.UUID
    CategoryIDs []uuid.UUID // many-to-many, категории параллельны друг другу

    // Details — характеристики, общие для ВСЕХ вариантов товара (материал,
    // коллекция и т.п.). Характеристики, различающие варианты по внешнему
    // виду (цвет, фактура), живут на ProductVariant.Attributes; влияющие
    // на цену/доступность (размер, толщина) — на ProductSKU.Attributes.
    Details map[string]LocalizedString

    // HasStock — вычисляется системой (пересчитывается при каждом
    // InventoryMovement, затрагивающем любой SKU любого варианта товара),
    // не устанавливается через API. Существует только для сортировки
    // списка товаров "сначала в наличии".
    HasStock bool

    Status ProductStatus // Draft | Active | Inactive

    CreatedAt int64
    UpdatedAt int64
    DeletedAt *int64
    ETag      string
}
```

### Methods
Create, Get, List, Update, SoftDelete

### Logic
- `Product` не хранит SKU, изображения, цену или остаток напрямую — это карточка-контейнер над `ProductVariant`/`ProductSKU`.
- `CategoryIDs` — товар может принадлежать нескольким категориям одновременно.
- `Details` — общие для всех вариантов характеристики, значения локализуются.

---

## 4a. Product Variants **[v1.1, новая сущность]**

`ProductVariant` — визуальное исполнение товара: один набор изображений, один набор внешних характеристик (цвет, фактура, узор).

```go
type ProductVariant struct {
    ID        uuid.UUID
    ProductID uuid.UUID

    Attributes map[string]LocalizedString // цвет/фактура/узор — влияет на внешний вид
    ImageIDs   []string                    // порядок показа, первый — главный

    Status ProductVariantStatus // Draft | Active | Inactive

    CreatedAt int64
    UpdatedAt int64
    DeletedAt *int64
    ETag      string
}
```

### Methods
Create, Get, List, Update, SoftDelete

### Logic
- Изображения (как в исходной ревизии 1.0) прикрепляются через HTTP upload-эндпоинт, возвращающий `id`; клиент прикрепляет `id` к `ProductVariant.ImageIDs` через `Update` — сам механизм загрузки не изменился, просто переехал с `Product` на `ProductVariant`.
- `ProductID` неизменяем после создания.

---

## 4b. Product SKUs **[v1.1, новая сущность]**

`ProductSKU` — покупаемая единица: одна цена (`ProductPrice`), один остаток (`Inventory`) на склад, характеристики, влияющие на цену/доступность (размер, толщина, фасовка).

```go
type ProductSKU struct {
    ID        uuid.UUID
    VariantID uuid.UUID

    SKU string // уникален среди активных записей — заменяет Product.SKU из ревизии 1.0

    Attributes map[string]LocalizedString // размер/толщина/фасовка — влияет на цену/доступность

    Status ProductSkuStatus // Draft | Active | Inactive

    CreatedAt int64
    UpdatedAt int64
    DeletedAt *int64
    ETag      string
}
```

### Methods
Create, Get, List, Update, SoftDelete

### Logic
- `VariantID`/`SKU` неизменяемы после создания.
- Цена (раздел 5) и остаток (раздел 6) ключуются по `SKUID`, не по `ProductID`/`VariantID`.

---

## 4c. Product Attribute Definitions **[v1.1, новая сущность]**

Справочник допустимых ключей `Product.Details`/`ProductVariant.Attributes`/`ProductSKU.Attributes` (материал, цвет, размер и т.п.).

```go
type ProductAttributeDefinition struct {
    ID    uuid.UUID
    Key   string
    Label LocalizedString

    IsPublic  bool  // ключ можно показывать на публичном сайте
    SortOrder int32

    CreatedAt int64
}
```

### Methods
List (read-only на уровне API)

### Logic
- **Нет CRUD через API** — записи заводятся напрямую в БД разработчиком, в отличие от всех остальных сущностей документа. Не проектировать для этого админ-форму.
- `IsPublic` — публичный сайт отдаёт только те ключи `Product.Details`, у которых `IsPublic = true`, остальные скрываются из каталога-витрины.

---

## 5. Prices **[v1.1: ключ — SKUID, не ProductID]**

```go
type ProductPrice struct {
    ID uuid.UUID

    SKUID uuid.UUID // [v1.1] было ProductID — один товар может иметь несколько SKU с разной ценой

    Price    decimal.Decimal
    Currency Currency

    DiscountPrice *decimal.Decimal

    CreatedAt int64
    UpdatedAt int64

    ETag string
}

type Currency string
// значение читается из конфига при старте сервиса, напр. Currency = "RSD"
```

### Methods
Create, Get, Update, History

### Logic
- Система работает с **одной валютой**, заданной в конфиге. Поле `Currency` в модели сохраняется для истории/отчётности, но не выбирается пользователем при создании цены.
- Не более одной активной цены на `SKUID`; прежние значения — в истории.
- `decimal.Decimal` при передаче через gRPC/хранении в Mongo сериализуется как **строка** (во избежание потери точности).

---

## 6. Inventory **[v1.1: ключ — (SKUID, WarehouseID), не (ProductID, WarehouseID)]**

```go
type Inventory struct {
    ID uuid.UUID

    SKUID       uuid.UUID // [v1.1] было ProductID
    WarehouseID uuid.UUID

    Quantity int

    UpdatedAt int64
    ETag      string // горячая точка гонок записи
}
```

### Methods
Get, List

### Logic
- Используется для отображения наличия и проверки перед продажей.
- **Уникальный индекс на (SKUID, WarehouseID)** — не может быть двух записей остатка для одной пары SKU/склад.
- **Кешируется в Redis** для быстрого чтения. Инвалидация кеша — при каждом `InventoryMovement`.

---

## 7. Inventory Movements **[v1.1: ключ — SKUID, не ProductID]**

```go
type InventoryMovement struct {
    ID uuid.UUID

    SKUID       uuid.UUID // [v1.1] было ProductID
    WarehouseID uuid.UUID

    Type MovementType
    // Receipt
    // Sale         — создаётся только сервером внутри SalesService.Create, не руками
    // WriteOff
    // Adjustment
    // Transfer

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
- При изменении остатка: создать `InventoryMovement` → обновить `Inventory` → инвалидировать Redis-кеш; также пересчитывает `Product.HasStock` для затронутого товара.
- Пример: приход `+100 Receipt`, продажа `-5 Sale`, перенос между складами — пара движений `Transfer` (списание с одного, приход на другой).

---

## 8. Clients

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

## 9. Sales **[v1.1: WarehouseID перенесён с продажи на позицию; ProductID → SKUID]**

```go
type Sale struct {
    ID uuid.UUID

    Number    int64      // [v1.1] человекочитаемый номер, назначается атомарно storage-слоем при вставке
    ClientID  uuid.UUID
    PartnerID *uuid.UUID // опционален — продажа может быть без партнёра

    Items []SaleItem

    TotalAmount int64 // [v1.1] basis points (было decimal.Decimal), сумма позиций после скидки

    Status SaleStatus

    CreatedBy uuid.UUID

    CreatedAt int64
    UpdatedAt int64

    DeletedAt *int64 // всегда nil — Sale не имеет пути soft-delete, поле хранится только для паритета со схемой
    ETag      string
}

type SaleItem struct {
    SKUID uuid.UUID // [v1.1] было ProductID — продажа списывает конкретный SKU, не абстрактный товар

    Quantity int64
    // PriceAmount фиксируется из текущей ProductPrice SKU в момент
    // создания (basis points) — последующие изменения цены не затрагивают
    // уже созданную продажу.
    PriceAmount int64

    DiscountPercentage int32 // всегда 0-100, никогда фиксированная сумма

    // WarehouseID — [v1.1] перенесено с уровня Sale на уровень позиции:
    // разные позиции одной продажи могут списываться с разных складов.
    WarehouseID uuid.UUID
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
- **[v1.1]** Продажа списывается **по позициям**, каждая — со своего склада (не с одного склада целиком, как в ревизии 1.0) — нужно для сценария "продать N штук, автоматически распределив между складами по остатку".
- Продажа может быть без партнёра (обычная розница), с партнёром-рефералом (партнёр приводит клиента — `ClientID` и `PartnerID` оба заданы), или напрямую партнёру (`PartnerID` без `ClientID`, партнёр покупает сам) — три валидных формы, см. `Partner.CommissionPercentage`/`DiscountPercentage` в разделе 10.
- `DiscountPercentage` в `SaleItem` — всегда в процентах. При прямой продаже партнёру (`PartnerID` без `ClientID`) сервер **принудительно подставляет** `Partner.DiscountPercentage`, переопределяя запрошенное значение позиции — доверять вводу оператора в этом случае нельзя.
- Клиент ищется/создаётся по email (find-or-create) прямо внутри `Create` — отдельного шага "сначала создать клиента" на API нет.
- Позиции неизменяемы после создания (нет метода Update) — ошибку исправляет `Cancel`, не редактирование.
- При создании продажи: на каждую позицию — проверить наличие на её `WarehouseID` → создать Sale → создать `InventoryMovement` (Sale) для каждой позиции → уменьшить `Inventory` → инвалидировать Redis-кеш. Это не кросс-коллекционная транзакция: если движение по одной из последующих позиций не удаётся записать, уже обработанные позиции остаются списанными, а ошибка логируется как `ErrorContext` (осознанный компромисс, см. `internal/services/sale/sale/service.go`).
- `Cancel` восстанавливает остаток по всем позициям тем же best-effort механизмом (движение типа `Adjustment`, а не `Sale`).

---

## 10. Partners **[v1.1: единый Percentage разделён на два поля + публичные поля]**

```go
type Partner struct {
    ID uuid.UUID

    Name  string
    Phone string // unique
    Email string

    // CommissionPercentage — [v1.1] выплата партнёру, когда он привёл
    // клиента (Sale.ClientID и Sale.PartnerID оба заданы); применяется к
    // Sale.TotalAmount, но не как скидка клиенту, а как начисление партнёру.
    CommissionPercentage int32
    // DiscountPercentage — [v1.1, новое поле] скидка, применяемая
    // автоматически ко всем позициям, когда партнёр покупает напрямую
    // (Sale.PartnerID задан, Sale.ClientID нет) — см. sale.Service.buildItems.
    DiscountPercentage int32

    Note string

    Address string // [v1.1] не локализуется — факт, не каталожный контент
    Website string // [v1.1, новое поле]

    // IsPublic — [v1.1, новое поле] управляет видимостью в публичном
    // списке дилеров на сайте-витрине (PartnersService.ListPublic).
    IsPublic bool

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
Create, Get, List, Update, SoftDelete, **ListPublic** [v1.1, новый метод]

### Logic
- `Phone` уникален глобально.
- Исходная ревизия описывала один `Percentage`; по факту реализации это две независимые ставки, применяемые в разных сценариях продажи (см. раздел 9) — комиссия за реферала и скидка при прямой покупке не одно и то же и не взаимоисключающи по смыслу, поэтому разделены.
- `ListPublic` — урезанная read-only выборка (`IsPublic = true` активные партнёры) для дилерской карты на публичном сайте (см. `TZ_PHOMI_Public_Site_v1.md`); не требует аутентификации.

---

## 11. Reports (computed)

Отдельных коллекций нет. Данные собираются из `Sales`, `SaleItems`, `InventoryMovements`.

- Отчёты **вычисляются "на лету" из MongoDB** (агрегационные пайплайны), не кешируются в Redis — в отличие от `Inventory`, где важна скорость чтения **[изменено, зафиксировано]**.

Отчёты:
- продажи за период;
- продажи по сотрудникам, партнёрам;
- популярные товары;
- оборот;
- остатки.

---

## 12. Warehouses

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

## 13. Списки и пагинация [изменено, новый раздел]

Все методы `List` во всех сущностях поддерживают:
- пагинацию (формат — limit/offset или cursor — будет предоставлен отдельно заказчиком);
- фильтрацию по основным полям (Status, даты создания и т.п.);
- сортировку.

Точный контракт (proto-схема параметров) фиксируется отдельным документом, здесь только требование о наличии.

---

## 14. Redis — назначение [изменено, новый раздел]

Redis используется **только для кеша `Inventory`** (быстрое чтение остатков). Отчёты и прочие сущности в Redis не кешируются — источник истины для них MongoDB.

Инвалидация кеша Inventory — синхронно при каждой записи `InventoryMovement`.

---

## 15. Notifications **[v1.1, новый раздел — отсутствовал в ревизии 1.0]**

Исходящая почта (`internal/services/notification`, `internal/pkg/mailer`) — не была описана в первой ревизии, реализована по ходу разработки.

### Logic
- Доставка — **SMTP или Resend HTTP API**, выбирается конфигом (`crm.mail`), без изменений в коде вызывающей стороны.
- **Client-key gating**: сервис уведомлений принимает вызовы только с пред-выданным ключом на вызывающий фронтенд (admin/public-site — разные ключи), а не с обычным bearer-токеном пользователя — так вызвать отправку письма не может произвольный аутентифицированный клиент, только доверенный фронтенд-процесс. Ключ передаётся вне обычного RBAC-интерцептора.
- Используется в двух местах: (1) публичный сайт — формы «Стать дилером»/«Контакт» отправляют письмо через этот сервис вместо создания записи в CRM (см. `TZ_PHOMI_Public_Site_v1.md` §7.1); (2) зарезервировано для будущих CRM-уведомлений (не задействовано на момент этой ревизии).

---

## Основные связи (обновлено, v1.1)

```
Category *---* Product
Brand     1---* Product
Product   1---* ProductVariant
ProductVariant 1---* ProductSKU
ProductSKU 1---* ProductPrice   [v1.1 — было Product 1---* ProductPrice]
ProductSKU 1---* Inventory      [v1.1 — было Product 1---* Inventory]
Warehouse 1---* Inventory
Warehouse 1---* InventoryMovement
Client    1---* Sale
Partner   0..1---* Sale
Sale      1---* SaleItem
SaleItem  *---1 ProductSKU      [v1.1 — было SaleItem *---1 Product]
SaleItem  *---1 Warehouse       [v1.1 — было Warehouse 1---* Sale на уровне продажи]
User      1---* Sale (CreatedBy)
User      1---* InventoryMovement (CreatedBy)
```
