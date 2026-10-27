package api

import (
	"net/http"

	"github.com/Yandex-Practicum/go_final_project/pkg/db"
)

func editTaskHandler(w http.ResponseWriter, r *http.Request) {
	task, err := decodeTask(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := checkID(task.ID); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := checkDate(task); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := db.UpdateTask(task); err != nil {
		writeTaskError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, struct{}{})
}
