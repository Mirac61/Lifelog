package ui

import "time"

// BestDaySuffix renders " · <date>" when a best date exists, or nothing.
func BestDaySuffix(date string) string {
	if date == "" {
		return ""
	}
	return " · " + FormatShortDate(date)
}

// FormatShortDate renders an ISO date as "2 Jan", like the heatmap month labels.
func FormatShortDate(date string) string {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	return t.Format("2 Jan")
}
