# ТЗ — MyBusiness CRM: Web Admin Panel (Frontend)

> Версия: 1.2. Основано на `docs/MyBusiness_CRM_TD_v1.md` (бэкенд) и текущих контрактах в `proto/` (все сервисы уже реализованы на бэкенде — см. `internal/`). Каждое решение ниже либо прямо следует из proto/бэкенд-ТЗ, либо помечено **[best practice]** с обоснованием, чтобы при реализации не пришлось гадать.
>
> **Изменения в 1.2**: цветовая схема (п.7) и layout (п.8) обновлены по итогам стилистической ревизии заказчика после первой реализации (Brand — эталонная сущность) — см. историю правок. Заменены: синий акцент → бирюзовый `#85C5CF`, тёмное меню — не чёрное, а серое `#333333`, поля форм — явные чёрные рамки. RightDrawer стал плоским списком (без промежуточного клика по разделу), LeftDrawer как навигация по сущностям упразднён — создание/редактирование записи открывается в выезжающем слева drawer поверх списка, а не на отдельной странице.

---

## 1. Стек и ключевые архитектурные решения

| Слой | Выбор | Почему |
|---|---|---|
| Фреймворк | **Nuxt 3** (Vue 3, Composition API, `<script setup>`) | SSR/BFF из коробки — нужен для п.4 (безопасное хранение токена) |
| Язык | **JavaScript** (не TypeScript), с JSDoc-аннотациями на публичных composables/утилитах | По заданию стек — JS. JSDoc даёт автодополнение в IDE без введения TS-тулчейна |
| UI-библиотека | **`@nuxt/ui` (Nuxt UI v3, на Tailwind CSS)** | Официальный модуль команды Nuxt, полностью темизируется одним конфигом (см. п.7), даёт готовые accessible-компоненты (таблицы, drawer/slideover, модалки, формы), не требует отдельной темы поверх |
| Состояние | **Pinia** | Стандарт для Nuxt 3 |
| Валидация форм | **Zod** (схемы 1:1 с ограничениями `buf.validate` из proto — см. п.11) | Единый источник правды для валидации на клиенте, зеркалит серверные ограничения |
| RPC-клиент | **Node.js gRPC-клиент (`@grpc/grpc-js`)**, используется **только внутри `server/api/**` (Nitro/Node), никогда в браузерном коде** | RPC-вызовы к бэкенду выполняет исключительно серверная часть Nuxt (Nitro/Node.js) — см. п.2, п.1.1 |
| i18n (интерфейс) | **`@nuxtjs/i18n`** | Стандартный модуль Nuxt для локализации UI (лейблы, кнопки, тексты ошибок) |
| Даты/числа | **`Intl` (нативно)**, обёрнутое в общие composables (см. п.9.6) | Не тянуть отдельную date-lib ради форматирования |
| Иконки | **`@iconify-json/lucide`** через `@nuxt/ui` | Согласовано с UI-китом |

### 1.1 RPC-транспорт

Бэкенд поднимает чистый gRPC-сервер на адресе из `configs/grpc.yaml` — без Connect, gRPC-Web или прокси перед ним; это не требуется и не меняется.

RPC-клиент к бэкенду живёт **только в серверной части Nuxt** (`server/api/**`, Nitro/Node.js) — обычный Node gRPC-клиент `@grpc/grpc-js`, с сервис-стабами под Node/TS (например `protoc-gen-grpc-js`, или динамическая загрузка `.proto` через `@grpc/proto-loader` без кодогена). Конкретный выбор кодогена — на усмотрение реализации; `protoc-gen-es` (уже настроенный в `proto/buf.gen.js.yaml`) можно оставить для типизации сообщений в server routes/Zod-схемах, но для самого RPC-вызова он не обязателен.

Браузерный код **никогда** не обращается к gRPC напрямую и не импортирует gRPC-клиент — только к Nuxt server routes через same-origin `fetch` (см. п.2).

**Ограничение деплоя**: Nuxt/Nitro должен деплоиться на **Node.js server runtime**, не на edge/serverless-пресет (Cloudflare Workers, Vercel Edge и т.п.) — `@grpc/grpc-js` требует Node-модули (`net`/`http2`), недоступные в edge-рантаймах.

---

## 2. Архитектура: BFF (Backend-for-Frontend) через Nuxt Server Routes

**Обязательное решение, не альтернатива.** Токен, выдаваемый `UsersService.Login`, — самоподписанный (HMAC), с TTL **72 часа по умолчанию** (конфигурируется на бэкенде, `crm.auth.tokenTTL`), и на бэкенде нет метода `Logout` (см. `internal/services/user`, `internal/transports/grpc/interceptors/auth` — token stateless, отозвать конкретный токен раньше истечения нельзя). Это означает: если токен утечёт (XSS), он скомпрометирован на весь оставшийся срок действия — вплоть до 72 часов — пока не истечёт сам или админ вручную не сменит пароль пользователю (что не отзывает уже выданные токены немедленно). Поэтому хранить токен в `localStorage`/`sessionStorage`/обычной JS-читаемой куке — недопустимо.

**Схема:**

```
Браузер (Vue-компоненты, Pinia)
   │  fetch('/api/...')  — same-origin, credentials: 'include'
   ▼
Nuxt Server Routes (server/api/**.js, Nitro — обычный Node.js-процесс)
   │  читает httpOnly-cookie с сессией, достаёт токен из server-side store
   │  вызывает обычный Node gRPC-клиент (@grpc/grpc-js) к бэкенду
   │  с Authorization: Bearer <token> в metadata — см. п.1.1
   ▼
gRPC backend (crm.grpc.*.v1 services, auth+RBAC-интерцепторы — см. п.4.2)
```

- После `UsersService.Login` Nuxt server route создаёт **httpOnly, Secure, SameSite=Strict** cookie с непрозрачным session id; сам gRPC-токен хранится **только на сервере** (in-memory LRU или Redis-backed session store — переиспользовать `configs/redis.yaml`, если Redis уже поднят для бэкенда, тем более что фронтенд и так живёт в той же инфраструктуре).
- Клиентский JS никогда не видит `UserLoginResponse.token`.
- Каждый server route (`server/api/[entity]/*.js`) — тонкая обёртка: достать сессию → собрать gRPC-запрос → вернуть JSON браузеру. Дублирования избегаем универсальным хелпером (см. п.9.1 `defineEntityHandler`).
- CSRF: поскольку основная cookie — `SameSite=Strict` и все мутации идут через `POST`/`PATCH` с проверкой `Origin`/`Sec-Fetch-Site` заголовка на сервере (встроенный Nitro-миддлвар, писать один раз в `server/middleware/csrf.js`).

**[best practice]** Такая схема также решает CORS (браузер всегда ходит same-origin на сам Nuxt) и скрывает реальный адрес gRPC-бэкенда от клиента.

---

## 3. Подключение к серверу — конфигурация, не хардкод

Никаких захардкоженных URL в коде. Конфигурация — **runtime**, не build-time (один и тот же собранный образ должен уметь указывать на разные окружения без пересборки — так же, как бэкенд грузит `configs/*.yaml` при старте).

**`nuxt.config.js`:**

