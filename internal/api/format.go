package api

import (
	"fmt"
	"strconv"
	"strings"
)

// deNumber groups thousands with a dot, as German prose does.
func deNumber(n int) string {
	if n < 0 {
		return "-" + deNumber(-n)
	}
	s := strconv.Itoa(n)
	var out strings.Builder
	for i, r := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out.WriteByte('.')
		}
		out.WriteRune(r)
	}
	return out.String()
}

// deDecimal renders one decimal place with a comma — "3,0" stays "3,0",
// because a figure in a row of figures keeps its place.
func deDecimal(v float64) string {
	return strings.Replace(fmt.Sprintf("%.1f", v), ".", ",", 1)
}

// deTrim is deDecimal for prose, where a trailing ",0" only adds noise:
// "9× so viel", not "9,0× so viel".
func deTrim(v float64) string {
	return strings.TrimSuffix(deDecimal(v), ",0")
}
