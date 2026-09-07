// Package ui holds the design-system building blocks — the sticky page head,
// the tab bar and the year stepper from DesignPrototype/assets/lifelog.css —
// plus the formatting helpers the pages share. It imports nothing from the
// pages, so any page can import it.
package ui

// Nav is how a control moves: Get swaps Target in place via htmx, Href
// navigates the page. Push is the browser URL to record while swapping, which
// is rarely the same as Get. The zero value is a dead end and renders
// disabled.
type Nav struct {
	Href   string
	Get    string
	Target string
	Swap   string
	Push   string
}

func (n Nav) isLink() bool { return n.Href != "" }

func (n Nav) isSwap() bool { return n.Get != "" }

// Tab is one entry of a tab bar.
type Tab struct {
	Label  string
	Count  string // muted number beside the label, like the design's "Dateien 11"
	Active bool
	Nav    Nav
}

// Stepper is the arrow · value · arrow control. An arrow follows its Nav: a
// Href navigates, a Get swaps, the zero value renders disabled.
type Stepper struct {
	Label     string
	Group     string // aria-label of the group, e.g. "Jahr wechseln"
	Prev      Nav
	PrevLabel string
	Next      Nav
	NextLabel string
}
