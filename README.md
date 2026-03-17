shuldan/skeleton/README.md
# Skeleton — Go Application Template

[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

**Skeleton** — `gonew`-совместимый шаблон Go-приложения с модульной DDD-архитектурой, встроенной системой событий, командной шиной, очередями, миграциями и Docker-окружением. Содержит два модуля (`task` и `payment`) в качестве примера — удалите или замените их на свой домен.

---

## Содержание

- [Быстрый старт](#быстрый-старт)
- [Архитектура](#архитектура)
  - [Высокоуровневая схема](#высокоуровневая-схема)
  - [Поток запроса](#поток-запроса)
  - [Система событий](#система-событий)
  - [Командная шина](#командная-шина)
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
  - [Полный цикл события в task-модуле](#полный-цикл-события-в-task-модуле)
- [Командная шина (Command Bus)](#командная-шина-command-bus)
  - [Обзор](#обзор)
  - [Определение команд и результатов](#определение-команд-и-результатов)
  - [Отправка команд (CommandClient)](#отправка-команд-commandclient)
  - [Обработка команд (CommandServer)](#обработка-команд-commandserver)
  - [Получение результатов (Future)](#получение-результатов-future)
  - [Регистрация в bootstrap](#регистрация-в-bootstrap)
  - [Полный цикл команды](#полный-цикл-команды)
  - [Когда использовать Command Bus vs Events](#когда-использовать-command-bus-vs-events)
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

# Завершить задачу (→ автоматически создаст invoice через Command Bus)
curl -X POST http://localhost:8080/api/v1/tasks/{id}/complete

# Проверить созданные счета
curl http://localhost:8080/api/v1/invoices
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
    BS --> M2["module/payment<br/><i>пример</i>"]
    BS --> M3["module/...<br/><i>ваш код</i>"]

    subgraph module["Структура модуля"]
        direction TB
        PR["presentation<br/>HTTP, Listeners, Jobs,<br/>Command Handlers"]
        AP["application<br/>Interactors, Operations,<br/>Emitters, Ports"]
        DM["domain<br/>Models, Value Objects,<br/>Interfaces"]
        IN["infrastructure<br/>Repositories, Adapters,<br/>Migrations"]

        PR --> AP
        AP --> DM
        IN --> DM
    end

    M1 -.-> module

    subgraph communication["Межмодульное взаимодействие"]
        EVT["Events<br/><i>уведомления</i>"]
        CMD["Command Bus<br/><i>запрос-ответ</i>"]
    end

    M1 --> EVT
    M1 --> CMD
    CMD --> M2
    EVT --> M2

    style CLI fill:#e1f5fe
    style BS fill:#e1f5fe
    style DM fill:#fff3e0
    style AP fill:#e8f5e9
    style PR fill:#f3e5f5
    style IN fill:#fce4ec
    style communication fill:#f5f5f5
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
    I->>E: emitter.Emit(ctx, []any)
    E->>D: dispatcher.Publish(ctx, event) × N
    I-->>H: nil / error
    H->>C: JSON Response
```

### Система событий

```mermaid
graph LR
    AGG["Aggregate<br/><i>record(event)</i>"] --> INT["Interactor<br/><i>ReleaseEvents()</i>"]
    INT --> EM["Emitter<br/><i>[]any</i>"]
    EM --> DISP["Dispatcher"]
    
    DISP --> L1["Listener<br/><i>in-process</i>"]
    DISP --> CMDSEND["CommandSender<br/><i>→ Command Bus</i>"]
    DISP --> JOB["Job Consumer<br/><i>async worker</i>"]

    style AGG fill:#fff3e0
    style INT fill:#e8f5e9
    style DISP fill:#fff3e0
    style L1 fill:#f3e5f5
    style CMDSEND fill:#e8eaf6
    style JOB fill:#fce4ec
```

**Поток событий:**

1. Агрегат **накапливает** события через `record()` при изменении состояния
2. Interactor вызывает `task.ReleaseEvents()` — забирает и очищает очередь
3. Emitter итерирует по слайсу и публикует каждое событие в Dispatcher
4. Dispatcher доставляет события подписчикам:
   - **Listener** (in-process обработка)
   - **CommandSender** (→ Command Bus для межмодульных команд)

### Командная шина

```mermaid
graph LR
    subgraph "Модуль-отправитель (task)"
        EVT["TaskCompleted<br/><i>событие</i>"]
        CL["CommandClient"]
    end
    
    EVT --> CL
    CL -->|"CommandEnvelope"| TR["Transport<br/><i>memory / NATS / ...</i>"]
    
    subgraph "Модуль-получатель (payment)"
        SRV["CommandServer"]
        CH["Handler[C]"]
        INTER["Interactor"]
    end
    
    TR --> SRV
    SRV --> CH
    CH --> INTER
    INTER -->|"Result via ReplySender"| CH
    CH -->|"ReplyEnvelope"| TR
    
    subgraph "Модуль-отправитель (task)"
        FUT["Future / TypedFuture"]
    end
    
    TR --> FUT

    style EVT fill:#fff3e0
    style CL fill:#e8eaf6
    style TR fill:#e1f5fe
    style SRV fill:#e8eaf6
    style CH fill:#f3e5f5
    style FUT fill:#e8eaf6
```

### Принципы

| Принцип | Реализация |
|---------|-----------|
| **DDD** | Агрегаты, value objects, domain events (накопление в агрегате), репозитории |
| **Clean Architecture** | Зависимости направлены внутрь: presentation → application → domain ← infrastructure |
| **Модульность** | Каждый домен — изолированный модуль с фасадом `module.go` |
| **Presenter-паттерн** | Домен не знает о JSON/HTTP; данные передаются через `TaskPresenter` |
| **Snapshot-паттерн** | Персистентность через плоские снимки агрегата, без экспорта приватных полей |
| **CQRS-lite** | Разделение операций записи (operations/interactors) и чтения |
| **Event-Driven** | Агрегат накапливает события → emitter публикует пакетом |
| **Command Bus** | Межмодульное взаимодействие через `CommandClient` / `CommandServer` с `Future` |
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
│   ├── command/
│   │   └── invoice.go               # Определения команд и результатов (Command Bus)
│   │
│   ├── event/
│   │   ├── event.go                 # Пакет доменных событий
│   │   └── task.go                  # Структуры событий задачи
│   │
│   ├── logger/
│   │   └── logger.go                # Интерфейс Logger (structural typing)
│   │
│   └── module/
│       ├── task/                     # ── Модуль Task (отправитель команд) ──
│       │   ├── module.go            # Фасад: Routes, Listeners, CommandSenders, etc.
│       │   ├── domain/
│       │   ├── application/
│       │   ├── infrastructure/
│       │   └── presentation/
│       │       ├── api/             # HTTP handlers
│       │       ├── listener/        # Event listeners + CommandClient sender
│       │       └── job/             # Queue consumers
│       │
│       └── payment/                  # ── Модуль Payment (получатель команд) ──
│           ├── module.go            # Фасад: Routes, CommandHandlers, Migrations
│           ├── domain/
│           ├── application/
│           ├── infrastructure/
│           └── presentation/
│               ├── api/             # HTTP handlers
│               └── commandhandler/  # Command Bus handlers
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
            LST["listener/<br/>event handlers +<br/>command senders"]
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
    events      []any   // очередь доменных событий
}

// Конструктор — создаёт задачу и записывает событие TaskCreated
func NewTask(id TaskID, title Title, description string) *Task {
    task := &Task{
        id: id, title: title, description: description,
        status: statusDraft, version: 1,
    }
    task.record(&event.TaskCreated{
        TaskID: task.id.String(),
        Title:  title.String(),
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
    t.record(&event.TaskCompleted{TaskID: t.id.String()})
    return nil
}

// ReleaseEvents — забирает накопленные события и очищает очередь
func (t *Task) ReleaseEvents() []any {
    evts := t.events
    t.events = nil
    return evts
}

func (t *Task) record(e any) {
    t.events = append(t.events, e)
}
```

> **Паттерн:** события не публикуются немедленно. Агрегат накапливает их через `record()`, а interactor забирает пакетом через `ReleaseEvents()` после успешного сохранения. Это гарантирует, что события публикуются только при успешной операции.

> **`events.Event` — это `any`:** библиотека `shuldan/events` определяет `Event` как `any`, поэтому доменные события — простые структуры без обязательных методов. Агрегат хранит `[]any`.

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

    i.emitter.Emit(ctx, task.ReleaseEvents())  // публикация пакета событий
    task.RepresentTo(output)                    // заполнение output через presenter
    return nil
}
```

> **Порядок:** сначала `ReleaseEvents` + `Emit` (публикация событий), затем `RepresentTo` (данные для ответа). События публикуются только после успешного сохранения в Operation.

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

**Emitter** — единый для всех событий агрегата. Принимает слайс `[]any` и публикует каждый элемент в Dispatcher:

```go
func (e *TaskEmitter) Emit(ctx context.Context, domainEvents []any) {
    publishCtx := context.WithoutCancel(ctx)  // события отправляются даже при отмене запроса

    for _, ev := range domainEvents {
        if err := e.dispatcher.Publish(publishCtx, ev); err != nil {
            e.log.Error("failed to emit event", ...)
        }
    }
}
```

> **`context.WithoutCancel`**: emitter публикует события в контексте, не привязанном к HTTP-запросу. Если клиент отключился, события всё равно доставляются подписчикам.

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

Адаптеры ввода: HTTP, event listeners, command handlers, queue consumers.

```
presentation/
├── api/
│   ├── create_task_handler.go     # POST /api/v1/tasks
│   ├── list_tasks_handler.go      # GET  /api/v1/tasks
│   ├── get_task_handler.go        # GET  /api/v1/tasks/{id}
│   ├── complete_task_handler.go   # POST /api/v1/tasks/{id}/complete
│   └── errors.go                  # API-специфичные ошибки
├── listener/
│   ├── task_completed_listener.go         # In-process обработка TaskCompleted
│   └── task_completed_command_sender.go   # Отправка CreateInvoice через CommandClient
├── commandhandler/                        # (в модуле payment)
│   └── create_invoice_handler.go          # Обработка команды CreateInvoice
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
func (m *Module) CommandSenders(client *commands.CommandClient)
func (m *Module) Consumers(qw *queueworker.Module, broker queue.Broker)
func (m *Module) Migrations(runner *migration.Runner)
```

Для модулей-получателей команд (payment):

```go
func NewModule(db *sql.DB, log logger.Logger) *Module
func (m *Module) Routes(router *httpserver.Router)
func (m *Module) CommandHandlers(server *commands.CommandServer)
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

Система событий состоит из двух уровней:

```mermaid
graph LR
    subgraph "Внутри агрегата"
        AGG["Aggregate<br/>record()"] --> REL["ReleaseEvents()"]
    end

    subgraph "In-process (Dispatcher)"
        REL --> EM["Emitter<br/>[]any"]
        EM --> DISP["Dispatcher"]
        DISP --> LST["Listener"]
        DISP --> CMDSND["CommandSender"]
    end

    style AGG fill:#fff3e0
    style EM fill:#e8f5e9
    style DISP fill:#fff3e0
    style LST fill:#f3e5f5
    style CMDSND fill:#e8eaf6
```

### Определение событий

События — простые структуры в `internal/event/`. Библиотека `shuldan/events` определяет `Event` как `any`, поэтому никаких обязательных методов реализовывать не нужно:

```go
// internal/event/task.go

// TaskCreated публикуется при создании задачи.
type TaskCreated struct {
    TaskID string `json:"task_id"`
    Title  string `json:"title"`
}

// EventKey реализует events.KeyedEvent для упорядоченной обработки.
func (e *TaskCreated) EventKey() string { return e.TaskID }

// TaskCompleted публикуется при завершении задачи.
type TaskCompleted struct {
    TaskID string `json:"task_id"`
}

func (e *TaskCompleted) EventKey() string { return e.TaskID }
```

> **`EventKey()`** — опциональный метод. Если событие реализует `events.KeyedEvent`, Dispatcher гарантирует, что события с одинаковым ключом обрабатываются последовательно (ordering).

### Накопление событий в агрегате

Агрегат **не публикует** события немедленно. Он записывает их во внутреннюю очередь:

```go
// Конструктор — записывает TaskCreated
func NewTask(id TaskID, title Title, description string) *Task {
    task := &Task{...}
    task.record(&event.TaskCreated{
        TaskID: task.id.String(),
        Title:  title.String(),
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
    t.record(&event.TaskCompleted{TaskID: t.id.String()})
    return nil
}

// ReleaseEvents — interactor забирает события после успешного сохранения
func (t *Task) ReleaseEvents() []any {
    evts := t.events
    t.events = nil
    return evts
}
```

> **Гарантия:** если `repo.Save()` вернул ошибку, interactor не вызывает `ReleaseEvents()` — события не публикуются. Если `Restore()` восстанавливает агрегат из БД — события не записываются.

### Emitter (публикация)

Единый emitter принимает **слайс `[]any`** и публикует каждый элемент в Dispatcher:

```go
func (e *TaskEmitter) Emit(ctx context.Context, domainEvents []any) {
    publishCtx := context.WithoutCancel(ctx)

    for _, ev := range domainEvents {
        if err := e.dispatcher.Publish(publishCtx, ev); err != nil {
            e.log.Error("failed to emit event", ...)
        }
    }
}
```

**Ключевые решения:**

| Решение | Причина |
|---------|---------|
| `context.WithoutCancel(ctx)` | Клиент отключился — события всё равно доставляются |
| Один emitter на агрегат | Проще, чем per-event emitter. Новое событие = 0 новых файлов |
| `[]any` в интерфейсе | Совпадает с `events.Event = any` — нет каста |

**Доменный интерфейс:**

```go
// domain/business/emitter/event_emitter.go
type EventEmitter interface {
    Emit(ctx context.Context, events []any)
}
```

### Listener (подписка)

In-process подписка через типизированный `events.Subscribe`:

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

// module.go — регистрация
func (m *Module) Listeners(d *events.Dispatcher) {
    events.Subscribe(d, listener.NewTaskCompletedListener(m.log))
}
```

> **Типизация:** `events.Subscribe` использует дженерики — обработчик получает конкретный тип `*event.TaskCompleted`, а не `any`. Маршрутизация по типу событий происходит автоматически через `reflect.TypeFor[E]()`.

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
    participant CS as CommandSender (listener)
    participant CL as CommandClient
    participant TR as Transport
    participant SRV as CommandServer
    participant PH as Payment Handler

    H->>I: Handle(ctx, input, output)
    I->>O: Complete(ctx, taskID)
    O->>T: task.Complete() → record(TaskCompleted)
    O->>R: Save(task)
    R-->>O: ok
    O-->>I: task
    I->>T: ReleaseEvents() → [TaskCompleted]
    I->>E: Emit(ctx, [TaskCompleted])
    E->>D: Publish(TaskCompleted)
    
    par In-process
        D->>L: Handle(TaskCompleted)
    and Command Bus
        D->>CS: Handle(TaskCompleted)
        CS->>CL: commands.Send[*InvoiceCreated](ctx, client, cmd)
        CL->>TR: Send(CommandEnvelope)
        TR->>SRV: dispatch → Handler.Handle(ctx, cmd, reply)
        SRV->>PH: Handle(ctx, CreateInvoice, reply)
        PH->>PH: reply.Send(ctx, InvoiceCreated)
        PH-->>TR: ReplyEnvelope
        TR-->>CL: handleReply → future.complete()
        CL-->>CS: future.Await() → InvoiceCreated
    end
    
    I->>T: RepresentTo(output)
    I-->>H: nil
    H->>Client: JSON Response
```

---

## Командная шина (Command Bus)

### Обзор

Командная шина обеспечивает **межмодульное взаимодействие** с семантикой **запрос-ответ** (request/reply). Построена на абстракциях `Transport` и `Codec` из пакета `shuldan/commands`:

- **`CommandClient`** — отправляет команды и получает `Future` с результатом
- **`CommandServer`** — принимает команды и вызывает типизированные обработчики
- **`Transport`** — абстракция транспорта (in-memory, NATS, RabbitMQ и т.д.)
- **`Codec`** — сериализация/десериализация (JSON, Protobuf и т.д.)

```mermaid
graph TB
    subgraph "Отправитель (task module)"
        EVT["TaskCompleted<br/><i>событие</i>"]
        LST["Listener"]
        CL["CommandClient<br/><i>Send → Future</i>"]
    end

    subgraph "Транспорт"
        TR["Transport<br/><i>memory / NATS / ...</i>"]
    end

    subgraph "Получатель (payment module)"
        SRV["CommandServer"]
        HDL["Handler[*CreateInvoice]"]
        INTER["Interactor"]
        RS["ReplySender"]
    end

    EVT -->|"Subscribe"| LST
    LST -->|"commands.Send"| CL
    CL -->|"CommandEnvelope"| TR
    TR -->|"dispatch"| SRV
    SRV --> HDL
    HDL --> INTER
    INTER --> HDL
    HDL -->|"reply.Send(result)"| RS
    RS -->|"ReplyEnvelope"| TR
    TR -->|"future.complete"| CL

    style EVT fill:#fff3e0
    style CL fill:#e8eaf6
    style TR fill:#e1f5fe
    style SRV fill:#e8eaf6
    style HDL fill:#f3e5f5
    style RS fill:#e8eaf6
```

Skeleton демонстрирует полный цикл:
- **Task** завершается → событие `TaskCompleted`
- Listener отправляет команду `CreateInvoice` через `CommandClient`
- **Payment** модуль получает команду, создаёт счёт, отправляет результат через `ReplySender`
- **Task** модуль получает результат `InvoiceCreated` через `TypedFuture`

### Определение команд и результатов

Команды и результаты определяются в `internal/command/` — общем пространстве, доступном обоим модулям:

```go
// internal/command/invoice.go

// CreateInvoice — команда создания счёта.
type CreateInvoice struct {
    TaskID string `json:"task_id"`
    Amount int    `json:"amount"`
}

func (c *CreateInvoice) CommandName() string { return "CreateInvoice" }

// InvoiceCreated — результат успешного создания.
type InvoiceCreated struct {
    InvoiceID string `json:"invoice_id"`
    TaskID    string `json:"task_id"`
    Amount    int    `json:"amount"`
    Status    string `json:"status"`
}

func (r *InvoiceCreated) ResultName() string { return "InvoiceCreated" }
```

**Интерфейсы из `shuldan/commands`:**

| Интерфейс | Методы | Назначение |
|-----------|--------|-----------|
| `commands.Command` | `CommandName() string` | Идентификация команды для маршрутизации |
| `commands.Result` | `ResultName() string` | Идентификация результата для десериализации |

### Отправка команд (CommandClient)

Команды отправляются через `CommandClient`. Типизированная функция `commands.Send[R]` возвращает `TypedFuture[R]`:

```go
// presentation/listener/task_completed_command_sender.go

type TaskCompletedCommandSender struct {
    client *commands.CommandClient
    log    logger.Logger
}

func (l *TaskCompletedCommandSender) Handle(
    ctx context.Context, e *event.TaskCompleted,
) error {
    cmd := &appcommand.CreateInvoice{
        TaskID: e.TaskID,
        Amount: 100,
    }

    // Типизированная отправка — возвращает TypedFuture[*InvoiceCreated]
    future, err := commands.Send[*appcommand.InvoiceCreated](
        ctx, l.client, cmd,
    )
    if err != nil {
        l.log.Error("failed to send CreateInvoice command", ...)
        return err
    }

    // Обработка результата асинхронно
    go func() {
        result, awaitErr := future.Await(context.Background())
        if awaitErr != nil {
            l.log.Error("CreateInvoice command failed", ...)
            return
        }

        l.log.Info("invoice created via command bus",
            "invoice_id", result.InvoiceID,
            "task_id",    result.TaskID,
        )
    }()

    return nil
}
```

**Регистрация в модуле:**

```go
// module.go (task)
func (m *Module) CommandSenders(
    d *events.Dispatcher, client *commands.CommandClient,
) {
    l := listener.NewTaskCompletedCommandSender(client, m.log)
    events.Subscribe(d, l)
}
```

> **`commands.Send[R]` vs `client.Send`**: типизированная функция `commands.Send[*InvoiceCreated]` автоматически десериализует результат в нужный тип через `Codec`. Нетипизированный `client.Send` возвращает `Future` с `Result` интерфейсом.

> **`Future.Await`** блокируется до получения ответа или таймаута. По умолчанию таймаут 30 секунд (настраивается через `commands.WithTimeout`).

### Обработка команд (CommandServer)

Обработчик реализует `commands.Handler[C]` и регистрируется на `CommandServer`:

```go
// presentation/commandhandler/create_invoice_handler.go (payment module)

type CreateInvoiceHandler struct {
    inter *interactor.CreateInvoiceInteractor
}

func (h *CreateInvoiceHandler) Handle(
    ctx context.Context,
    cmd *appcommand.CreateInvoice,    // типизированная команда
    reply commands.ReplySender,       // для отправки ответа
) error {
    output := &createInvoiceOutput{}
    input := &createInvoiceInput{cmd: cmd}

    if err := h.inter.Handle(ctx, input, output); err != nil {
        return reply.SendError(ctx, err)  // ошибка → ErrorPayload
    }

    return reply.Send(ctx, &appcommand.InvoiceCreated{
        InvoiceID: output.ID,
        TaskID:    output.TaskID,
        Amount:    output.Amount,
        Status:    output.Status,
    })
}
```

**Регистрация через дженерик `commands.Register`:**

```go
// module.go (payment)
func (m *Module) CommandHandlers(server *commands.CommandServer) {
    handler := commandhandler.NewCreateInvoiceHandler(m.createInteractor)

    if err := commands.Register[*command.CreateInvoice](server, handler); err != nil {
        m.log.Error("failed to register command handler", ...)
    }
}
```

> **`commands.Register[C]`** автоматически десериализует payload в тип `C` через `Codec` и маршрутизирует по `CommandName()`. Один обработчик на одну команду.

> **`ReplySender`** — отправляет результат обратно клиенту. `Send` для успешного результата, `SendError` для ошибки. Ответ автоматически доставляется через `Transport` к `Future` на стороне клиента.

### Получение результатов (Future)

`Future` / `TypedFuture` — механизм получения результата команды:

```go
// TypedFuture[R] — типобезопасный future
type TypedFuture[R Result] interface {
    Await(ctx context.Context) (R, error)     // блокируется до результата
    Done() <-chan struct{}                     // канал завершения
    Result() (R, error, bool)                 // неблокирующая проверка
}
```

**Варианты использования:**

```go
// 1. Блокирующее ожидание
result, err := future.Await(ctx)

// 2. Ожидание с таймаутом
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()
result, err := future.Await(ctx)

// 3. Асинхронная обработка
go func() {
    result, err := future.Await(context.Background())
    // ...
}()

// 4. Неблокирующая проверка
if result, err, ok := future.Result(); ok {
    // результат готов
}

// 5. Select
select {
case <-future.Done():
    result, err, _ := future.Result()
    // ...
case <-ctx.Done():
    // таймаут
}
```

> **Таймаут:** если ответ не получен за `defaultTimeout` (30 сек), future завершается с `commands.ErrTimeout`. Настраивается через `commands.WithTimeout()` при создании клиента или `commands.WithSendTimeout()` при отправке.

### Регистрация в bootstrap

Все компоненты Command Bus собираются в `bootstrap/app.go`:

```go
func Run(ctx context.Context) error {
    // ...

    // 1. Transport — абстракция доставки (in-memory для одного процесса)
    transport := memorytransport.New()

    // 2. Codec — сериализация (JSON)
    codec := codecjson.New()

    // 3. CommandClient — отправитель команд
    cmdClient, err := commands.NewCommandClient(transport, codec,
        commands.WithTimeout(10*time.Second),
    )

    // 4. CommandServer — получатель команд
    cmdServer, err := commands.NewCommandServer(transport, codec)

    // 5. Регистрация обработчиков ДО открытия сервера
    paymentMod.CommandHandlers(cmdServer)

    // 6. Модуль жизненного цикла (Open/Close)
    cmdModule := commandbus.NewModule(
        commandbus.WithClient(cmdClient),
        commandbus.WithServer(cmdServer),
    )

    // 7. Регистрация event listener для отправки команд
    taskMod.CommandSenders(bus.Dispatcher(), cmdClient)

    // ...
}
```

**Порядок инициализации:**

1. Создать `Transport` и `Codec`
2. Создать `CommandClient` и `CommandServer`
3. Зарегистрировать обработчики на сервере (`commands.Register`)
4. Обернуть в `commandbus.Module` (управление жизненным циклом)
5. `Module.Start` вызовет `server.Open` и `client.Open`
6. `Module.Stop` вызовет `client.Close` и `server.Close`

### Полный цикл команды

```mermaid
sequenceDiagram
    participant T as Task Module
    participant D as Dispatcher
    participant LS as Listener (sender)
    participant CL as CommandClient
    participant TR as Transport
    participant SRV as CommandServer
    participant HDL as Handler[CreateInvoice]
    participant RS as ReplySender
    participant FUT as TypedFuture

    Note over T: task.Complete() → TaskCompleted
    T->>D: Publish(TaskCompleted)
    D->>LS: Handle(TaskCompleted)
    LS->>CL: commands.Send[*InvoiceCreated](ctx, client, cmd)
    CL->>CL: codec.Encode(cmd) → payload
    CL->>CL: Store future by correlationID
    CL->>TR: Send(CommandEnvelope)
    CL-->>LS: TypedFuture[*InvoiceCreated]
    
    Note over TR: Доставка через Transport
    
    TR->>SRV: dispatch(CommandEnvelope)
    SRV->>SRV: codec.Decode(payload) → *CreateInvoice
    SRV->>HDL: Handle(ctx, *CreateInvoice, reply)
    HDL->>HDL: interactor.Handle(ctx, input, output)
    HDL->>RS: reply.Send(ctx, InvoiceCreated)
    RS->>RS: codec.Encode(result) → payload
    RS->>TR: Reply(ReplyEnvelope)
    
    TR->>CL: handleReply(ReplyEnvelope)
    CL->>CL: Lookup future by correlationID
    CL->>CL: codec.Decode(payload) → *InvoiceCreated
    CL->>FUT: complete(InvoiceCreated)
    
    Note over LS: Асинхронно
    LS->>FUT: future.Await(ctx) → *InvoiceCreated, nil
```

### Когда использовать Command Bus vs Events

| Характеристика | Events | Command Bus |
|---------------|--------|-------------|
| **Семантика** | «Что-то произошло» (уведомление) | «Сделай это» (запрос) |
| **Получатели** | 0..N подписчиков | Ровно 1 обработчик |
| **Ответ** | Нет (fire-and-forget) | Есть (`Future` с `Result` или ошибкой) |
| **Связанность** | Слабая (отправитель не знает о подписчиках) | Средняя (отправитель знает имя команды и тип результата) |
| **Таймаут** | Нет | Встроенный (настраиваемый) |
| **Типобезопасность** | Дженерики на подписке | Дженерики на отправке и обработке |
| **Примеры** | Логирование, метрики, уведомления | Создание связанных сущностей, платежи |

**В skeleton оба механизма работают вместе:**

```mermaid
graph LR
    TC["TaskCompleted<br/><i>событие</i>"]
    
    TC --> L1["Listener<br/><i>логирование</i>"]
    TC --> CS["CommandSender<br/><i>CreateInvoice → Future</i>"]
    
    style TC fill:#fff3e0
    style L1 fill:#f3e5f5
    style CS fill:#e8eaf6
```

Одно событие `TaskCompleted` порождает:
1. **Listener** — синхронное логирование (in-process)
2. **CommandSender** — создание счёта в payment-модуле (Command Bus с `Future`)

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
// task module
func (m *Module) Routes(router *httpserver.Router) {
    group := router.Group("/api/v1/tasks")

    group.POST("", api.NewCreateTaskHandler(m.createInteractor))
    group.GET("", api.NewListTasksHandler(m.listInteractor))
    group.GET("/{id}", api.NewGetTaskHandler(m.getInteractor))
    group.POST("/{id}/complete", api.NewCompleteTaskHandler(m.completeInteractor))
}

// payment module
func (m *Module) Routes(router *httpserver.Router) {
    group := router.Group("/api/v1/invoices")

    group.GET("", api.NewListInvoicesHandler(m.listInteractor))
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

> **Префиксы ошибок**: доменные ошибки используют префикс `TASK` (`taskCode`), API-специфичные — `TASK_API` (`apiCode`), payment — `INVOICE` (`invoiceCode`). Это позволяет различать источник ошибки в логах и ответах.

---

## Presenter-паттерн

Домен не знает о формате вывода. Данные передаются через интерфейс `TaskPresenter`:

```mermaid
graph TB
    AGG["Task aggregate<br/><i>приватные поля</i>"] -->|RepresentTo| PI["TaskPresenter<br/><i>интерфейс</i>"]

    PI --> HTTP["CreateTaskOutput<br/><i>JSON struct</i>"]
    PI --> CMD["createInvoiceOutput<br/><i>Command handler</i>"]
    PI --> TEST["mockPresenter<br/><i>для тестов</i>"]

    style AGG fill:#fff3e0
    style PI fill:#e1f5fe
    style HTTP fill:#f3e5f5
    style CMD fill:#e8eaf6
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

// presentation/commandhandler/ — для результата команды
type createInvoiceOutput struct {
    ID     string
    TaskID string
    // ...
}
func (o *createInvoiceOutput) SetID(v string) model.InvoicePresenter { o.ID = v; return o }
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
├── task/
│   ├── create.http                # Создание задач (+ ошибки)
│   ├── list.http                  # Список задач
│   ├── get.http                   # Получение по ID (+ ошибки)
│   ├── complete.http              # Завершение (+ ошибки)
│   └── scenario.http              # Полный E2E сценарий с assertions
└── invoice/
    └── list.http                  # Список счетов
```

Файл `scenario.http` содержит полный end-to-end сценарий с проверками:

```http
### Создать задачу
POST {{BASE_URL}}/api/v1/tasks
Content-Type: application/json
{"title": "Первая задача", "description": "Описание"}

> {% client.assert(response.status === 201); %}
> {% client.global.set("TASK_ID", response.body.id); %}

### Завершить задачу (→ автоматически создаст invoice через Command Bus)
POST {{BASE_URL}}/api/v1/tasks/{{TASK_ID}}/complete

> {% client.assert(response.body.status === "done"); %}

### Проверить, что invoice создан
GET {{BASE_URL}}/api/v1/invoices

> {% client.assert(response.body.invoices.length > 0); %}
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
| `make migrate-down` | Откатить последнюю |
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
    events []any   // накопление доменных событий
}

func NewOrder(...) *Order {
    order := &Order{...}
    order.record(&event.OrderCreated{...})
    return order
}

func (o *Order) ReleaseEvents() []any {
    evts := o.events
    o.events = nil
    return evts
}
```

### 3. Реализуйте фасад модуля

Для модулей с событиями и командами:

```go
// internal/module/order/module.go
package order

type Module struct { ... }

func NewModule(
    db *sql.DB,
    dispatcher *events.Dispatcher,
    log logger.Logger,
) *Module { ... }

func (m *Module) Routes(router *httpserver.Router)
func (m *Module) Listeners(d *events.Dispatcher)
func (m *Module) CommandSenders(d *events.Dispatcher, client *commands.CommandClient)
func (m *Module) Consumers(qw *queueworker.Module, broker queue.Broker)
func (m *Module) Migrations(runner *migration.Runner)
```

Для модулей-получателей команд:

```go
func NewModule(db *sql.DB, log logger.Logger) *Module
func (m *Module) Routes(router *httpserver.Router)
func (m *Module) CommandHandlers(server *commands.CommandServer)
func (m *Module) Migrations(runner *migration.Runner)
```

### 4. Зарегистрируйте модуль в bootstrap

```go
// internal/bootstrap/app.go — в функции Run
orderMod := order.NewModule(dbm.Default(), bus.Dispatcher(), log)

// В registerCommands
orderMod.Listeners(bus.Dispatcher())
orderMod.CommandSenders(bus.Dispatcher(), cmdClient)
orderMod.Consumers(qw, broker)
orderMod.Migrations(runner)
```

```go
// internal/bootstrap/router.go — добавьте маршруты
func buildRouter(
    log *logger.Logger,
    taskMod *task.Module,
    paymentMod *payment.Module,
    orderMod *order.Module,
) *httpserver.Router {
    router := httpserver.NewRouter()
    router.Use(...)

    taskMod.Routes(router)
    paymentMod.Routes(router)
    orderMod.Routes(router)

    return router
}
```

### 5. Определите команды (если нужно межмодульное взаимодействие)

```go
// internal/command/shipment.go
type CreateShipment struct {
    OrderID string `json:"order_id"`
    Address string `json:"address"`
}

func (c *CreateShipment) CommandName() string { return "CreateShipment" }

type ShipmentCreated struct {
    ShipmentID string `json:"shipment_id"`
    OrderID    string `json:"order_id"`
}

func (r *ShipmentCreated) ResultName() string { return "ShipmentCreated" }
```

---

## Лицензия

[MIT](LICENSE) © 2025 Seytumerov Mustafa
