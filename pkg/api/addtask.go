package api

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Yandex-Practicum/go_final_project/pkg/db"
)

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	task, err := decodeTask(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := checkDate(task); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	id, err := db.AddTask(task)
	if err != nil {
		log.Printf("Add task: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Не удалось сохранить задачу"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"id": strconv.FormatInt(id, 10)})
}

func checkDate(task *db.Task) error {
	now := time.Now()
	today := now.Format(DateFormat)
	if task.Date == "" {
		task.Date = today
	}
	if _, err := time.Parse(DateFormat, task.Date); err != nil {
		return fmt.Errorf("Некорректная дата задачи: %w", err)
	}
	if strings.TrimSpace(task.Title) == "" {
		return errors.New("Не указан заголовок задачи")
	}
	var next string
	if task.Repeat != "" {
		var err error
		next, err = NextDate(now, task.Date, task.Repeat)
		if err != nil {
			return fmt.Errorf("Некорректное правило повторения: %w", err)
		}
	}
	if task.Date < today {
		task.Date = today
		if task.Repeat != "" {
			task.Date = next
		}
	}
	return nil
}
