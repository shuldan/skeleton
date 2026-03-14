# Skeleton — Go Application Template

[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

**Skeleton** — `gonew`-совместимый шаблон Go-приложения с модульной DDD-архитектурой, встроенной системой событий, очередями, миграциями и Docker-окружением. Содержит один модуль (`task`) в качестве примера — удалите или замените его на свой домен.

---

## Содержание

- [Быстрый старт](#быстрый-старт)
- [Архитектура](#архитектура)
  - [Высокоуровневая схема](#высокоуровневая-схема)
  - [Поток запроса](#поток-запроса)
  - [Система событий](#система-событий)
  - [Принципы](#принципы)
- [Структура проекта](#структура-проекта)
- [Анатомия модуля](#анатомия-модуля)
  - [Domain](#domain)
  - [Application](#application)
  - [Infrastructure](#infrastructure)
  - [Presentation](#presentation)
  - [module.go — фасад](#modulego--фасад)
- [Конфигурация](#конфигурация)
  - [Файлы конфигурации](#файлы-конфигурации)
  - [Переменные окружения](#переменные-окружения)
  - [Профили окружения](#профили-окружения)
- [CLI-команды](#cli-команды)
- [База данных](#база-данных)
  - [Подключения](#подключения)
  - [Миграции](#миграции)
  - [Репозитории](#репозитории)
- [События](#события)
  - [Определение событий](#определение-событий)
  - [Накопление событий в агрегате](#накопление-событий-в-агрегате)
  - [Emitter (публикация)](#emitter-публикация)
  - [Listener (подписка)](#listener-подписка)
  - [OutboundRelay (пересылка в очередь)](#outboundrelay-пересылка-в-очередь)
  - [Полный цикл события в task-модуле](#полный-цикл-события-в-task-модуле)
- [Очереди](#очереди)
- [HTTP-сервер](#http-сервер)
  - [Маршрутизация](#маршрутизация)
  - [Middleware](#middleware)
  - [Обработчики](#обработчики)
  - [Обработка ошибок](#обработка-ошибок)
- [Presenter-паттерн](#presenter-паттерн)
- [Docker](#docker)
  - [Production](#production)
  - [Development](#development)
  - [Управление](#управление)
- [Тестирование API](#тестирование-api)
- [Makefile](#makefile)
- [Создание нового модуля](#создание-нового-модуля)
- [Лицензия](#лицензия)

---

## Быстрый старт

### Создание проекта из шаблона

```bash
gonew github.com/shuldan/skeleton github.com/yourname/myapp
cd myapp
```

`gonew` автоматически заменит пути импортов во всех `.go`-файлах и `go.mod`.

### Локальный запуск

```bash
# 1. Скопируйте файлы окружения
cp .env.example .env
cp deployments/.env.example deployments/.env

# 2. Поднимите PostgreSQL (или используйте существующий)
make docker-dev

# 3. Выполните миграции
make migrate

# 4. Запустите сервер
make run
```

Сервер доступен на `http://localhost:8080`.

### Проверка

```bash
# Создать задачу
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{"title": "Hello", "description": "World"}'

# Список задач
curl http://localhost:8080/api/v1/tasks
```

---

## Архитектура

### Высокоуровневая схема

```mermaid
graph TB
    CLI["cmd/app/main.go<br/><i>точка входа, CLI</i>"]
    BS["internal/bootstrap/<br/><i>инициализация, DI, маршруты</i>"]
    
    CLI --> BS
    
    BS --> M1["module/task<br/><i>пример</i>"]
    BS --> M2["module/order<br/><i>ваш код</i>"]
    BS --> M3["module/...<br/><i>ваш код</i>"]

    subgraph module["Структура модуля"]
        direction TB
        PR["presentation<br/>HTTP, Listeners, Jobs"]
        AP["application<br/>Interactors, Operations,<br/>Emitters, Ports"]
        DM["domain<br/>Models, Value Objects,<br/>Interfaces"]
        IN["infrastructure<br/>Repositories, Adapters,<br/>Migrations"]

        PR --> AP
        AP --> DM
        IN --> DM
    end

    M1 -.-> module

    style CLI fill:#e1f5fe
    style BS fill:#e1f5fe
    style DM fill:#fff3e0
    style AP fill:#e8f5e9
    style PR fill:#f3e5f5
    style IN fill:#fce4ec
```

### Поток запроса

```mermaid
sequenceDiagram
    participant C as Client
    participant R as Router
    participant MW as Middleware
    participant H as Handler
    participant I as Interactor
    participant O as Operation
    participant Repo as Repository
    participant Agg as Aggregate
    participant E as Emitter
    participant D as Dispatcher

    C->>R: HTTP Request
    R->>MW: Recovery → RequestID → Logging
    MW->>H: handler.handle(w, r)
    H->>H: Bind input, create output
    H->>I: interactor.Handle(ctx, input, output)
    I->>I: Validate (NewTitle, etc.)
    I->>O: operation.Create(ctx, ...)
    O->>Agg: NewTask() → record(TaskCreated)
    O->>Repo: repo.Save(ctx, task)
    Repo-->>O: nil / error
    O-->>I: *Task, error
    I->>I: task.RepresentTo(output)
    I->>I: task.ReleaseEvents()
    I->>E: emitter.Emit(ctx, []event.Event)
    E->>D: dispatcher.Publish(ctx, event) × N
    I-->>H: nil / error
    H->>C: JSON Response
```

### Система событий

```mermaid
graph LR
    AGG["Aggregate<br/><i>record(event)</i>"] --> INT["Interactor<br/><i>ReleaseEvents()</i>"]
    INT --> EM["Emitter<br/><i>[]event.Event</i>"]
    EM --> DISP["Dispatcher"]
    
    DISP --> L1["Listener<br/><i>in-process</i>"]
    DISP --> RELAY["OutboundRelay"]
    RELAY --> BROKER["Broker<br/><i>memory / redis / ...</i>"]
    BROKER --> JOB["Job / Consumer<br/><i>async worker</i>"]

    style AGG fill:#fff3e0
    style INT fill:#e8f5e9
    style DISP fill:#fff3e0
    style L1 fill:#f3e5f5
    style BROKER fill:#e1f5fe
    style JOB fill:#fce4ec
```

**Поток событий:**

1. Агрегат **накапливает** события через `record()` при изменении состояния
2. Interactor вызывает `task.ReleaseEvents()` — забирает и очищает очередь
3. Emitter итерирует по слайсу и публикует каждое событие в Dispatcher
4. Dispatcher доставляет события в Listener (in-process) и OutboundRelay (→ очередь)
5. Job/Consumer читает из очереди и обрабатывает асинхронно

### Принципы

| Принцип | Реализация |
|---------|-----------|
| **DDD** | Агрегаты, value objects, domain events (накопление в агрегате), репозитории |
| **Clean Architecture** | Зависимости направлены внутрь: presentation → application → domain ← infrastructure |
| **Модульность** | Каждый домен — изолированный модуль с фасадом `module.go` |
| **Presenter-паттерн** | Домен не знает о JSON/HTTP; данные передаются через `TaskPresenter` |
| **Snapshot-паттерн** | Персистентность через плоские снимки агрегата, без экспорта приватных полей |
| **CQRS-lite** | Разделение операций записи (operations/interactors) и чтения |
| **Event-Driven** | Агрегат накапливает события → emitter публикует пакетом → relay пересылает в очередь |
| **Structural typing** | Модули определяют свой `Logger` интерфейс, совместимый с `framework/logger` без импорта |

---

## Структура проекта

```
skeleton/
├── cmd/
│   └── app/
│       └── main.go                  # Точка входа
│
├── config/
│   ├── config.yaml                  # Базовая конфигурация
│   ├── config.development.yaml      # Переопределения для dev
│   └── config.production.yaml       # Переопределения для prod
│
├── internal/
│   ├── bootstrap/
│   │   ├── app.go                   # Инициализация: ядро, БД, модули, команды
│   │   └── router.go                # Сборка HTTP-маршрутизатора
│   │
│   ├── event/
│   │   ├── event.go                 # Интерфейс Event (доменная граница)
│   │   └── task.go                  # Глобальные определения событий
│   │
│   ├── logger/
│   │   └── logger.go                # Интерфейс Logger (structural typing)
│   │
│   └── module/
│       └── task/                    # ── Пример модуля ──
│           ├── module.go            # Фасад: Routes, Listeners, Consumers, Migrations
│           ├── domain/              # Ядро бизнес-логики
│           ├── application/         # Use-case оркестрация
│           ├── infrastructure/      # Реализации интерфейсов
│           └── presentation/        # HTTP, listeners, jobs
│
├── deployments/
│   ├── .env.example
│   ├── docker-compose.yml           # Production
│   ├── docker-compose.dev.yml       # Development overlay
│   └── docker/
│       ├── Dockerfile               # Multi-stage production
│       ├── golang/Dockerfile        # Dev-контейнер
│       └── postgres/                # PostgreSQL + init-скрипты
│
├── test/
│   └── http/                        # HTTP-клиент файлы
│
├── Makefile
├── go.mod
├── .env.example
├── .gitignore
├── .dockerignore
└── LICENSE
```

---

## Анатомия модуля

```mermaid
graph TB
    subgraph module["module/task"]
        FACADE["module.go<br/><i>фасад</i>"]

        subgraph presentation
            API["api/<br/>HTTP handlers"]
            LST["listener/<br/>event handlers"]
            JOB["job/<br/>queue consumers"]
        end

        subgraph application
            INT["interactor/<br/>use-case"]
            BIZ_OP["business/operation/<br/>реализации"]
            BIZ_EM["business/emitter/<br/>реализации"]
            PORT["port/<br/>внешние интерфейсы"]
        end

        subgraph domain
            MDL["model/<br/>агрегаты, VO, ошибки"]
            PER["persistence/<br/>интерфейс репозитория"]
            D_OP["business/operation/<br/>интерфейсы"]
            D_EM["business/emitter/<br/>интерфейсы"]
        end

        subgraph infrastructure
            I_PER["persistence/<br/>SQL-репозиторий"]
            MIG["migration/<br/>миграции"]
            ADP["adapter/<br/>адаптеры портов"]
        end

        API --> INT
        LST --> INT
        JOB --> PORT

        INT --> BIZ_OP
        INT --> BIZ_EM
        BIZ_OP --> D_OP
        BIZ_EM --> D_EM
        BIZ_OP --> PER

        I_PER --> PER
        I_PER --> MDL
        ADP --> PORT

        FACADE --> API
        FACADE --> LST
        FACADE --> JOB
        FACADE --> MIG
    end

    style FACADE fill:#e1f5fe
    style domain fill:#fff3e0,stroke:#e65100
    style application fill:#e8f5e9,stroke:#2e7d32
    style presentation fill:#f3e5f5,stroke:#6a1b9a
    style infrastructure fill:#fce4ec,stroke:#b71c1c
```

### Domain

Ядро бизнес-логики. Не зависит ни от чего внешнего.

```
domain/
├── model/
│   ├── task.go              # Агрегат (приватные поля, поведение, накопление событий)
│   ├── task_id.go           # Value object TaskID
│   ├── title.go             # Value object Title (с валидацией)
│   ├── status.go            # Value object Status (state machine)
│   ├── task_snapshot.go     # Snapshot / Restore для персистентности
│   ├── task_presenter.go    # Интерфейс TaskPresenter (output boundary)
│   └── errors.go            # Типизированные доменные ошибки
├── persistence/
│   └── task_repository.go   # Интерфейс TaskRepository
└── business/
    ├── operation/
    │   ├── creating_operation.go    # Интерфейс CreatingOperation
    │   └── completing_operation.go  # Интерфейс CompletingOperation
    └── emitter/
        └── event_emitter.go         # Интерфейс EventEmitter
```

**Агрегат** инкапсулирует состояние, бизнес-правила и накапливает доменные события:

```go
type Task struct {
    id          TaskID
    title       Title
    description string
    status      status
    version     int
    events      []event.Event   // очередь доменных событий
}

// Конструктор — создаёт задачу и записывает событие TaskCreated
func NewTask(id TaskID, title Title, description string) *Task {
    task := &Task{
        id: id, title: title, description: description,
        status: statusDraft, version: 1,
    }
    task.record(&event.TaskCreated{
        BaseEvent: events.NewBaseEvent("TaskCreated", task.id.String()),
        TaskID:    task.id.String(),
        Title:     title.String(),
    })
    return task
}

// Поведение — изменение состояния + запись события
func (t *Task) Complete() error {
    newStatus, err := t.status.transitionTo(statusDone)
    if err != nil {
        return err
    }
    t.status = newStatus
    t.record(&event.TaskCompleted{...})
    return nil
}

// ReleaseEvents — забирает накопленные события и очищает очередь
func (t *Task) ReleaseEvents() []event.Event {
    evts := t.events
    t.events = nil
    return evts
}

func (t *Task) record(e event.Event) {
    t.events = append(t.events, e)
}
```

> **Паттерн:** события не публикуются немедленно. Агрегат накапливает их через `record()`, а interactor забирает пакетом через `ReleaseEvents()` после успешного сохранения. Это гарантирует, что события публикуются только при успешной операции.

**Value objects** содержат валидацию при создании:

```go
func NewTitle(raw string) (Title, error) {
    if strings.TrimSpace(raw) == "" {
        return Title{}, ErrTitleRequired
    }
    return Title{value: raw}, nil
}
```

**Status** реализует state machine с явными переходами:

```mermaid
stateDiagram-v2
    [*] --> draft : NewTask()
    draft --> in_progress
    draft --> done
    in_progress --> done
    done --> [*]
```

```go
var transitions = map[string][]status{
    "draft":       {statusInProgress, statusDone},
    "in_progress": {statusDone},
}
```

> **Инкапсуляция:** тип `status` и его методы неэкспортируемые. Вне пакета `model` нельзя создать произвольный статус или вызвать переход напрямую — только через методы агрегата.

### Application

Оркестрация use-case'ов. Зависит только от domain.

```
application/
├── interactor/
│   ├── create_task_interactor.go
│   ├── complete_task_interactor.go
│   ├── get_task_interactor.go
│   └── list_tasks_interactor.go
├── business/
│   ├── operation/
│   │   ├── creating_operation.go     # Реализация CreatingOperation
│   │   └── completing_operation.go   # Реализация CompletingOperation
│   └── emitter/
│       └── task_emitter.go           # Единый emitter для всех событий Task
└── port/
    └── notification_port.go          # Порт для внешних уведомлений
```

**Interactor** — точка входа use-case'а. Принимает `Input` и `Output` интерфейсы:

```go
func (i *CreateTaskInteractor) Handle(
    ctx context.Context,
    input CreateTaskInput,    // интерфейс: GetTitle(), GetDescription()
    output CreateTaskOutput,  // интерфейс: TaskPresenter
) error {
    title, err := model.NewTitle(input.GetTitle())
    if err != nil {
        return err
    }

    task, err := i.creatingOp.Create(ctx, title, input.GetDescription())
    if err != nil {
        return err
    }

    task.RepresentTo(output)              // заполнение output через presenter
    i.emitter.Emit(ctx, task.ReleaseEvents())  // публикация пакета событий
    return nil
}
```

> **Порядок:** сначала `RepresentTo` (данные для ответа), затем `ReleaseEvents` + `Emit` (публикация событий). События публикуются только после успешного сохранения в Operation.

**Operation** — изолированная бизнес-операция (запись в репозиторий):

```go
func (o *CreatingOperation) Create(
    ctx context.Context, title model.Title, description string,
) (*model.Task, error) {
    id := model.NewTaskID(uuid.New().String())
    task := model.NewTask(id, title, description)  // → record(TaskCreated)

    if err := o.repo.Save(ctx, task); err != nil {
        return nil, err
    }
    return task, nil
}
```

**Emitter** — единый для всех событий агрегата. Принимает слайс событий и публикует каждое в Dispatcher:

```go
func (e *TaskEmitter) Emit(ctx context.Context, domainEvents []event.Event) {
    publishCtx := context.WithoutCancel(ctx)  // события отправляются даже при отмене запроса

    for _, ev := range domainEvents {
        dispatchable, ok := ev.(events.Event)
        if !ok {
            e.log.Error("event does not implement events.Event", ...)
            continue
        }
        if err := e.dispatcher.Publish(publishCtx, dispatchable); err != nil {
            e.log.Error("failed to emit event", ...)
        }
    }
}
```

> **`context.WithoutCancel`**: emitter публикует события в контексте, не привязанном к HTTP-запросу. Если клиент отключился, события всё равно доставляются подписчикам.

> **Граница доменного и инфраструктурного события:** домен работает с интерфейсом `event.Event` (определён в `internal/event`), а emitter кастит к `events.Event` из библиотеки для публикации через Dispatcher. Домен не зависит от инфраструктурной библиотеки событий.

**Port** — интерфейс к внешней системе, реализуемый в infrastructure:

```go
type NotificationPort interface {
    Send(ctx context.Context, taskID, message string) error
}
```

### Infrastructure

Реализация интерфейсов домена. Зависит от domain (реализует его интерфейсы).

```
infrastructure/
├── persistence/
│   └── task_repository.go       # SQL-реализация TaskRepository
├── migration/
│   └── create_tasks_table.go    # Миграция: CREATE TABLE tasks
└── adapter/
    └── notification_adapter.go  # Реализация NotificationPort (лог)
```

**Репозиторий** использует Snapshot-паттерн — агрегат не экспортирует поля напрямую:

```go
func scanTask(sc repository.Scanner) (*model.Task, error) {
    var s model.TaskSnapshot
    if err := sc.Scan(
        &s.ID, &s.Title, &s.Description, &s.Status, &s.Version,
    ); err != nil {
        return nil, err
    }
    return s.Restore()  // восстановление агрегата из снимка (без записи событий)
}

func taskValues(t *model.Task) []any {
    s := t.Snapshot()   // получение плоских данных для INSERT/UPDATE
    return []any{s.ID, s.Title, s.Description, s.Status, s.Version}
}
```

> **Snapshot vs конструктор:** `NewTask()` записывает `TaskCreated` событие, а `Restore()` — нет. При чтении из БД события не генерируются.

**Маппинг ошибок репозитория:**

```go
func (r *taskRepository) FindByID(ctx context.Context, id model.TaskID) (*model.Task, error) {
    task, err := r.repo.Find(ctx, id.String())
    if errors.Is(err, repository.ErrNotFound) {
        return nil, model.ErrTaskNotFound.WithDetail("id", id.String())
    }
    return task, err
}
```

Инфраструктурные ошибки (`repository.ErrNotFound`, `repository.ErrConcurrentModification`) транслируются в доменные (`model.ErrTaskNotFound`, `model.ErrConcurrentModification`).

### Presentation

Адаптеры ввода: HTTP, event listeners, queue consumers.

```
presentation/
├── api/
│   ├── create_task_handler.go     # POST /api/v1/tasks
│   ├── list_tasks_handler.go      # GET  /api/v1/tasks
│   ├── get_task_handler.go        # GET  /api/v1/tasks/{id}
│   ├── complete_task_handler.go   # POST /api/v1/tasks/{id}/complete
│   └── errors.go                  # API-специфичные ошибки
├── listener/
│   └── task_completed_listener.go # In-process обработка TaskCompleted
└── job/
    └── send_notification_job.go   # Queue consumer: отправка уведомлений
```

**Handler** реализует Input/Output интерфейсы interactor'а:

```go
func (h *createTaskHandler) handle(w http.ResponseWriter, r *http.Request) error {
    var body createTaskBody
    if err := httpserver.Bind(r, &body); err != nil {
        return err
    }

    output := &CreateTaskOutput{}   // реализует TaskPresenter
    if err := h.interactor.Handle(
        r.Context(), &createTaskInput{body: body}, output,
    ); err != nil {
        return err
    }

    httpserver.Created(w, output)   // 201 + JSON
    return nil
}
```

**UUID-валидация** происходит в interactor'ах через `uuid.MustParse()`. При невалидном ID — паника, которую перехватывает middleware `Recovery` и возвращает клиенту `500 Internal Error`. Для кастомной ошибки валидации можно добавить проверку в handler.

### module.go — фасад

Единственная точка контакта модуля с внешним миром. Собирает граф зависимостей и предоставляет методы регистрации:

```go
// NewModule собирает внутренний граф зависимостей
func NewModule(db *sql.DB, dispatcher *events.Dispatcher, log logger.Logger) *Module

// Регистрация компонентов — вызываются из bootstrap
func (m *Module) Routes(router *httpserver.Router)
func (m *Module) Listeners(d *events.Dispatcher)
func (m *Module) Relays(relay *eventbus.OutboundRelay)
func (m *Module) Consumers(qw *queueworker.Module, broker queue.Broker)
func (m *Module) Migrations(runner *migration.Runner)
```

> **`logger.Logger`**: модуль принимает интерфейс `logger.Logger` из `internal/logger`, а не конкретный тип из фреймворка. Благодаря structural typing, `framework/logger.Logger` удовлетворяет этому интерфейсу автоматически.

---

## Конфигурация

### Файлы конфигурации

Конфигурация использует YAML с поддержкой Go-шаблонов:

```yaml
# config/config.yaml — базовая конфигурация
database:
  connections:
    default:
      driver: postgres
      dsn: "{{ env \"DATABASE_URL\" | default \"postgres://...\" }}"
      max_open_conns: 25
```

### Переменные окружения

Переменные окружения подставляются через функцию `env` в шаблоне:

| Переменная | Описание | Значение по умолчанию |
|-----------|---------|---------------------|
| `APP_ENV` | Профиль окружения | `development` |
| `DATABASE_URL` | DSN подключения к БД | `postgres://postgres:postgres@localhost:5432/skeleton?sslmode=disable` |

### Профили окружения

```mermaid
graph LR
    BASE["config.yaml<br/><i>базовые значения</i>"] --> MERGE["Итоговая конфигурация"]
    PROFILE["config.{APP_ENV}.yaml<br/><i>переопределения</i>"] --> MERGE
    ENV["Переменные окружения<br/><i>env template func</i>"] --> MERGE

    style BASE fill:#e8f5e9
    style PROFILE fill:#fff3e0
    style ENV fill:#e1f5fe
```

При `APP_ENV=development` загружаются файлы в порядке:

1. `config/config.yaml` — базовые значения
2. `config/config.development.yaml` — переопределения

```yaml
# config/config.development.yaml
log:
  level: debug
  format: text

# config/config.production.yaml
server:
  port: 443
log:
  level: warn
```

### Просмотр итоговой конфигурации

```bash
make config
# или
go run ./cmd/app config:dump
```

---

## CLI-команды

Приложение использует единый бинарник с подкомандами:

| Команда | Описание |
|---------|---------|
| `serve` | Запуск HTTP-сервера + фоновые воркеры |
| `queue:work` | Запуск только воркеров очередей (без HTTP) |
| `migrate:up` | Применить все pending-миграции |
| `migrate:down` | Откатить последнюю миграцию |
| `migrate:status` | Показать статус миграций |
| `migrate:plan` | Показать план pending-миграций |
| `health` | Проверка здоровья (БД и др.) |
| `config:dump` | Вывод итоговой конфигурации |

```bash
# Через go run
go run ./cmd/app serve
go run ./cmd/app migrate:up

# Через собранный бинарник
./bin/app serve
./bin/app queue:work
```

---

## База данных

### Подключения

Поддерживается несколько именованных подключений. По умолчанию используется `default`:

```yaml
database:
  connections:
    default:
      driver: postgres
      dsn: "{{ env \"DATABASE_URL\" }}"
      max_open_conns: 25
      max_idle_conns: 5
      conn_max_lifetime: 5m
```

`database.Manager` управляет пулами соединений и предоставляет `*sql.DB` по имени:

```go
dbm.Default()                // *sql.DB для "default"
dbm.Connection("analytics")  // *sql.DB для другого подключения
```

### Миграции

Миграции определяются в коде модуля и регистрируются через фасад:

```go
// infrastructure/migration/create_tasks_table.go
func CreateTasksTable() migrator.Migration {
    return migrator.CreateMigration(
        "20240101_001_create_tasks",    // уникальный ID
        "Create tasks table",           // описание
    ).CreateTable("tasks",
        "id UUID PRIMARY KEY",
        "title VARCHAR(255) NOT NULL",
        "description TEXT NOT NULL DEFAULT ''",
        "status VARCHAR(50) NOT NULL DEFAULT 'draft'",
        "version INTEGER NOT NULL DEFAULT 1",
        "created_at TIMESTAMP NOT NULL DEFAULT NOW()",
        "updated_at TIMESTAMP NOT NULL DEFAULT NOW()",
    ).MustBuild()
}

// module.go
func (m *Module) Migrations(runner *migration.Runner) {
    runner.Register("default", taskmigration.CreateTasksTable())
}
```

Миграции используют advisory lock для безопасного параллельного запуска.

```bash
make migrate          # Применить
make migrate-down     # Откатить последнюю
make migrate-status   # Статус
make migrate-plan     # План
```

### Репозитории

Репозиторий строится на основе `shuldan/repository` с поддержкой:

- **Optimistic locking** через `version` column
- **Snapshot-паттерн** — маппинг агрегата через `Snapshot()` / `Restore()`
- **Типизированные ошибки** — `ErrNotFound`, `ErrConcurrentModification`

```go
repo := repository.New(
    db,
    repository.Postgres(),
    repository.Simple(repository.SimpleConfig[*model.Task]{
        Table:  taskTable,
        Scan:   scanTask,
        Values: taskValues,
    }),
)
```

---

## События

Система событий состоит из трёх уровней:

```mermaid
graph LR
    subgraph "Внутри агрегата"
        AGG["Aggregate<br/>record()"] --> REL["ReleaseEvents()"]
    end

    subgraph "Синхронно (in-process)"
        REL --> EM["Emitter<br/>[]event.Event"]
        EM --> DISP["Dispatcher"]
        DISP --> LST["Listener"]
    end

    subgraph "Асинхронно (через очередь)"
        DISP --> RLY["OutboundRelay"]
        RLY --> BRK["Broker"]
        BRK --> JOB["Job"]
    end

    style AGG fill:#fff3e0
    style EM fill:#e8f5e9
    style DISP fill:#fff3e0
    style LST fill:#f3e5f5
    style BRK fill:#e1f5fe
    style JOB fill:#fce4ec
```

### Определение событий

Глобальные структуры событий объявляются в `internal/event/`:

```go
// internal/event/event.go — доменная граница событий
type Event interface {
    EventName() string
    OccurredAt() time.Time
    AggregateID() string
}
```

```go
// internal/event/task.go — конкретные события
type TaskCreated struct {
    events.BaseEvent
    TaskID string `json:"task_id"`
    Title  string `json:"title"`
}

type TaskCompleted struct {
    events.BaseEvent
    TaskID string `json:"task_id"`
}
```

> **Граница:** домен зависит от `internal/event.Event` (свой интерфейс), а не от `shuldan/events.Event` (библиотека). Emitter в application-слое кастит к библиотечному типу при публикации. Так домен остаётся независимым от инфраструктуры.

### Накопление событий в агрегате

Агрегат **не публикует** события немедленно. Он записывает их во внутреннюю очередь:

```go
// Конструктор — записывает TaskCreated
func NewTask(id TaskID, title Title, description string) *Task {
    task := &Task{...}
    task.record(&event.TaskCreated{
        BaseEvent: events.NewBaseEvent("TaskCreated", task.id.String()),
        TaskID:    task.id.String(),
        Title:     title.String(),
    })
    return task
}

// Метод поведения — записывает TaskCompleted
func (t *Task) Complete() error {
    newStatus, err := t.status.transitionTo(statusDone)
    if err != nil {
        return err
    }
    t.status = newStatus
    t.record(&event.TaskCompleted{
        BaseEvent: events.NewBaseEvent("TaskCompleted", t.id.String()),
        TaskID:    t.id.String(),
    })
    return nil
}

// ReleaseEvents — interactor забирает события после успешного сохранения
func (t *Task) ReleaseEvents() []event.Event {
    evts := t.events
    t.events = nil
    return evts
}
```

> **Гарантия:** если `repo.Save()` вернул ошибку, interactor не вызывает `ReleaseEvents()` — события не публикуются. Если `Restore()` восстанавливает агрегат из БД — события не записываются.

### Emitter (публикация)

Единый emitter принимает **слайс событий** и публикует каждое в Dispatcher:

```go
func (e *TaskEmitter) Emit(ctx context.Context, domainEvents []event.Event) {
    publishCtx := context.WithoutCancel(ctx)

    for _, ev := range domainEvents {
        dispatchable, ok := ev.(events.Event)
        if !ok {
            e.log.Error("event does not implement events.Event", ...)
            continue
        }
        if err := e.dispatcher.Publish(publishCtx, dispatchable); err != nil {
            e.log.Error("failed to emit event", ...)
        }
    }
}
```

**Ключевые решения:**

| Решение | Причина |
|---------|---------|
| `context.WithoutCancel(ctx)` | Клиент отключился — события всё равно доставляются |
| Один emitter на агрегат | Проще, чем per-event emitter. Новое событие = 0 новых файлов в emitter |
| Каст `event.Event` → `events.Event` | Домен не зависит от библиотеки; адаптация на границе |

**Доменный интерфейс:**

```go
// domain/business/emitter/event_emitter.go
type EventEmitter interface {
    Emit(ctx context.Context, events []event.Event)
}
```

### Listener (подписка)

In-process подписка через типизированный `Subscribe`:

```go
// presentation/listener/task_completed_listener.go
type TaskCompletedListener struct {
    log logger.Logger
}

func (l *TaskCompletedListener) Handle(
    _ context.Context, e *event.TaskCompleted,
) error {
    l.log.Info("task completed event received", "task_id", e.TaskID)
    return nil
}

// module.go — регистрация struct-based listener
func (m *Module) Listeners(d *events.Dispatcher) {
    events.Subscribe(d, listener.NewTaskCompletedListener(m.log))
}
```

> **`events.Subscribe` vs `events.SubscribeFunc`**: skeleton использует struct-based `Subscribe` — listener реализует `Handle(ctx, *EventType) error`. Для простых случаев можно использовать `SubscribeFunc` с замыканием.

### OutboundRelay (пересылка в очередь)

Для асинхронной обработки события пересылаются из Dispatcher в message broker через `OutboundRelay`:

```go
func (m *Module) Relays(relay *eventbus.OutboundRelay) {
    relay.Forward("TaskCompleted", "task.completed",
        eventbus.WithTransform(func(e events.Event) ([]byte, error) {
            return json.Marshal(map[string]string{
                "task_id": e.AggregateID(),
                "event":   e.EventName(),
            })
        }),
    )
}
```

> **`WithTransform` vs Envelope:** по умолчанию `OutboundRelay` оборачивает события в стандартный `Envelope` (см. документацию фреймворка). `WithTransform` обходит Envelope и позволяет задать кастомную сериализацию. Используйте `WithTransform` для простых случаев или интеграции со сторонними системами; Envelope — для межсервисной коммуникации с трассировкой.

### Полный цикл события в task-модуле

```mermaid
sequenceDiagram
    participant H as Handler
    participant I as Interactor
    participant O as Operation
    participant T as Task (aggregate)
    participant R as Repository
    participant E as Emitter
    participant D as Dispatcher
    participant L as Listener (in-process)
    participant RL as OutboundRelay
    participant B as Broker
    participant J as Job (async)
    participant N as NotificationPort

    H->>I: Handle(ctx, input, output)
    I->>O: Create(ctx, title, desc)
    O->>T: NewTask() → record(TaskCreated)
    O->>R: Save(task)
    R-->>O: ok
    O-->>I: task
    I->>T: RepresentTo(output)
    I->>T: ReleaseEvents() → [TaskCreated]
    I->>E: Emit(ctx, [TaskCreated])
    E->>D: Publish(TaskCreated)
    D->>L: Handle(TaskCreated)
    D->>RL: Forward → transform → Produce
    RL->>B: Produce("task.completed", data)
    Note over B,J: Асинхронно (queue:work)
    B->>J: Consume("task.completed")
    J->>N: Send(taskID, message)
```

**Для `Complete()` цикл аналогичен:** `task.Complete()` → `record(TaskCompleted)` → `ReleaseEvents()` → Emitter → Dispatcher → Listener + OutboundRelay → Broker → Job.

---

## Очереди

По умолчанию используется in-memory broker (`memorymq`). Для production замените на вашу реализацию `queue.Broker` (RabbitMQ, Kafka, Redis Streams и т.д.).

```go
// bootstrap/app.go
broker := memorymq.New()

// module.go — регистрация consumer'а
func (m *Module) Consumers(qw *queueworker.Module, broker queue.Broker) {
    j := job.NewSendNotificationJob(broker, m.notifier)
    qw.Register(queueworker.Registration{
        Name: "task-notification",
        Run:  j.Run,
    })
}
```

**Job** потребляет сообщения из topic'а:

```go
func (j *SendNotificationJob) Run(ctx context.Context) error {
    return j.broker.Consume(ctx, "task.completed", func(data []byte) error {
        return j.process(ctx, data)
    })
}
```

Запуск воркера:

```bash
make run-worker
# или
go run ./cmd/app queue:work
```

> **`serve` vs `queue:work`**: команда `serve` запускает HTTP-сервер **и** воркеры очередей. Команда `queue:work` запускает **только** воркеры (без HTTP). Используйте `queue:work` для выделенных worker-нод.

---

## HTTP-сервер

### Маршрутизация

Маршруты регистрируются в `module.go` через `httpserver.Router`:

```go
func (m *Module) Routes(router *httpserver.Router) {
    group := router.Group("/api/v1/tasks")

    group.POST("", api.NewCreateTaskHandler(m.createInteractor))
    group.GET("", api.NewListTasksHandler(m.listInteractor))
    group.GET("/{id}", api.NewGetTaskHandler(m.getInteractor))
    group.POST("/{id}/complete", api.NewCompleteTaskHandler(m.completeInteractor))
}
```

### Middleware

Глобальные middleware настраиваются в `bootstrap/router.go`:

```go
router.Use(
    middleware.Recovery(log.Error),   // Перехват паник → 500 JSON
    middleware.RequestID(),           // X-Request-ID генерация/проброс
    middleware.Logging(log.Info),     // Логирование запросов
)
```

> **Recovery и UUID-валидация:** `uuid.MustParse()` в interactor'ах паникует при невалидном ID. Middleware `Recovery` перехватывает панику и возвращает клиенту `500 Internal Error` с JSON-телом, логируя стектрейс.

### Обработчики

Обработчики оборачиваются через `httpserver.Wrap`, который принимает `func(w, r) error` и автоматически обрабатывает ошибки:

```go
func NewCreateTaskHandler(inter *interactor.CreateTaskInteractor) http.HandlerFunc {
    h := &createTaskHandler{interactor: inter}
    return httpserver.Wrap(h.handle)
}
```

### Обработка ошибок

Типизированные ошибки (`shuldan/errors`) автоматически маппятся в HTTP-ответы:

| Kind | HTTP Status |
|------|------------|
| `errors.Validation` | 400 Bad Request |
| `errors.NotFound` | 404 Not Found |
| `errors.Conflict` | 409 Conflict |
| `errors.DomainRule` | 422 Unprocessable Entity |
| `errors.Infrastructure` | 503 Service Unavailable |

```go
// Определение ошибки (domain/model/errors.go)
var taskCode = errors.WithPrefix("TASK")

var ErrTitleRequired = taskCode("TITLE_REQUIRED").
    Kind(errors.Validation).
    New("task title is required")

// При возврате из handler → автоматически 400 с JSON:
// {"code": "TASK_TITLE_REQUIRED", "message": "task title is required"}
```

> **Префиксы ошибок**: доменные ошибки используют префикс `TASK` (`taskCode`), API-специфичные — `TASK_API` (`apiCode`). Это позволяет различать источник ошибки в логах и ответах.

---

## Presenter-паттерн

Домен не знает о формате вывода. Данные передаются через интерфейс `TaskPresenter`:

```mermaid
graph TB
    AGG["Task aggregate<br/><i>приватные поля</i>"] -->|RepresentTo| PI["TaskPresenter<br/><i>интерфейс</i>"]

    PI --> HTTP["CreateTaskOutput<br/><i>JSON struct</i>"]
    PI --> EVT["taskEvent<br/><i>данные для event</i>"]
    PI --> TEST["mockPresenter<br/><i>для тестов</i>"]

    style AGG fill:#fff3e0
    style PI fill:#e1f5fe
    style HTTP fill:#f3e5f5
    style EVT fill:#e8f5e9
    style TEST fill:#fce4ec
```

```go
// domain/model/task_presenter.go
type TaskPresenter interface {
    SetID(id string) TaskPresenter
    SetTitle(title string) TaskPresenter
    SetDescription(description string) TaskPresenter
    SetStatus(status string) TaskPresenter
    SetVersion(version int) TaskPresenter
}

// Агрегат заполняет presenter (fluent API)
func (t *Task) RepresentTo(p TaskPresenter) {
    p.SetID(t.id.String()).
        SetTitle(t.title.String()).
        SetDescription(t.description).
        SetStatus(t.status.String()).
        SetVersion(t.version)
}
```

Каждый слой реализует presenter по-своему:

```go
// presentation/api/ — для HTTP-ответа
type CreateTaskOutput struct {
    ID     string `json:"id"`
    Title  string `json:"title"`
    Status string `json:"status"`
    // ...
}
func (o *CreateTaskOutput) SetID(v string) model.TaskPresenter { o.ID = v; return o }
```

---

## Structural typing для Logger

Skeleton определяет собственный интерфейс логгера в `internal/logger/`:

```go
// internal/logger/logger.go
type Logger interface {
    Debug(msg string, args ...any)
    Info(msg string, args ...any)
    Warn(msg string, args ...any)
    Error(msg string, args ...any)
}
```

Модули зависят от **этого интерфейса**, а не от конкретного `*framework/logger.Logger`. Благодаря structural typing в Go, `*logger.Logger` из фреймворка удовлетворяет этому интерфейсу автоматически — без адаптеров, без импорта фреймворка в доменный код.

```go
// module.go — принимает интерфейс
func NewModule(db *sql.DB, dispatcher *events.Dispatcher, log logger.Logger) *Module

// bootstrap/app.go — передаёт конкретный тип
taskMod := task.NewModule(dbm.Default(), bus.Dispatcher(), log)  // *framework/logger.Logger
```

---

## Docker

### Production

```mermaid
graph LR
    subgraph Docker Compose
        PG["postgres<br/>PostgreSQL 16"]
        MIG["migrate<br/><i>one-shot</i>"]
        APP["app<br/>HTTP serve"]
        WRK["worker<br/>queue:work"]
    end

    MIG -->|depends_on healthy| PG
    APP -->|depends_on completed| MIG
    WRK -->|depends_on completed| MIG

    CLIENT["Client"] --> APP

    style PG fill:#e1f5fe
    style MIG fill:#fff3e0
    style APP fill:#e8f5e9
    style WRK fill:#f3e5f5
```

Multi-stage сборка с distroless runtime:

```bash
# Сборка образа
make docker-build

# Запуск (PostgreSQL + миграции + app + worker)
make docker-up

# Остановка
make docker-down
```

### Development

Dev-overlay использует контейнер с полным Go SDK и монтированием исходников:

```bash
# Запуск dev-окружения
make docker-dev

# Остановка
make docker-dev-down
```

Особенности dev-режима:

- Hot-reload через `go run` (исходники монтируются из хоста)
- `APP_ENV=development`
- Тот же PostgreSQL, что и в production compose

### Управление

```bash
make docker-logs SVC=app     # Логи конкретного сервиса
make docker-logs              # Логи всех сервисов
make docker-ps                # Статус контейнеров
make docker-migrate           # Запуск миграций в Docker
```

### Настройка окружения

```bash
cp deployments/.env.example deployments/.env
```

Ключевые переменные:

| Переменная | Описание | По умолчанию |
|-----------|---------|-------------|
| `COMPOSE_PROJECT_NAME` | Имя проекта Docker Compose | `skeleton` |
| `GO_VERSION` | Версия Go для сборки | `1.25-alpine` |
| `POSTGRES_DB` | Имя базы данных | `skeleton` |
| `POSTGRES_USER` | Пользователь БД | `postgres` |
| `POSTGRES_PASSWORD` | Пароль БД | `postgres` |
| `APP_PORT` | Порт приложения на хосте | `8080` |
| `APP_ENV` | Окружение | `production` |

---

## Тестирование API

В директории `test/http/` находятся файлы для HTTP-клиентов JetBrains и VS Code REST Client:

```
test/http/
├── .env.http                      # Переменные окружения
├── http-client.env.json           # Окружения JetBrains
└── task/
    ├── create.http                # Создание задач (+ ошибки)
    ├── list.http                  # Список задач
    ├── get.http                   # Получение по ID (+ ошибки)
    ├── complete.http              # Завершение (+ ошибки)
    └── scenario.http              # Полный E2E сценарий с assertions
```

Файл `scenario.http` содержит полный end-to-end сценарий с проверками:

```http
### Создать задачу
POST {{BASE_URL}}/api/v1/tasks
Content-Type: application/json
{"title": "Первая задача", "description": "Описание"}

> {% client.assert(response.status === 201); %}
> {% client.global.set("TASK_ID", response.body.id); %}

### Завершить задачу
POST {{BASE_URL}}/api/v1/tasks/{{TASK_ID}}/complete

> {% client.assert(response.body.status === "done"); %}
```

---

## Makefile

| Target | Описание |
|--------|---------|
| **Локальная разработка** | |
| `make build` | Сборка бинарника в `bin/app` |
| `make run` | Запуск HTTP-сервера (`serve`) |
| `make run-worker` | Запуск воркера очередей (`queue:work`) |
| `make migrate` | Применить миграции |
| `make migrate-down` | Откатить последнюю миграцию |
| `make migrate-status` | Статус миграций |
| `make migrate-plan` | План pending-миграций |
| `make health` | Проверка здоровья |
| `make config` | Дамп конфигурации |
| `make test` | Запуск тестов (`-race -count=1`) |
| `make lint` | Линтер (`golangci-lint`) |
| `make fmt` | Форматирование кода (`gofmt` + `goimports`) |
| **Scaffolding** | |
| `make module` | Создать структуру нового модуля (интерактивный ввод) |
| **Docker (production)** | |
| `make docker-build` | Сборка Docker-образа |
| `make docker-up` | Запуск production-стека |
| `make docker-down` | Остановка production-стека |
| **Docker (development)** | |
| `make docker-dev` | Запуск dev-окружения |
| `make docker-dev-down` | Остановка dev-окружения |
| **Docker (утилиты)** | |
| `make docker-migrate` | Миграции в Docker |
| `make docker-logs SVC=app` | Логи сервиса |
| `make docker-ps` | Статус контейнеров |

---

## Создание нового модуля

### 1. Сгенерируйте структуру

```bash
make module
# Enter module name: order
# ✅ Module 'order' created at internal/module/order
```

Команда создаёт полное дерево директорий и заготовку `module.go`:

```
internal/module/order/
├── module.go                              # package order
├── domain/
│   ├── model/
│   ├── persistence/
│   └── business/
│       ├── emitter/
│       └── operation/
├── application/
│   ├── interactor/
│   ├── business/
│   │   ├── emitter/
│   │   └── operation/
│   └── port/
├── infrastructure/
│   ├── persistence/
│   ├── migration/
│   └── adapter/
└── presentation/
    ├── api/
    ├── job/
    └── listener/
```

> Если модуль с таким именем уже существует — команда завершится с ошибкой без перезаписи.

### 2. Определите доменную модель

```go
// internal/module/order/domain/model/order.go
type Order struct {
    id     OrderID
    total  Money
    status Status
    events []event.Event   // накопление доменных событий
}

func NewOrder(...) *Order {
    order := &Order{...}
    order.record(&event.OrderCreated{...})
    return order
}

func (o *Order) ReleaseEvents() []event.Event {
    evts := o.events
    o.events = nil
    return evts
}
```

### 3. Реализуйте фасад модуля

```go
// internal/module/order/module.go
package order

type Module struct { ... }

func NewModule(
    db *sql.DB,
    dispatcher *events.Dispatcher,
    log logger.Logger,             // internal/logger.Logger (structural typing)
) *Module { ... }

func (m *Module) Routes(router *httpserver.Router)           { ... }
func (m *Module) Listeners(d *events.Dispatcher)             { ... }
func (m *Module) Relays(relay *eventbus.OutboundRelay)       { ... }
func (m *Module) Consumers(qw *queueworker.Module, broker queue.Broker) { ... }
func (m *Module) Migrations(runner *migration.Runner)        { ... }
```

### 4. Зарегистрируйте модуль в bootstrap

```go
// internal/bootstrap/app.go — в функции Run
orderMod := order.NewModule(dbm.Default(), bus.Dispatcher(), log)

// В registerCommands — передайте модуль
orderMod.Listeners(bus.Dispatcher())
orderMod.Relays(relay)
orderMod.Consumers(qw, broker)
orderMod.Migrations(runner)
```

```go
// internal/bootstrap/router.go — добавьте маршруты
func buildRouter(
    log *logger.Logger,
    taskMod *task.Module,
    orderMod *order.Module,
) *httpserver.Router {
    router := httpserver.NewRouter()
    router.Use(...)

    taskMod.Routes(router)
    orderMod.Routes(router)

    return router
}
```

---

## Лицензия

[MIT](LICENSE) © 2025 Seytumerov Mustafa
