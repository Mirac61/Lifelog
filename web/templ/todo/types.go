package todo

import (
	"time"

	"github.com/Mirac61/lifelog/internal/store"
)

type DayPage struct {
	Day      time.Time
	Todos    []store.Todo
	Upcoming []store.Todo // due after Day, earliest first
	Planned  int
}