```js
export default defineNuxtConfig({
  runtimeConfig: {
    // Приватное — доступно ТОЛЬКО на server routes, никогда не попадает в клиентский бандл
    grpc: {
      baseUrl: '',        // env: NUXT_GRPC_BASE_URL — адрес gRPC-сервера (configs/grpc.yaml), для @grpc/grpc-js в server routes (см. п.1.1)
      timeoutMs: 15000,   // env: NUXT_GRPC_TIMEOUT_MS
    },
    images: {
      // Адрес HTTP-сервера бэкенда (configs/http.yaml), нужен только
      // серверному коду для проксирования GET/POST /images — см. п.4.6.
      // Если images проксируются через Nuxt (рекомендуется), в клиентский
      // бандл не попадает вообще.
      baseUrl: '',         // env: NUXT_IMAGES_BASE_URL
      maxSizeBytes: 10485760, // env: NUXT_IMAGES_MAX_SIZE_BYTES — зеркалит crm.images.maxSizeBytes на бэкенде, для клиентской preflight-проверки (п.12.6)
    },
    session: {
      secret: '',         // env: NUXT_SESSION_SECRET — подпись cookie
      ttlSeconds: 60 * 60 * 12, // env: NUXT_SESSION_TTL_SECONDS
    },
    public: {
      // Публичное — можно читать в браузере
      appName: 'MyBusiness CRM',
      defaultLocale: 'sr',
      supportedLocales: ['sr', 'en', 'ru'],
      imagesMaxSizeBytes: 10485760, // публичное зеркало runtimeConfig.images.maxSizeBytes, см. п.12.6
    },
  },
})
```

- Значения приходят из переменных окружения (`NUXT_GRPC_BASE_URL` и т.д. — стандартный механизм Nuxt `runtimeConfig` ↔ env) или из `.env`/оркестратора (Docker/K8s secret) — по аналогии с тем, как бэкенд использует `CRM_`-префиксные env-оверрайды поверх `configs/*.yaml`.
- **Не** использовать `process.env` напрямую в компонентах/composables — только через `useRuntimeConfig()`, и только `public.*` вне server-контекста.
- Один и тот же Docker-образ фронтенда конфигурируется под dev/stage/prod исключительно переменными окружения при деплое.

---

## 4. Аутентификация и безопасность

### 4.1 Поток логина

1. Форма логина (`login` = телефон или email, `password`) → `POST /api/auth/login`.
2. Server route вызывает `UsersService.Login`, на успех — создаёт сессию (см. п.2), ставит cookie, возвращает браузеру **только** публичные поля `User` (без токена): `id, name, lastName, phone, email, role, status`.
3. Ошибки от бэкенда мапятся на понятные сообщения (см. таблицу кодов в п.4.4).
4. **Logout**: бэкенд не имеет RPC `Logout` (сознательное решение бэкенда — токен stateless, отдельного отзыва нет; см. п.2). Frontend-«логаут» — это удаление серверной сессии/cookie (`POST /api/auth/logout`), сам токен на бэкенде при этом остаётся валидным ещё до 72 часов (или сколько осталось до истечения TTL). **Обязательно отразить в UI** (например, текст в настройках профиля: «выход из аккаунта на этом устройстве», а не «отзыв доступа») — не вводить пользователя в заблуждение.
5. Инвалидность сессии (cookie истекла/удалена) → редирект на `/login`, сохраняя `redirect`-query для возврата на исходную страницу после логина.

### 4.2 RBAC

Бэкенд хранит таблицу `роль → разрешения` в конфиге (`internal/rbac`, `configs/crm.yaml`) и **реально проверяет её на каждый gRPC-запрос** через выделенный RBAC-интерцептор (`internal/transports/grpc/interceptors/rbac`), который выполняется сразу после аутентификации. Это полноценная граница безопасности на бэкенде, не просто конфигурация для показа.

Особенности enforcement, которые важно знать фронтенду:
- **Deny by default**: метод, для которого на бэкенде не заведено разрешение, возвращает `PermissionDenied` для всех ролей, кроме той, у которой в конфиге стоит `"*"` (обычно `admin`). Не полагаться на «если не запрещено явно — значит можно».
- `Login` не требует аутентификации (иначе на нём не на что было бы залогиниться). `ChangePassword` доступен **любой** аутентифицированной роли без специального permission — это self-service-действие (сам запрос уже требует знания текущего пароля).
- Точные строки permission на каждый метод — см. п.10 (сверено с `internal/transports/grpc/interceptors/rbac/permissions.go` на бэкенде).

**Следствие для фронтенда:**
- UI обязан скрывать разделы/кнопки, недоступные роли (UX и защита от лишних кликов) — см. п.4.3.
- Но UI-скрытие — это **только UX**, не сама граница безопасности (та — на бэкенде). Каждая data-мутация всё равно должна корректно обрабатывать `PermissionDenied` от сервера (см. п.4.4) через единый error-handler (см. п.9.4) — например, если у пользователя отозвали право между открытием страницы и кликом по кнопке.
- Не проектировать логику вида «раз кнопка спрятана — значит действие невозможно»: клиент всегда может обойти скрытый UI (dev tools, прямой вызов API), поэтому реальная защита — только на бэкенде, а фронтенд-гейтинг существует исключительно для удобства.

### 4.3 Модель разрешений на фронтенде

Бэкенд не отдаёт список permissions пользователю и не будет (нет такого RPC, добавлять не планируется) — фронтенд ведёт **свою копию** таблицы роль → права в одном файле, независимо от `configs/crm.yaml` на бэкенде. Раз это два независимых источника правды, каждая мутация всё равно обязана проходить через единый error-handler (см. п.4.4/9.4), корректно обрабатывающий `PermissionDenied`, а не полагаться только на то, что кнопка была скрыта:

```js
// app/config/permissions.js
export const ROLE_PERMISSIONS = {
  admin: ['*'],
  employee: [
    'products:read', 'products:create',
    'categories:read', 'categories:create',
    'brands:read', 'brands:create',
    'sales:create',
  ],
  guest: [],
}

// Формат прав: `${entityKey}:${action}`, action ∈ read|create|update|delete
// (значения выше — пример из шаблона configs/crm.yaml на бэкенде, реальный
// список зависит от того, что реально настроено в проде; полный набор
// возможных ключей — см. п.10).
// ChangePassword не гейтится через permission вообще (см. п.4.2) — не
// добавлять для него запись сюда, usePermission() для него не вызывается.
```

**Единая точка входа для проверки** — composable `usePermission()`:

```js
// app/composables/usePermission.js
export function usePermission() {
  const { user } = useAuth() // Pinia store, см. п.9
  function can(permission) {
    if (!user.value) return false
    const perms = ROLE_PERMISSIONS[user.value.role] ?? []
    return perms.includes('*') || perms.includes(permission)
  }
  return { can }
}
```

Используется везде одинаково: `<UButton v-if="can('products:create')">`, гварды роутов, генерация меню (п.8) — **никогда не проверять `user.role === 'admin'` напрямую в компонентах**, только через `can(...)`. Так смена правил доступа — правка одного файла `permissions.js`, а не поиск по всему проекту.

### 4.4 Маппинг ошибок сервера → UI

Бэкенд возвращает стандартные gRPC-коды (см. `MapError` в каждом `internal/transports/grpc/handlers/*/handler.go`). Единый маппер (см. п.9.4) переводит их в:

