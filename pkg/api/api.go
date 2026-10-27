package api

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

var initOnce sync.Once

func Init() {
	initOnce.Do(func() {
		http.HandleFunc("/api/nextdate", nextDayHandler)
		http.HandleFunc("/api/signin", signinHandler)
		http.HandleFunc("/api/task", auth(taskHandler))
		http.HandleFunc("/api/task/done", auth(doneTaskHandler))
		http.HandleFunc("/api/tasks", auth(tasksHandler))
	})
}

func nextDayHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	query := r.URL.Query()
	now := time.Now()
	if query.Has("now") {
		var err error
		now, err = time.Parse(DateFormat, query.Get("now"))
		if err != nil {
			http.Error(w, "invalid now date", http.StatusBadRequest)
			return
		}
	}
	next, err := NextDate(now, query.Get("date"), query.Get("repeat"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprint(w, next)
}
