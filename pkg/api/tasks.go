package api

import (
	"log"
	"net/http"

	"github.com/Yandex-Practicum/go_final_project/pkg/db"
)

type TasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Метод не поддерживается"})
		return
	}
	tasks, err := db.SearchTasks(50, r.URL.Query().Get("search"))
	if err != nil {
		log.Printf("Get tasks: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Не удалось получить задачи"})
		return
	}
	writeJSON(w, http.StatusOK, TasksResp{Tasks: tasks})
}
