package persistence

import (
	"context"
	"database/sql"
	"errors"

	"github.com/shuldan/repository"

	"github.com/shuldan/skeleton/internal/module/task/domain/model"
	domainpersistence "github.com/shuldan/skeleton/internal/module/task/domain/persistence"
)

var taskTable = repository.Table{
	Name:          "tasks",
	PrimaryKey:    "id",
	Columns:       []string{"id", "title", "description", "status", "version"},
	VersionColumn: "version",
	CreatedAt:     "created_at",
	UpdatedAt:     "updated_at",
}

type taskRepository struct {
	repo *repository.Repository[*model.Task]
}

// NewTaskRepository создаёт SQL-реализацию репозитория.
func NewTaskRepository(
	db *sql.DB,
) domainpersistence.TaskRepository {
	r := repository.New(
		db,
		repository.Postgres(),
		repository.Simple(repository.SimpleConfig[*model.Task]{
			Table:  taskTable,
			Scan:   scanTask,
			Values: taskValues,
		}),
	)

	return &taskRepository{repo: r}
}

func (r *taskRepository) FindByID(
	ctx context.Context, id model.TaskID,
) (*model.Task, error) {
	task, err := r.repo.Find(ctx, id.String())
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, model.ErrTaskNotFound.
				WithDetail("id", id.String())
		}

		return nil, err
	}

	return task, nil
}

func (r *taskRepository) FindAll(
	ctx context.Context,
) ([]*model.Task, error) {
	return r.repo.FindBy(ctx, nil)
}

func (r *taskRepository) Save(
	ctx context.Context, task *model.Task,
) error {
	err := r.repo.Save(ctx, task)
	if err != nil {
		if errors.Is(err, repository.ErrConcurrentModification) {
			return model.ErrConcurrentModification.
				WithDetail("id", task.ID().String()).
				WithCause(err)
		}

		return err
	}

	return nil
}

func scanTask(sc repository.Scanner) (*model.Task, error) {
	var s model.TaskSnapshot
	if err := sc.Scan(
		&s.ID, &s.Title, &s.Description,
		&s.Status, &s.Version,
	); err != nil {
		return nil, err
	}

	return s.Restore()
}

func taskValues(t *model.Task) []any {
	s := t.Snapshot()

	return []any{
		s.ID, s.Title, s.Description,
		s.Status, s.Version,
	}
}