| gRPC-код | Типичная причина | UI-реакция |
|---|---|---|
| `NotFound` | Сущность/связанная сущность не найдена | Toast + для detail-страницы — редирект на список |
| `AlreadyExists` | Конфликт уникальности (SKU/phone/email) | Инлайн-ошибка на поле формы |
| `Aborted` (`ErrStaleEntity`) | Кто-то изменил запись, пока вы её редактировали (etag не совпал) | Модалка «Запись изменена другим пользователем» → кнопка «Обновить и повторить» (перезагрузить актуальные данные, не потеряв ввод пользователя, если возможно — иначе просто reload формы) |
| `FailedPrecondition` | Бизнес-правило (недостаточно остатка, склад неактивен, sale в терминальном статусе и т.п.) | Toast с точным текстом причины (коды `errs.*` уже человекочитаемы на английском — см. п.9.4 про локализацию сообщений) |
| `InvalidArgument` | Валидация | Инлайн на поле, если можем сопоставить поле; иначе — toast |
| `Unauthenticated` | Сессия истекла/невалидна | Редирект на `/login` |
| `PermissionDenied` | RBAC (после появления enforcement) | Toast «Недостаточно прав» + скрыть действие в UI впредь (обновить кэш permissions) |
| `Unimplemented` | Метод-заглушка на бэкенде | Toast «Функция временно недоступна» (не должно происходить в проде, но не должно и падать белым экраном) |
| прочее / `Internal` | Непредвиденная ошибка | Общий toast + логирование на клиенте (см. п.13) |

### 4.5 Общие требования безопасности

- **CSP**: `default-src 'self'`; запретить `unsafe-inline` для скриптов (Nuxt поддерживает nonce из коробки через `@nuxt/security` или ручной nonce-миддлвар) — обязательно для митигации XSS, раз токен живёт вечно.
- **`@nuxt/security`** (модуль) — включить целиком: CSP, HSTS, X-Frame-Options (`DENY`, чтобы CRM нельзя было завернуть в iframe/clickjacking), `X-Content-Type-Options: nosniff`, `Referrer-Policy: strict-origin-when-cross-origin`.
- Все пользовательские данные, рендерящиеся как HTML (нигде в этом приложении такого не должно быть — только текстовые интерполяции Vue, которые экранируют по умолчанию) — **никогда не использовать `v-html`** с данными из API.
- Пароли: те же ограничения, что на бэкенде (`min 8, max 128` символов) — валидировать на клиенте до отправки (UX), финальная валидация всё равно на сервере.
- Rate-limiting формы логина на уровне Nuxt server route (простая in-memory защита от брутфорса — N попыток / IP / минуту) — бэкенд не лимитирует `Login` отдельно.
- Загрузка изображений товара (см. п.4.6) — валидировать MIME/размер на клиенте до аплоада (UX), но не доверять этой проверке как единственной защите.

### 4.6 Загрузка изображений товара

Отдельный **HTTP-эндпоинт** (не gRPC) на том же бэкенд-хосте, что и gRPC-сервер, но на HTTP-порту (`configs/http.yaml`) — `internal/transports/http/handlers/image`:

- **`POST /images`** — админ-только (проверяется тот же bearer-токен, что и в gRPC, через `Authorization: Bearer <token>`; роль вызывающего должна быть `admin` — единственный HTTP-эндпоинт с прямой role-проверкой, не через RBAC-интерцептор, т.к. он специфичен для gRPC). Multipart form, поле файла — `file`. По умолчанию максимум 10 MiB, допустимые MIME (по содержимому, не по расширению) — `image/jpeg`, `image/png`, `image/gif`, `image/webp`. Ответ: `{ "id": "<uuid>" }`.
- **`GET /images/{id}`** — без аутентификации (изображение товара — публичный ресурс, как и остальной каталог). Отдаёт байты файла с корректным `Content-Type`/`Content-Length` и долгим `Cache-Control` (immutable — id один на файл, перезаписи по тому же id не бывает).
- Полученный `id` из ответа `POST /images` прикрепляется к `Product.imageIds` обычным gRPC `ProductsService.Update` (FieldMask `imageIds`) — сам upload не создаёт и не трогает `Product` напрямую.

**Как встраивать в BFF-схему:**
- **Upload** — обязательно через Nuxt server route (`POST /api/products/[id]/images` или похожий), как и все остальные мутации: браузер не должен знать реальный bearer-токен, значит не может сам вызвать `POST /images` на бэкенде напрямую. Server route читает файл из multipart-запроса браузера, добавляет токен из серверной сессии (см. п.2) и проксирует на бэкенд.
- **Serve** (`GET /images/{id}`) — не требует токена, поэтому **можно** отдавать браузеру прямой URL на HTTP-хост бэкенда для `<img src>` (быстрее — без лишнего прыжка через Nitro на каждую картинку в списке товаров), но тогда: (а) нужен публично резолвящийся адрес бэкенд-HTTP-хоста и CORS/сетевой доступ до него из браузера пользователя, что противоречит принципу «скрывать реальный адрес бэкенда» из п.2; (б) при желании отдать этот адрес — прокинуть его через `runtimeConfig.public` (см. п.3), не хардкодить. **Рекомендация [best practice]**: проксировать и `GET`-запросы тоже через `server/api/images/[id].get.js` (простой passthrough, без бизнес-логики) — тогда всё остаётся same-origin, бэкенд-адрес нигде не светится, а разница в перформансе на практике незаметна (статика с `Cache-Control: immutable` эффективно кэшируется браузером после первого запроса).

---

## 5. Мультиязычность

Два независимых слоя локализации — не путать:

1. **Язык интерфейса** (лейблы кнопок, заголовки, сообщения об ошибках) — через `@nuxtjs/i18n`, файлы `i18n/locales/{sr,en,ru}.json`. Переключатель языка — в шапке (см. п.8.1).
2. **Локализованный контент сущностей** (`LocalizedString` — `Brand.name`, `Product.description`, и т.д.) — это данные, не UI-строки, хранятся как `{ values: { sr: '...', en: '...' } }`.

### 5.1 Правила (зеркалят бэкенд)

- Обязательный ключ — **`sr`**. Остальные (`en`, `ru`, ...) — опциональны.
- При создании/редактировании: поле `sr` — обязательное (звёздочка, валидация Zod), остальные локали — опциональны, можно добавлять/убирать динамически (не жёстко зашитый список — админ теоретически может ввести любой код локали, хотя UI по умолчанию предлагает только `sr/en/ru`, синхронно с `public.supportedLocales`).
- Отображение в списках/на детальной странице: показываем текст на **текущем языке интерфейса**, если он есть в `values`; иначе — **всегда `sr`** (гарантированный фолбэк, как на бэкенде).

### 5.2 Универсальный компонент `<LocalizedStringInput>`

Единственное место, где живёт логика редактирования `LocalizedString` — переиспользуется в формах Brand/Category/Warehouse/Product/User/Product.details:

```vue
<!-- app/components/form/LocalizedStringInput.vue -->
<LocalizedStringInput
  v-model="form.name"         
  :locales="supportedLocales"  
  required-locale="sr"
  label="Название"
  multiline
/>
```

- Рендерит табы/аккордеон по локалям (сколько задано в `supportedLocales` + уже существующие в данных, если там есть локаль вне списка — не терять её молча).
- `v-model` работает с форматом `{ [locale]: string }`, а не с proto-обёрткой напрямую — маппинг в `{ values: {...} }` делает API-слой (п.9.2), не компонент.

### 5.3 Универсальный компонент `<LocalizedText>` (только чтение)

```vue
<LocalizedText :value="brand.name" />
```
Инкапсулирует правило фолбэка (текущая локаль → `sr` → пусто). Используется во всех таблицах/детальных страницах вместо `{{ brand.name.values[locale] }}` в лоб — одно место, где живёт fallback-логика.

### 5.4 Формат чисел/дат

