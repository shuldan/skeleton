package model

// TaskID — идентификатор задачи (строковое значение).
// Генерация и парсинг UUID происходят вне пакета model.
type TaskID struct {
	value string
}

// NewTaskID создаёт TaskID из строки. Валидация формата — на вызывающей стороне.
func NewTaskID(raw string) TaskID {
	return TaskID{value: raw}
}

// String возвращает строковое представление (единственный экспортируемый метод ID).
func (id TaskID) String() string {
	return id.value
}
