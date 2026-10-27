package api

import (
	"net/http"
	"time"

	"github.com/Yandex-Practicum/go_final_project/pkg/db"
)

func doneTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Метод не поддерживается"})
		return
	}
	id := r.URL.Query().Get("id")
	if err := checkID(id); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	task, err := db.GetTask(id)
	if err != nil {
		writeTaskError(w, err)
		return
	}
	if task.Repeat == "" {
		err = db.DeleteTask(id)
	} else {
		next, nextErr := NextDate(time.Now(), task.Date, task.Repeat)
		if nextErr != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": nextErr.Error()})
			return
		}
		err = db.UpdateDate(next, id)
	}
	if err != nil {
		writeTaskError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, struct{}{})
}
