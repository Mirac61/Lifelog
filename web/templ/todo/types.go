package todo

import (
	"time"

	"github.com/Mirac61/lifelog/internal/store"
)

// Deliberately free of event data: the day detail view composes both.
type DayPage struct {
	Day     time.Time
	Todos   []store.Todo
	Planned int
}
