package main

import "math"

// named is a working set of the CSS/X11 colour keywords. It's not the
// full list of 147 - just enough to be useful. Add more as needed; the
// only rule is that each value must be a valid 6-digit hex colour.
var named = map[string]string{
	"black":      "#000000",
	"white":      "#ffffff",
	"gray":       "#808080",
	"silver":     "#c0c0c0",
	"red":        "#ff0000",
	"maroon":     "#800000",
	"orange":     "#ffa500",
	"yellow":     "#ffff00",
	"olive":      "#808000",
	"lime":       "#00ff00",
	"green":      "#008000",
	"teal":       "#008080",
	"cyan":       "#00ffff",
	"aqua":       "#00ffff",
	"blue":       "#0000ff",
	"navy":       "#000080",
	"purple":     "#800080",
	"magenta":    "#ff00ff",
	"pink":       "#ffc0cb",
	"hotpink":    "#ff69b4",
	"crimson":    "#dc143c",
	"salmon":     "#fa8072",
	"coral":      "#ff7f50",
	"tomato":     "#ff6347",
	"gold":       "#ffd700",
	"khaki":      "#f0e68c",
	"lavender":   "#e6e6fa",
	"indigo":     "#4b0082",
	"violet":     "#ee82ee",
	"orchid":     "#da70d6",
	"turquoise":  "#40e0d0",
	"skyblue":    "#87ceeb",
	"steelblue":  "#4682b4",
	"royalblue":  "#4169e1",
	"slateblue":  "#6a5acd",
	"chocolate":  "#d2691e",
	"sienna":     "#a0522d",
	"brown":      "#a52a2a",
	"tan":        "#d2b48c",
	"beige":      "#f5f5dc",
	"ivory":      "#fffff0",
	"linen":      "#faf0e6",
	"plum":       "#dda0dd",
	"orchidred":  "#ff69b4",
	"forestgreen": "#228b22",
	"seagreen":   "#2e8b57",
	"olivedrab":  "#6b8e23",
	"darkgreen":  "#006400",
	"darkred":    "#8b0000",
	"darkblue":   "#00008b",
	"darkorange": "#ff8c00",
	"darkviolet": "#9400d3",
	"deeppink":   "#ff1493",
	"firebrick":  "#b22222",
	"goldenrod":  "#daa520",
	"midnightblue": "#191970",
	"peru":       "#cd853f",
	"slategray":  "#708090",
	"dimgray":    "#696969",
	"lightgray":  "#d3d3d3",
	"whitesmoke": "#f5f5f5",
}

// labColor caches the Lab coordinates for a named colour so nearest
// doesn't have to reconvert the whole table on every call.
type labColor struct {
	name       string
	hex        string
	l, a, b    float64
}

var labTable []labColor

func init() {
	labTable = make([]labColor, 0, len(named))
	for name, hex := range named {
		r, g, b, err := parseHex(hex)
		if err != nil {
			// A bad entry here is a bug in this file, not user input.
			panic("colors.go: invalid built-in colour " + hex + ": " + err.Error())
		}
		l, a, bb := rgbToLab(r, g, b)
		labTable = append(labTable, labColor{name: name, hex: hex, l: l, a: a, b: bb})
	}
}

// nearest returns the closest named colour to the given Lab coordinates
// under the supplied deltaE metric, along with the distance.
func nearest(l, a, b float64, deltaE func(l1, a1, b1, l2, a2, b2 float64) float64) (name, hex string, dist float64) {
	best := math.Inf(1)
	for _, c := range labTable {
		d := deltaE(l, a, b, c.l, c.a, c.b)
		if d < best {
			best = d
			name, hex = c.name, c.hex
		}
	}
	return name, hex, best
}
