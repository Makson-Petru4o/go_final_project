package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func DeleteTask(id string) error {
	return execTaskChange(`DELETE FROM scheduler WHERE id = ?`, id)
}

func UpdateDate(next string, id string) error {
	return execTaskChange(`UPDATE scheduler SET date = ? WHERE id = ?`, next, id)
}

func execTaskChange(query string, args ...any) error {
	if db == nil {
		return errors.New("database is not initialized")
	}
	result, err := db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("change task: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("count changed tasks: %w", err)
	}
	if count == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func GetTask(id string) (*Task, error) {
	if db == nil {
		return nil, errors.New("database is not initialized")
	}
	var task Task
	var taskID int64
	err := db.QueryRow(`SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`, id).
		Scan(&taskID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		return nil, fmt.Errorf("get task: %w", err)
	}
	task.ID = strconv.FormatInt(taskID, 10)
	return &task, nil
}

func UpdateTask(task *Task) error {
	if db == nil {
		return errors.New("database is not initialized")
	}
	if task == nil {
		return errors.New("task is nil")
	}
	result, err := db.Exec(`UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`,
		task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return fmt.Errorf("update task: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("count updated tasks: %w", err)
	}
	if count == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func AddTask(task *Task) (id int64, err error) {
	if db == nil {
		return 0, errors.New("database is not initialized")
	}
	if task == nil {
		return 0, errors.New("task is nil")
	}
	tx, err := db.Begin()
	if err != nil {
		return 0, fmt.Errorf("begin task insert: %w", err)
	}
	defer func() {
		if err != nil {
			rollbackErr := tx.Rollback()
			if !errors.Is(rollbackErr, sql.ErrTxDone) {
				err = errors.Join(err, rollbackErr)
			}
		}
	}()
	result, err := tx.Exec(`INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`,
		task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, fmt.Errorf("insert task: %w", err)
	}
	id, err = result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get inserted task id: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit task insert: %w", err)
	}
	return id, nil
}
