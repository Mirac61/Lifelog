// Package ui holds the building blocks every page shares: the title row, the
// pill bar, the stepper, and the formatting helpers that go with them. It
// imports nothing from the pages, so any page can import it.
package ui

import "github.com/a-h/templ"

// Nav is how a control moves. Href navigates the page, Get swaps Target in
// place via htmx — a control is one or the other. The zero value is a dead
// end and renders disabled.
type Nav struct {
	Href   string
	Get    string
	Target string
	Swap   string
}

func (n Nav) isLink() bool { return n.Href != "" }

func (n Nav) isSwap() bool { return n.Get != "" }

// Pill is one entry of a pill bar.
type Pill struct {
	Label  string
	Active bool
	Nav    Nav
}

// isTablist: pills that swap a panel in place are tabs. Pills that navigate
// away are links and must not claim the tab roles.
func isTablist(items []Pill) bool {
	for _, p := range items {
		if !p.Nav.isSwap() {
			return false
		}
	}
	return len(items) > 0
}

// Stepper is the arrow · value · arrow control. A zero Prev or Next renders
// that arrow disabled.
type Stepper struct {
	Value     string
	Wide      bool // roomier value box, for word labels like "Gestern"
	Prev      Nav
	PrevLabel string
	Next      Nav
	NextLabel string
}

// Head is the page title row. The IDs are only needed where htmx swaps that
// part out of band; they go through spread attributes so the markup keeps its
// exact spacing.
type Head struct {
	Title      string
	Sub        string
	SubID      string
	ControlsID string
}

func (h Head) subAttrs() templ.Attributes { return idAttr(h.SubID) }

func (h Head) controlsAttrs() templ.Attributes { return idAttr(h.ControlsID) }

func idAttr(id string) templ.Attributes {
	if id == "" {
		return nil
	}
	return templ.Attributes{"id": id}
}