- Все даты в proto — `int64` unix-секунды. Единый composable `useFormatDate(unixSeconds, style)`, использующий `Intl.DateTimeFormat(currentLocale, ...)`.
- Деньги — `priceAmount`/`totalAmount`/`discountAmount` в **basis points** (реальная сумма ×100, см. proto-комментарии). Единый composable `useFormatMoney(basisPoints, currencyCode)` → `Intl.NumberFormat(locale, { style: 'currency', currency })`. `currencyCode` берём из самого объекта `ProductPrice.currency` (сервер уже проставляет его при создании цены) — не хардкодить и не запрашивать отдельно.
- Проценты (`discountPercentage`, `commissionPercentage`) — обычные целые 0-100, форматировать как `${n}%`.

---

## 6. Адаптивность (поддержка основных разрешений экрана)

Брейкпоинты — стандартные Tailwind (те же, что использует `@nuxt/ui`), не изобретать свои:

| Название | Ширина | Поведение |
|---|---|---|
| `sm` | ≥ 640px | Базовый мобильный |
| `md` | ≥ 768px | Планшет |
| `lg` | ≥ 1024px | Десктоп — оба drawer'а видны как постоянные панели (не overlay) |
| `xl` / `2xl` | ≥ 1280 / 1536px | Широкий десктоп — таблицы получают больше колонок по умолчанию |

**Обязательные паттерны:**
- **< `lg`**: правый drawer (разделы) и левый drawer (карточки сущностей, п.8.2) становятся **overlay** (открываются по кнопке-гамбургеру / свайпу), не занимают место в layout.
- **< `md`**: таблицы списков (`<EntityDataTable>`, п.9.3) переключаются в режим **карточек** (каждая строка → карточка с ключевыми полями), а не горизонтальный скролл таблицы — иначе на телефоне нечитаемо.
- Формы — всегда одна колонка на `< md`, до двух колонок на `≥ md` (компонент `<FormGrid>`, см. п.9.3).
- Тестировать на: 360×640 (мобильный), 768×1024 (планшет), 1366×768 и 1920×1080 (десктоп) — минимальный набор для ручной проверки.

---

## 7. Дизайн-система (design tokens — одна точка изменения)

Требование заказчика (уточнено после ревизии стилистики, см. п.0): **насыщенный бирюзовый акцент (`#85C5CF`)** на TopBar/кнопках/активном пункте меню, **тёмно-серое (`#333333`) меню** (не чёрное), **поля форм — белый фон, чёрная скруглённая рамка, чёрный текст**, и «всё переиспользуемо, менять не в тысяче мест». Реализация — **единый файл токенов**, из которого берёт значения и Tailwind-конфиг, и сам `@nuxt/ui` theme:

```js
// app/design/tokens.js
export const tokens = {
  color: {
    brand: {
      // Бирюзовая шапка/кнопки/активный пункт меню. Один основной оттенок + шкала.
      50:  '#f4f9fa',
      500: '#85c5cf',   // ← основной бирюзовый, TopBar/UButton primary/активный пункт меню
      600: '#6ba9b2',
      700: '#558790',
    },
    text: {
      primary: '#000000',    // «чёрный текст» — чистый чёрный, как и заголовки
      secondary: '#525252',
      inverse: '#ffffff',    // текст на бирюзовой шапке / тёмном меню
    },
    surface: {
      base: '#ffffff',       // «белый фон» полей ввода
      subtle: '#f9fafb',     // фон карточек/строк-зебра
      border: '#000000',     // чёрная рамка полей ввода (слегка скруглённая — --ui-radius)
      menu: '#333333',       // фон TopBar/меню (RightDrawer) — тёмно-серый, не чёрный
    },
    status: {
      // Единая палитра статусов — переиспользуется StatusBadge (п.9.3) для ВСЕХ enum'ов
      active:    { bg: '#dcfce7', text: '#166534' }, // Active/Completed/Paid
      inactive:  { bg: '#f3f4f6', text: '#374151' }, // Inactive/Draft/Unspecified
      warning:   { bg: '#fef9c3', text: '#854d0e' }, // Shipped/pending states
      danger:    { bg: '#fee2e2', text: '#991b1b' }, // Cancelled
      info:      { bg: '#dbeafe', text: '#1e40af' }, // Refunded/neutral-info
    },
  },
  radius: { sm: '4px', md: '8px', lg: '12px' },
  spacing: { unit: '4px' }, // всё остальное — кратно 4px (Tailwind по умолчанию так и делает)
}
```

- `assets/css/main.css` (`@theme static` — Tailwind v4 читает цвета из CSS, не JS) и `app.config.js` (`@nuxt/ui` theme, `ui.colors.primary: 'brand'`) **зеркалят значения из `tokens.js`**. Смена акцентного цвета — правка `tokens.js` **и** блока `--color-brand-*` в `main.css` (два места, а не один, так как Tailwind v4 не читает JS-конфиг напрямую — см. комментарий в файле).
- Белый фон/чёрная рамка/чёрный текст полей ввода реализованы не per-компонентно, а через переопределение семантических CSS-токенов `@nuxt/ui` (`--ui-bg`, `--ui-border-accented`, `--ui-text-highlighted`, `--ui-radius`) в `main.css` — подхватывается всеми `UInput`/`USelect`/`UTextarea` и т.п. автоматически.
- Экран логина — отдельный кейс: фон разделён 50/50 (`#333333` / `tokens.color.brand[500]`), белая карточка формы по центру поверх обеих половин.
- **`<StatusBadge :status="brand.status" :map="STATUS_COLOR_MAP.brand" />`** — единый компонент бейджа статуса для всех enum'ов проекта (`BrandStatus`, `CategoryStatus`, `ProductStatus`, `SaleStatus`, `UserStatus`, `WarehouseStatus`, `PartnerStatus`, `MovementType`), каждый маппится на 5 цветовых семейств из `tokens.color.status`, а не на свой набор цветов.
- Тёмная тема — **вне scope** (заказчик указал конкретную схему: бирюзовый/серый/белый/чёрный), но токены проектируются так, чтобы тёмную тему можно было добавить, не трогая компоненты (только `tokens.js` + CSS-переменные).

---

## 8. Layout приложения

```
┌─────────────────────────────────────────────────────────────────┐
│  TopBar (бирюзовый фон, белый текст): лого | 🌐 язык | 👤 профиль/логаут │
├───────────────┬─────────────────────────────────────────────────┤
│               │                                                   │
│  SideMenu     │                Основной контент                  │
│  (тёмно-серый,│      (Dashboard ИЛИ список сущности —             │
│  #333333):    │       карточки/таблица + кнопка «Создать»)         │
│  Dashboard    │                                                   │
│  Бренды       │                                                   │
│  Категории    │  Клик по строке / иконке ✏️ / кнопке «Создать» —   │
│  ...          │  открывает EditDrawer справа поверх контента       │
│  (плоский     │  (см. 8.3), список остаётся видимым сбоку.         │
│  список, без  │                                                   │
│  разделов)    │                                                   │
└───────────────┴─────────────────────────────────────────────────┘
```

### 8.1 TopBar

- Бирюзовый фон (`tokens.color.brand[500]`, `#85C5CF`), белый/светлый текст (`tokens.color.text.inverse`).
- Слева: лого/название приложения (клик → `/` — Dashboard).
- Справа: переключатель языка интерфейса (`sr/en/ru` — флаг/код, не иконка страны конкретной, чтобы не привязываться политически), затем меню профиля (имя пользователя, роль, «Сменить пароль» → модалка на `UsersService.ChangePassword`, «Выйти»).
- На `< lg` — гамбургер-иконка слева от лого, открывающая левый SideMenu как overlay (см. п.6).

### 8.2 SideMenu — плоское меню слева (постоянный на `≥ lg`, overlay на `< lg`)

