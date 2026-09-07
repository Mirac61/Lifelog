package api

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/Mirac61/lifelog/internal/store"
	"github.com/Mirac61/lifelog/web/templ/git"
	"github.com/Mirac61/lifelog/web/templ/ui"
)

// prShown caps the pull request list; the rest becomes a line of prose, the
// way the design does it.
const prShown = 8

var ordinals = [...]string{"", "", "zweiten", "dritten", "vierten", "fünften", "sechsten", "siebten", "achten", "neunten", "zehnten"}

var prLabels = map[string]string{"open": "offen", "merged": "merged", "closed": "geschlossen"}

// gitView is one year's raw material. Every method on it turns numbers into
// the strings the design's panes print.
type gitView struct {
	year      int
	prevYear  int
	label     string // plural, e.g. "Beiträge"
	singular  string // "ein Beitrag"
	eventType string
	stats     store.Stats
	prev      store.Stats
	totals    []store.DailyTotal
	repos     []store.RepoCommit
	prs       []store.PullRequest
	months    []git.MonthBar
}

func (v gitView) headMeta() string {
	parts := []string{fmt.Sprintf("%s %s", deNumber(int(v.stats.Total)), v.label)}
	if len(v.repos) > 0 {
		parts = append(parts, fmt.Sprintf("%d Repos", len(v.repos)))
	}
	if len(v.prs) > 0 && v.eventType != "pull_request" {
		parts = append(parts, fmt.Sprintf("%d Pull Requests", len(v.prs)))
	}
	return strings.Join(parts, " · ")
}

func (v gitView) lead() git.Lead {
	lead := git.Lead{
		Total:    deNumber(int(v.stats.Total)),
		Caption:  fmt.Sprintf("%s in %d", v.label, v.year),
		Sentence: v.sentence(),
	}
	if v.stats.ActiveDays == 0 {
		return lead
	}
	lead.Stats = []git.Figure{
		{Value: fmt.Sprint(v.stats.ActiveDays), Label: fmt.Sprintf("aktive Tage von %d", ui.DaysInYear(v.year))},
		{Value: fmt.Sprint(v.stats.LongestStreak), Label: "längste Serie in Tagen"},
		{Value: deDecimal(v.stats.Total / float64(v.stats.ActiveDays)), Label: "Ø je aktivem Tag"},
	}
	if len(v.repos) > 0 {
		lead.Stats = append(lead.Stats, git.Figure{Value: fmt.Sprint(len(v.repos)), Label: "Repos berührt"})
	}
	return lead
}

// cadence reads the year as a rhythm — "an jedem dritten Tag" — because a
// count of active days says nothing on its own.
func (v gitView) cadence() string {
	if v.stats.ActiveDays == 0 {
		return ""
	}
	every := float64(ui.DaysInYear(v.year)) / float64(v.stats.ActiveDays)
	switch n := int(math.Round(every)); {
	case every < 1.35:
		return "An fast jedem Tag " + v.singular
	case n < len(ordinals):
		return fmt.Sprintf("An jedem %s Tag %s", ordinals[n], v.singular)
	default:
		return fmt.Sprintf("An jedem %d. Tag %s", n, v.singular)
	}
}

func (v gitView) sentence() git.Sentence {
	if v.stats.Total == 0 {
		return git.Sentence{Before: fmt.Sprintf("Keine %s in %d.", v.label, v.year)}
	}
	cadence := v.cadence()
	ratio := 0.0
	if v.prev.Total > 0 {
		ratio = v.stats.Total / v.prev.Total
	}
	switch {
	case ratio >= 1.15:
		return git.Sentence{
			Before: cadence + " — ",
			Em:     deTrim(ratio) + "× so viel",
			After:  fmt.Sprintf(" wie %d.", v.prevYear),
		}
	case ratio > 0 && ratio <= 0.85:
		return git.Sentence{
			Before: cadence + " — ",
			Em:     fmt.Sprintf("%d %%", int(ratio*100+0.5)),
			After:  fmt.Sprintf(" dessen, was %d trug.", v.prevYear),
		}
	case v.stats.ActiveDays > 0:
		return git.Sentence{
			Before: cadence + " — im Schnitt ",
			Em:     deDecimal(v.stats.Total / float64(v.stats.ActiveDays)),
			After:  " je aktivem Tag.",
		}
	}
	return git.Sentence{Before: cadence + "."}
}

