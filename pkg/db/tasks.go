package db

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func Tasks(limit int) ([]*Task, error) {
	return queryTasks(limit, nil, `SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date ASC, id ASC LIMIT ?`, limit)
}

func SearchTasks(limit int, search string) ([]*Task, error) {
	if search == "" {
		return Tasks(limit)
	}
	if date, err := time.Parse("02.01.2006", search); err == nil {
		return queryTasks(limit, nil, `SELECT id, date, title, comment, repeat FROM scheduler WHERE date = ? ORDER BY date ASC, id ASC LIMIT ?`, date.Format("20060102"), limit)
	}
	search = strings.ToLower(search)
	return queryTasks(limit, func(task *Task) bool {
		return strings.Contains(strings.ToLower(task.Title), search) || strings.Contains(strings.ToLower(task.Comment), search)
	}, `SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date ASC, id ASC`)
}

func queryTasks(limit int, matches func(*Task) bool, query string, args ...any) ([]*Task, error) {
	if db == nil {
		return nil, errors.New("database is not initialized")
	}
	if limit < 1 {
		return nil, errors.New("task limit must be positive")
	}
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query tasks: %w", err)
	}
	defer rows.Close()
	tasks := make([]*Task, 0)
	for rows.Next() {
		var task Task
		var id int64
		if err := rows.Scan(&id, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		task.ID = strconv.FormatInt(id, 10)
		if matches != nil && !matches(&task) {
			continue
		}
		tasks = append(tasks, &task)
		if len(tasks) == limit {
			break
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read tasks: %w", err)
	}
	return tasks, nil
}