Единственный уровень верхнеуровневой навигации — **без** промежуточного клика по разделу. Тёмно-серый фон (`tokens.color.surface.menu`, `#333333`), активный пункт подсвечен бирюзовым (`tokens.color.brand[500]`). Список — `Dashboard` + одна строка на каждую сущность из `EntityRegistry` (п.9.1), генерируется автоматически:

```js
// app/config/navigation.js
export const NAV_ITEMS = [
  { key: 'dashboard', label: 'nav.dashboard', icon: 'i-lucide-layout-dashboard', to: '/' },
  // остальное дописывается в SideMenu.vue из listEntityConfigs() —
  // не дублируется здесь; добавление сущности в EntityRegistry
  // автоматически добавляет её в меню, отдельно ничего не редактируется.
]
```

- Пункт, для которого `can('<entity>:read')` — `false`, **не рендерится вообще** (не disabled, а отсутствует).
- Клик по пункту — обычная навигация на `entity.route` (список сущности), никакого раскрытия/промежуточного состояния.
- По мере роста числа сущностей (Products, Categories, Sales, Warehouse и т.д. — см. п.10) список станет длиннее; группировка на разделы **осознанно не восстанавливается** — таково явное решение по итогам стилистической ревизии заказчика (см. п.0). Если список станет неудобно длинным, разбивку стоит обсуждать отдельно, а не тихо возвращать прежнюю двухуровневую схему.

### 8.3 EditDrawer — создание/редактирование в выезжающем справа drawer

Клик по строке списка, по иконке ✏️ в колонке «Действия» (п.9.3) или по кнопке «Создать» открывает `USlideover` со стороны `right` поверх текущего контента (список остаётся виден и не теряет состояние/скролл/фильтры) — **не** переход на отдельную страницу `/[entity]/new` или `/[entity]/[id]`. Внутри — тот же generic `<EntityForm>` (п.9.3), в режиме `mode="drawer"`: вместо `router.push` после сохранения/удаления форма эмитит `saved`/`cancel`/`deleted`, хозяин (`EntityListPage`) закрывает drawer и обновляет таблицу на месте.

- Владелец состояния drawer (открыт/закрыт, id редактируемой записи) — сама `EntityListPage`, не глобальный layout-стор: drawer нужен только там, где есть список. `<EntityForm>` внутри рендерится с `:key="editingId ?? 'create'"`, чтобы при смене редактируемой записи компонент пересоздавался (а не переиспользовался с устаревшим состоянием) и заново подгружал данные.
- Страницы `/[entity]/new.vue` и `/[entity]/[id].vue` (см. п.14) продолжают существовать как прямые ссылки (deep link, «открыть эту запись по URL») и рендерят `<EntityForm mode="page">` без изменений — но обычный путь пользователя через UI всегда идёт через drawer, кнопки на них не ведут.

### 8.4 Главная страница (`/`) — аналитика

Дашборд строится из `ReportsService`:
- Виджет «Оборот» (`GetTurnover`) — линейный график по дням за выбранный период (селектор периода: сегодня / 7 дней / 30 дней / произвольный диапазон, компонент `<PeriodFilter>`, переиспользуется и в фильтрах списков, где есть `createdAt`-фильтр).
- Виджет «Продажи за период» (`GetSalesReport`) — количество продаж + сумма, тот же период.
- Виджет «Топ товаров» (`GetPopularProducts`, `limit=10`) — таблица/бар-чарт.
- Виджет «Продажи по сотрудникам» (`GetSalesByStaff`) — таблица (сейчас `userId` без резолва в имя, т.к. фронтенд может дополнительно подтянуть `UsersService.Get` по каждому `userId` и закэшировать — см. п.9.5 про батч-резолв справочников).
- Виджет «Продажи по партнёрам» (`GetSalesByPartner`) — таблица с `commissionAmount`.
- Виджет «Остатки» (`GetStockLevels`) — компактная таблица/сводка (можно топ-N по минимальному остатку — «требует внимания»), с ссылкой на полный список Inventory.
- Каждый виджет — **отдельный компонент, независимо грузящий свои данные** (свой loading/error state), чтобы падение одного отчёта не блокировало остальной дашборд (`<DashboardWidget>` — обёртка с единым skeleton/error UI, п.9.3).
- Виджеты, зависящие от `sales:read`/`reports:read`, скрываются по `can(...)`, если пользователь не должен видеть финансовую аналитику (например, `guest`).

---

## 9. Универсальные строительные блоки

Это — центральное требование задачи («не менять значения в тысяче мест»). Ничего специфичного к конкретной сущности не должно жить в more than one файле.

### 9.1 `EntityRegistry` — единый реестр сущностей

Один файл-источник правды на каждую сущность: как её показывать в списке, как редактировать, какие права нужны.

```js
// app/config/entities/products.js
export default {
  key: 'products',
  label: 'entities.products.label',        // ключ i18n
  icon: 'i-lucide-box',
  route: '/products',
  permissions: {
    read: 'products:read', create: 'products:create',
    update: 'products:update', delete: 'products:delete',
  },
  api: 'products', // см. п.9.2 — имя Connect-сервиса/обёртки

  list: {
    columns: [
      { key: 'sku', label: 'fields.sku', sortable: false },
      { key: 'name', label: 'fields.name', component: 'LocalizedText' },
      { key: 'brandId', label: 'fields.brand', component: 'RelationLabel', relation: 'brands' },
      { key: 'status', label: 'fields.status', component: 'StatusBadge', statusMap: 'product' },
      { key: 'createdAt', label: 'fields.createdAt', component: 'DateLabel' },
    ],
    filters: [
      { key: 'statuses', type: 'multiselect', label: 'fields.status', optionsFrom: 'enum:ProductStatus' },
      { key: 'brandIds', type: 'multiselect', label: 'fields.brand', optionsFrom: 'relation:brands' },
      { key: 'categoryIds', type: 'multiselect', label: 'fields.categories', optionsFrom: 'relation:categories' },
      { key: 'skus', type: 'tags', label: 'fields.sku' },
      { key: 'createdAt', type: 'periodFilter', label: 'fields.createdAt' },
    ],
    sort: [
      { field: 'FIELD_CREATED_AT', label: 'sort.createdAt' },
      { field: 'FIELD_NAME', label: 'sort.name' },
    ],
    defaultSort: { field: 'FIELD_CREATED_AT', direction: 'DESC' },
  },

  form: {
    fields: [
      { key: 'sku', type: 'text', label: 'fields.sku', required: true, maxLength: 64, immutableOnEdit: true },
      { key: 'name', type: 'localizedString', label: 'fields.name', required: true },
      { key: 'description', type: 'localizedString', label: 'fields.description' },
      { key: 'brandId', type: 'relation', relation: 'brands', label: 'fields.brand', required: true },
      { key: 'categoryIds', type: 'relationMulti', relation: 'categories', label: 'fields.categories' },
      { key: 'details', type: 'keyValueLocalized', label: 'fields.details' },
      { key: 'status', type: 'enum', enum: 'ProductStatus', label: 'fields.status', editOnly: true },
    ],
  },
}
```

- Регистрация всех сущностей — `app/config/entities/index.js`, экспортирующий массив/мапу. `LeftDrawer` (п.8.3), роутинг, генератор списков/форм (`EntityListPage`, `EntityFormPage` — п.9.3) — все читают **только** этот реестр.
- Добавить новую сущность или новое поле в список/форму = добавить/поправить один объект конфига. Новый Vue-файл не нужен, если сущность укладывается в стандартный CRUD-паттерн (см. п.9.3). Не укладывается (Sale, User, Warehouse, Price, InventoryMovement — см. п.12) — переопределяются точечные части (кастомные действия, кастомные шаги формы), но список/базовые поля всё равно берутся из реестра.