// insights reads the same numbers a second time, one sentence per finding.
// Each one is skipped when its data is missing, so a thin year stays honest
// instead of printing zeros.
func (v gitView) insights() []git.Insight {
	var out []git.Insight
	if v.prev.Total > 0 {
		out = append(out, git.Insight{
			Key: deTrim(v.stats.Total/v.prev.Total) + "×",
			Text: fmt.Sprintf(
				"Gegenüber %d: von %s auf %s %s, von %d auf %d aktive Tage.",
				v.prevYear, deNumber(int(v.prev.Total)), deNumber(int(v.stats.Total)),
				v.label, v.prev.ActiveDays, v.stats.ActiveDays),
		})
	}
	if best, err := time.Parse("2006-01-02", v.stats.BestDate); err == nil && v.stats.BestValue > 0 {
		out = append(out, git.Insight{
			Key: deNumber(int(v.stats.BestValue)),
			Text: fmt.Sprintf("Der %d. %s allein trägt %s %s — %s %% des ganzen Jahres an einem einzigen Tag.",
				best.Day(), ui.MonthLong[best.Month()], deNumber(int(v.stats.BestValue)), v.label,
				deTrim(v.stats.BestValue/v.stats.Total*100)),
		})
	}
	if len(v.repos) > 0 {
		var sum float64
		for _, r := range v.repos {
			sum += r.Value
		}
		top := v.repos[0]
		out = append(out, git.Insight{
			Key: deNumber(int(top.Value)),
			Text: fmt.Sprintf("%s trägt %s der %s Commits aus %d Repos.",
				top.Name, deNumber(int(top.Value)), deNumber(int(sum)), len(v.repos)),
		})
	}
	if month, value, share := v.peakMonth(); month != "" {
		out = append(out, git.Insight{
			Key: deNumber(int(value)),
			Text: fmt.Sprintf("%s ist der stärkste Monat und trägt %s der %s %s — %s %% des Jahres.",
				month, deNumber(int(value)), deNumber(int(v.stats.Total)), v.label, deTrim(share)),
		})
	}
	return out
}

func (v gitView) peakMonth() (name string, value, share float64) {
	if v.stats.Total == 0 {
		return "", 0, 0
	}
	for i, m := range v.months {
		if m.Peak {
			value = v.monthValue(i)
			return ui.MonthLong[i+1], value, value / v.stats.Total * 100
		}
	}
	return "", 0, 0
}

// monthValue re-reads the month sum from the totals; MonthBar only keeps the
// formatted string.
func (v gitView) monthValue(index int) float64 {
	var sum float64
	for _, t := range v.totals {
		d, err := time.Parse("2006-01-02", t.LocalDate)
		if err != nil || d.Year() != v.year || int(d.Month()) != index+1 {
			continue
		}
		sum += t.Value
	}
	return sum
}

func (v gitView) repoRows() ([]git.RepoRow, string) {
	if len(v.repos) == 0 {
		return nil, ""
	}
	top := v.repos
	if len(top) > 5 {
		top = top[:5]
	}
	peak := v.repos[0].Value
	rows := make([]git.RepoRow, 0, len(top))
	for i, r := range top {
		owner, name := "", r.NameWithOwner
		if slash := strings.LastIndex(r.NameWithOwner, "/"); slash >= 0 {
			owner, name = r.NameWithOwner[:slash+1], r.NameWithOwner[slash+1:]
		}
		percent := 0.0
		if peak > 0 {
			percent = r.Value / peak * 100
		}
		rows = append(rows, git.RepoRow{
			Rank:    fmt.Sprintf("%02d", i+1),
			Owner:   owner,
			Name:    name,
			Percent: percent,
			Value:   deNumber(int(r.Value)),
		})
	}
	note := fmt.Sprintf("%d im Jahr", len(v.repos))
	if len(v.repos) > len(rows) {
		note = fmt.Sprintf("Top %d von %d", len(rows), len(v.repos))
	}
	return rows, note
}

// prGroups keeps the newest pull requests under their month heading and turns
// the remainder into one line, so a busy year does not become a wall.
func (v gitView) prGroups() ([]git.PRGroup, string, string) {
	if len(v.prs) == 0 {
		return nil, "", ""
	}
	shown := v.prs
	if len(shown) > prShown {
		shown = shown[:prShown]
	}
	var groups []git.PRGroup
	for _, pr := range shown {
		d, err := time.Parse("2006-01-02", pr.LocalDate)
		if err != nil {
			continue
		}
		month := ui.MonthLong[d.Month()]
		if len(groups) == 0 || groups[len(groups)-1].Month != month {
			groups = append(groups, git.PRGroup{Month: month})
		}
		state := strings.ToLower(pr.State)
		group := &groups[len(groups)-1]
		group.Rows = append(group.Rows, git.PRRow{
			Date:  ui.DayMonth(d),
			State: state,
			Label: prLabels[state],
			Title: pr.Title,
			Repo:  pr.Repo,
			URL:   pr.URL,
		})
	}
	return groups, fmt.Sprintf("%d im Jahr", len(v.prs)), v.prRest(len(shown))
}

func (v gitView) prRest(shown int) string {
	rest := len(v.prs) - shown
	if rest <= 0 {
		return ""
	}
	newest, err1 := time.Parse("2006-01-02", v.prs[shown].LocalDate)
	oldest, err2 := time.Parse("2006-01-02", v.prs[len(v.prs)-1].LocalDate)
	noun := fmt.Sprintf("%d weitere Pull Requests", rest)
	if rest == 1 {
		noun = "Ein weiterer Pull Request"
	}
	if err1 != nil || err2 != nil {
		return noun + "."
	}
	if oldest.Month() == newest.Month() {
		return fmt.Sprintf("%s aus %s.", noun, ui.MonthLong[oldest.Month()])
	}
	return fmt.Sprintf("%s aus %s bis %s.", noun, ui.MonthLong[oldest.Month()], ui.MonthLong[newest.Month()])
}
