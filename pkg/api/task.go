package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/Yandex-Practicum/go_final_project/pkg/db"
)

func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getTaskHandler(w, r)
	case http.MethodPost:
		addTaskHandler(w, r)
	case http.MethodPut:
		editTaskHandler(w, r)
	case http.MethodDelete:
		deleteTaskHandler(w, r)
	default:
		w.Header().Set("Allow", "GET, POST, PUT, DELETE")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func checkID(id string) error {
	if id == "" {
		return errors.New("Не указан идентификатор")
	}
	for _, digit := range id {
		if digit < '0' || digit > '9' {
			return errors.New("Некорректный идентификатор")
		}
	}
	value, err := strconv.ParseInt(id, 10, 64)
	if err != nil || value <= 0 {
		return errors.New("Некорректный идентификатор")
	}
	return nil
}

func decodeTask(r *http.Request) (*db.Task, error) {
	var task db.Task
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&task); err != nil {
		return nil, errors.New("Некорректный JSON")
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return nil, errors.New("Ожидается один JSON-объект")
	}
	return &task, nil
}

func writeTaskError(w http.ResponseWriter, err error) {
	if errors.Is(err, sql.ErrNoRows) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Задача не найдена"})
		return
	}
	log.Printf("Task database operation: %v", err)
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Ошибка работы с базой данных"})
}