### 9.2 API-слой — одна фабрика на все сущности

```js
// app/composables/useEntityApi.js
export function useEntityApi(entityKey) {
  const config = getEntityConfig(entityKey)
  return {
    async list(params) { /* POST /api/${entityKey}/list, курсорная пагинация */ },
    async get(id) { /* POST /api/${entityKey}/get */ },
    async create(payload) { /* POST /api/${entityKey}/create */ },
    async update(id, payload, etag) { /* PATCH /api/${entityKey}/update — шлёт FieldMask только для изменённых полей + etag */ },
    async remove(id, etag) { /* DELETE /api/${entityKey}/delete */ },
  }
}
```

- **FieldMask на update** формируется автоматически: сравнивается исходный объект (полученный при открытии формы) с текущим состоянием формы, в маску попадают только реально изменённые пути — не «отправить все поля всегда». Общая утилита `buildUpdateMask(original, current)`, используется всеми формами.
- **etag** хранится в скрытом состоянии формы с момента загрузки записи и всегда передаётся в `update`/`delete`/спецдействия (`Cancel`, `UpdateStatus`, `Deactivate`) — компонент формы не должен «забывать» его прокинуть; это гарантируется тем, что `useEntityForm` (см. ниже) сам добавляет etag в payload, вызывающий код его не трогает.
- Server route на бэкенде (`server/api/[entity]/*.js`) — **тоже генерируется по одному шаблону** через `defineEntityHandler(entityKey, connectClient)` (Nitro-хелпер), а не пишется вручную на каждую сущность.
- Сущности с нестандартными методами (`Login`, `ChangePassword`, `Deactivate`, `UpdateStatus`, `Cancel`, `GetHistory`, `GetByLogin`-подобные) добавляют дополнительные функции в свой `useEntityApi`-модуль поверх базового набора — не ломая стандартный контракт для generic-компонентов.

### 9.3 Общие UI-компоненты (реестр)

| Компонент | Назначение | Используется |
|---|---|---|
| `<EntityListPage :entity="key" />` | Полностью generic страница списка: таблица + фильтры + сортировка + пагинация + кнопка «Создать» (если `can(create)`) + действия в строке (edit/delete, если есть права) | Все сущности со стандартным списком (Brand, Category, Warehouse, Partner, Client, Product, Inventory, InventoryMovement, User, Sale — с точечными доп. колонками/действиями из реестра) |
| `<EntityDataTable>` | Таблица с server-side пагинацией (курсор), на `< md` — авто-переключение в карточки (п.6) | Внутри `EntityListPage`, плюс отдельно на детальных страницах (например, история цены/движений товара) |
| `<EntityForm :entity="key" :id="id?" mode="drawer"\|"page" />` | Generic форма создания/редактирования по `entity.form.fields`, с валидацией Zod, сборкой FieldMask, обработкой etag/409. `mode="page"` (default) — прежнее поведение (`router.push`) для прямых страниц `/[entity]/[id]`; `mode="drawer"` — используется `EntityListPage` внутри `EditDrawer` (п.8.3), эмитит `saved`/`cancel`/`deleted` вместо навигации | Все сущности со стандартной формой |
| `<PeriodFilter>` | Пресеты периода + произвольный диапазон → `PeriodFilter{from,to}` (unix-секунды) | Все списки с `createdAt`-фильтром, все отчёты |
| `<LocalizedStringInput>` / `<LocalizedText>` | См. п.5.2/5.3 | Везде, где `LocalizedString` |
| `<RelationSelect>` / `<RelationLabel>` | Выбор/отображение связанной сущности по id (с поиском, debounced-запрос к `list`, кэш в Pinia — см. п.9.5) | brandId, categoryIds, productId, clientId, warehouseId, partnerId и т.д. — везде одна реализация |
| `<StatusBadge>` | См. п.7 | Все enum-статусы |
| `<MoneyLabel>` / `<PercentLabel>` | См. п.5.4 | Price, Sale, Partner |
| `<ConfirmDialog>` | Единая модалка подтверждения (delete, cancel sale, deactivate warehouse) с обязательным текстом последствия — не голое «Вы уверены?» | Все деструктивные/необратимые действия |
| `<EtagConflictDialog>` | См. п.4.4 (`Aborted`) | Все формы редактирования |
| `<DashboardWidget>` | Обёртка виджета: заголовок, skeleton, error-state, retry | Дашборд (п.8.4) |
| `<FormGrid>` | Адаптивная сетка формы (1 колонка / 2 колонки, п.6) | Все формы |
| `<PermissionGate permission="...">` | Обёртка контента, скрывающая/показывающая слот по `can()` — альтернатива `v-if="can(...)"` для больших блоков | Разделы, кнопки действий, целые виджеты |

### 9.4 Единая обработка ошибок

`useEntityApi`/`useEntityForm` не ловят ошибки локально бесконтрольно — все server-route ответы об ошибке имеют единый JSON-формат `{ error: { code, message, field? } }` (маппинг из gRPC-статуса делает сервер route по таблице из п.4.4), и единый композабл `useApiErrorHandler()`:
- показывает toast или инлайн-ошибку на поле (`field`, если есть — сматченный с именем в `entity.form.fields`),
- для `Aborted` — открывает `<EtagConflictDialog>`,
- для `Unauthenticated` — редиректит на `/login`.

Ни один компонент не пишет свой `try/catch` с кастомным текстом — всегда через этот хендлер, иначе тексты ошибок расползутся по проекту.

### 9.5 Кэш справочников (resolve relations)

`brandId`, `categoryIds`, `productId`, `clientId`, `warehouseId`, `partnerId`, `userId` (в отчётах/логах) — везде нужно резолвить id → человекочитаемое имя. Единый Pinia store `useReferenceCache()`:
- `resolve(entityKey, id)` — если есть в кэше, отдаёт сразу; если нет — батчит запросы (собирает id за 1 тик события) и грузит через `list` с фильтром по id (там, где такой фильтр есть — Product/Brand/Category/Warehouse/Partner/Client это поддерживают через `filter.*Ids`/`skus`; для `User` в отчётах — по одному `Get`, т.к. `UsersListRequest.Filter` фильтрует по ролям/статусам, не по списку id, so батчить нечем — просто параллельные `Get` с общим кэшем).
- TTL кэша — сессия (не персистить между заходами, справочники меняются нечасто, но должны быть свежими в рамках работы).

### 9.6 Формат дат/чисел (composables)

`useFormatDate(unixSeconds, style = 'short' | 'long' | 'relative')`, `useFormatMoney(basisPoints, currencyCode)`, `useFormatPercent(value)` — три общих функции, никаких инлайновых `new Date(...)`/`toFixed(2)` по компонентам.

---

## 10. Каталог сущностей и прав

Легенда действий: **C**reate / **R**ead(List+Get) / **U**pdate / **D**elete / доп. действия.

| Сущность | gRPC-сервис | Действия | Permission-ключи | Особенности |
|---|---|---|---|---|
| Brand | `BrandsService` | C R U D | `brands:read/create/update/delete` | Delete запрещён, если есть товары (`FailedPrecondition`) |
| Category | `CategoriesService` | C R U D | `categories:read/create/update/delete` | `parentId` — плоское дерево (select с отступами по глубине в `<RelationSelect>`); Delete запрещён при наличии товаров |
| Warehouse | `WarehousesService` | C R U D + **Deactivate** | `warehouses:read/create/update/delete` (Deactivate тоже гейтится `warehouses:update` — отдельного permission-ключа для него на бэкенде нет, RPC разные, право одно) | Deactivate — отдельная кнопка (не через статус в форме!) с `<ConfirmDialog>`; блокируется `FailedPrecondition`, если есть остаток — текст ошибки должен явно сказать «перенесите остатки на другой склад» |
| Partner | `PartnersService` | C R U D | `partners:read/create/update/delete` | `commissionPercentage` — 0-100 (`<PercentLabel>`/number input с суффиксом %) |
| Client | `ClientsService` | C R U D | `clients:read/create/update/delete` | Простейшая сущность-эталон для generic CRUD |
| User | `UsersService` | C R U D + **Login** (публично) + **ChangePassword** | `users:read/create/update/delete` | Раздел виден только `admin` (и себе — своя карточка профиля вне общего списка); список **не** показывает пароль/хэш (бэкенд их и не отдаёт); `ChangePassword` — модалка из меню профиля, не через `EntityForm` |
| Product | `ProductsService` | C R U D | `products:read/create/update/delete` | `sku` неизменяем после создания (readonly при edit); `details` — key→LocalizedString редактор (см. компонент ниже); `imageIds` — загрузка через отдельный HTTP-эндпоинт, см. п.4.6; на detail-странице — вкладка «Цена» (см. Price) и вкладка «Остатки по складам» (Inventory, отфильтровано по `productId`) |
| ProductPrice | `PricesService` | C R U D + **GetHistory** | `prices:read/create/update/delete` | Не отдельный список в меню — живёт как вкладка на detail-странице Product (`GetByProductID`); история цен — таблица (readonly) под текущей ценой; суммы — basis points → `<MoneyLabel>` |
| Inventory | `InventoryService` | R only | `inventory:read` | Список с фильтром по складу/товару/диапазону количества; ссылки на Product/Warehouse; отдельный пункт меню **и** вкладка на Product |
| InventoryMovement | `InventoryMovementsService` | C (только Create — ручная операция «Приход»/«Списание»/«Корректировка») + R + **GetHistory** | `inventorymovements:read/create` (**без подчёркивания** — ключ ресурса на бэкенде `inventorymovements`, не `inventory_movements`, несмотря на то что gRPC-пакет называется `inventory_movement`) | Форма Create — только для типов `Receipt/WriteOff/Adjustment/Transfer` (тип `Sale` создаётся только бэкендом внутри `SalesService.Create`, руками через форму создавать «Sale»-движение не давать — исключить его из select типа в форме); список — неизменяемый лог, без edit/delete в UI вообще (кнопок нет физически, не просто disabled) |
| Sale | `SalesService` | C R + **UpdateStatus** + **Cancel** (нет generic Update/Delete) | `sales:read/create/update` (**оба** `UpdateStatus` и `Cancel` гейтятся одним ключом `sales:update` — отдельных `updateStatus`/`cancel` на бэкенде нет, оба RPC — просто переходы статуса существующей продажи) | См. п.12.3 — отдельный, не generic, флоу |
| Report* | `ReportsService` | R only (6 методов) | `reports:read` | См. п.8.4 — только виджеты дашборда + отдельная страница `/reports` с теми же виджетами в полноэкранном/детальном виде и экспортом в CSV (**[best practice]**: экспорт — частый запрос для отчётов, простой client-side CSV из уже загруженных строк, без нового бэкенд-эндпоинта) |

---

## 11. Валидация форм (зеркало `buf.validate` из proto)

Единая Zod-схема на каждый `entity.form` — генерируется из декларации полей в `EntityRegistry` (п.9.1) через общую функцию `buildZodSchema(fields)`, а не пишется вручную по сущности. Constraints 1:1 с proto:

| Тип поля | Zod-эквивалент |
|---|---|
| `text` + `required` | `z.string().min(1)` |
| `text` + `maxLength: N` | `.max(N)` |
| email (Client/Partner/User) | `.email()` |
| phone | `.max(32)` — формат не валидируется строже, чем бэкенд (бэкенд не навязывает формат, только длину) |
| `localizedString` + `required` | кастомный refine: `values.sr` непустой |
| `enum` | `z.nativeEnum(...)`, значение `*_UNSPECIFIED` (`= 0`) исключено из опций select везде (не показывать пользователю как выбираемый вариант) |
| `relation` | `z.string().uuid()` |
| числовые проценты | `z.number().int().min(0).max(100)` |
| пароль (User create) | `z.string().min(8).max(128)` |
| SaleItem.quantity | `z.number().int().positive()` |
| InventoryMovement.quantity | `z.number().int().refine(v => v !== 0)` |

**Важно:** это дублирование правил (клиент + сервер) — намеренно, для UX (мгновенная обратная связь). Финальный источник истины всегда сервер; клиентская валидация никогда не «доверяется» настолько, чтобы пропускать данные без ответа сервера.

---

## 12. Нестандартные флоу (не покрываются generic `EntityForm`)

### 12.1 Product — вкладки на detail-странице

`/products/[id]` — не просто форма, а страница с вкладками:
1. **Общее** — стандартная `EntityForm` (название, описание, бренд, категории, характеристики, статус).
2. **Изображения** — `ProductImageUploader` (см. п.4.6/12.6, изолированный компонент).
3. **Цена** — текущая `ProductPrice` (создать, если нет; редактировать `priceAmount`/`discountAmount`; кнопка «История» → таблица `GetHistory`).
4. **Остатки** — `<EntityDataTable entity="inventory" :fixed-filter="{ productId: id }" />` (readonly).
5. **Движения склада** — `<EntityDataTable entity="inventoryMovements" :fixed-filter="{ productIds: [id] }" />` + кнопка «Новое движение» (открывает форму Create InventoryMovement с предзаполненным `productId`).

### 12.2 Warehouse — Deactivate

Отдельная кнопка на detail-странице (не через смену `status` в форме — в proto `Deactivate` вообще отдельный RPC, а `status` в `WarehouseUpdateRequest` даже не редактируется напрямую пользователем, т.к. `WarehouseUpdateRequest` не содержит поля `status`!). Список складов показывает статус read-only (в бейдже), меняется только через `Deactivate`.

### 12.3 Sale — самый сложный флоу

**Создание** (`/sales/new`, мастер в 2 шага, не одна форма):
1. Шаг 1: выбрать `clientId` (с опцией «создать нового клиента прямо здесь» — модалка `EntityForm entity="clients"` внутри мастера, чтобы не терять контекст), `warehouseId` (только активные склады в select), опционально `partnerId`.
2. Шаг 2: добавление позиций — `<RelationSelect relation="products">` на каждой строке + `quantity` + `discountPercentage`; **цена не вводится** (proto не принимает `priceAmount` в Create — сервер сам берёт текущую цену), но UI обязан **показать текущую цену и предварительный итог** (клиентский расчёт: `priceAmount × quantity × (100 − discount) / 100`, посчитанный по актуальной `ProductPrice`, подтянутой при выборе товара — это предварительный расчёт для UX, финальный `totalAmount` всегда берётся из ответа сервера, не из клиентского расчёта). Показывать остаток на выбранном складе рядом с каждой позицией (запрос к `InventoryService.Get`), с предупреждением (не блокировкой ввода — сервер всё равно даст авторитетный `FailedPrecondition`, если не хватит), если запрошенное количество превышает остаток.
3. После создания — редирект на `/sales/[id]`.

**Detail-страница** (`/sales/[id]`):
- Позиции — **всегда readonly** (proto: «Items are immutable once created»).
- Статус — бейдж + кнопка «Изменить статус» → модалка со списком **разрешённых переходов** (клиентская state-machine для UX, зеркалящая то, что реально проверяет сервис: из `Cancelled`/`Refunded` переходов нет вообще — кнопка скрыта, если статус терминальный; во всех остальных случаях предлагать все статусы, кроме текущего, финальное решение — за сервером).
- Кнопка «Отменить продажу» (`Cancel`) — отдельно от `UpdateStatus`, с обязательным полем «Причина» (`reason`, опционален в proto, но в UI сделать **практически обязательным** полем — это единственный способ впоследствии понять, почему продажа отменена, раз это не хранится нигде, кроме комментария к движению склада) и `<ConfirmDialog>` с текстом «Остаток будет возвращён на склад автоматически».
- Показывать `partnerId` (если есть) со ссылкой на партнёра и (если есть права на отчёты) — расчётную комиссию (`totalAmount × partner.commissionPercentage / 100`, клиентский расчёт для справки — авторитетное значение живёт только в отчёте `GetSalesByPartner`).

### 12.4 InventoryMovement — форма Create

Селектор типа движения (`type`) **не включает `MOVEMENT_TYPE_SALE`** (это внутренний, серверный тип — руками не создаётся) и **не включает `MOVEMENT_TYPE_UNSPECIFIED`**. Знак `quantity` подсказывается UI по типу (Receipt → предлагать положительное, WriteOff → предлагать отрицательное с авто-минусом при вводе, Adjustment/Transfer — оба знака), но сервер — источник правды по итоговому ограничению (не уйти в минус).

### 12.5 User — профиль и смена пароля

Своя карточка «Профиль» (не через `/users/[id]` со списочными правами, а отдельный маршрут `/profile`, доступный любому залогиненному) — редактирование своих `name/lastName/description`, кнопка «Сменить пароль» (`ChangePassword`, требует текущий пароль). Список `/users` со всеми CRUD-действиями — только `admin`.

### 12.6 ProductImageUploader

Эндпоинт на бэкенде уже реализован (см. п.4.6), фичефлаг не нужен — компонент полностью активен с самого начала. Ключевые требования к реализации:
- Upload — `multipart/form-data` через `POST /api/products/[id]/images` (Nuxt server route, проксирует на `POST /images`, см. п.4.6), не напрямую на бэкенд-хост.
- Клиентская preflight-проверка перед отправкой (UX, не единственная защита — финальная проверка всё равно на сервере): MIME по содержимому файла (не по расширению) ∈ `image/jpeg|png|gif|webp`, размер ≤ лимиту (значение лимита — из `runtimeConfig.public`, зеркалит `crm.images.maxSizeBytes` на бэкенде, дефолт 10 MiB, если явно не задан).
- После успешной загрузки — полученный `id` добавляется в `form.imageIds` (не отправляется на бэкенд немедленно сам по себе): сохранение прикрепления происходит вместе с остальной формой Product через обычный `Update`/FieldMask, как любое другое изменённое поле — не отдельным запросом сразу после загрузки файла.
- Отображение уже загруженных изображений — через `<img :src="`/api/images/${id}`">` (или прямой URL, если выбран не-проксирующий вариант из п.4.6), с возможностью переупорядочить (drag-and-drop список, `imageIds` — упорядоченный массив, порядок значим — «Full replacement, in display order») и удалить (просто убрать id из массива, файл на диске бэкенда при этом не удаляется — таков контракт: `imageIds` только ссылается на уже загруженные файлы, а не управляет их временем жизни).

---

## 13. Наблюдаемость на клиенте (best practice)

- Глобальный `vue:error` / `app:error` хук Nuxt → отправка в лог-эндпоинт сервера (`server/api/log/client-error.js`) с контекстом (route, user id, timestamp) — не просто `console.error` в консоли пользователя, который никто не увидит.
- Идентификатор запроса (если бэкенд его генерирует через `request_id`-интерцептор — судя по `configs/grpc.yaml`, интерцептор есть) — пробрасывать из ответа сервера в клиентский error-toast мелким текстом, чтобы при обращении в поддержку можно было найти конкретный запрос в логах бэкенда.

---

## 14. Структура проекта (рекомендация)

```
app/
  components/
    form/          # LocalizedStringInput, RelationSelect, FormGrid, ...
    display/       # LocalizedText, StatusBadge, MoneyLabel, RelationLabel, ...
    layout/         # TopBar, SideMenu (плоское меню слева, см. п.8.2)
    common/         # ConfirmDialog, EtagConflictDialog, DashboardWidget, PermissionGate
    entity/         # EntityListPage, EntityDataTable, EntityForm (generic)
  composables/       # useEntityApi, usePermission, useAuth, useReferenceCache, useFormat*
  config/
    entities/        # один файл на сущность — единственное место с полями/колонками/правами
    navigation.js
    permissions.js
  design/tokens.js
  pages/
    index.vue                # Dashboard
    login.vue
    profile.vue
    reports/index.vue
    [entity]/index.vue        # generic список — рендерит <EntityListPage :entity="route.params.entity"/>
    [entity]/new.vue
    [entity]/[id].vue
    sales/new.vue              # кастомный мастер (не generic)
    sales/[id].vue              # кастомная detail-страница
server/
  api/
    auth/login.post.js
    auth/logout.post.js
    [entity]/list.post.js      # генерируется хелпером defineEntityHandler
    [entity]/get.post.js
    [entity]/create.post.js
    [entity]/update.patch.js
    [entity]/delete.delete.js
  middleware/csrf.js
  utils/grpcClient.js           # инициализация Connect-клиентов из runtimeConfig
  utils/session.js               # server-side session store (in-memory/Redis)
i18n/locales/{sr,en,ru}.json
```

---

## 15. Definition of Done для этого этапа

- [ ] Node.js gRPC-клиент (п.1.1) поднят внутри `server/api/**`, деплой Nitro сконфигурирован на Node server runtime (не edge/serverless-пресет).
- [ ] BFF-схема (п.2) реализована, токен не покидает сервер.
- [ ] Все 11 сущностей из бэкенд-ТЗ доступны через `EntityRegistry`, generic список+форма работают минимум для Brand/Category/Warehouse/Partner/Client (эталонный простой CRUD).
- [ ] Product/Price/Inventory/InventoryMovement — вкладочный флоу (п.12.1), включая рабочий `ProductImageUploader` (п.4.6/12.6).
- [ ] Sale — мастер создания + detail с UpdateStatus/Cancel (п.12.3).
- [ ] User — список (admin-only) + `/profile` + смена пароля.
- [ ] Dashboard — все 6 виджетов отчётов, каждый независимо грузится.
- [ ] `permissions.js` (п.4.3) сверен с реальным `configs/crm.yaml`/`internal/transports/grpc/interceptors/rbac/permissions.go` на бэкенде на момент деплоя (ключи из п.10) — расхождение не ловится автоматически, проверять руками при каждом релизе, пока не появится RPC со списком permissions.
- [ ] RBAC UI-гейтинг работает по `permissions.js`, и весь путь мутаций проходит через единый error-handler (п.9.4), корректно обрабатывающий `PermissionDenied` от реального enforcement на бэкенде (п.4.2) — не только скрывает кнопки, но и переживает случай «право отозвали между открытием страницы и кликом».
- [ ] i18n интерфейса (sr/en/ru) + LocalizedString-компоненты работают с фолбэком на `sr`.
- [ ] Адаптивность проверена на 4 разрешениях из п.6.
- [ ] CSP/security-заголовки включены (`@nuxt/security`), CSRF-миддлвар работает.
- [ ] Design tokens — единый файл, смена акцентного цвета не требует правок вне `tokens.js`.
