// nearcol answers one question: given an sRGB colour, which named colour
// is it perceptually closest to, and how far away is it?
package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: nearcol <hex-colour>   e.g. nearcol #3366ff")
		os.Exit(1)
	}

	r, g, b, err := parseHex(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "nearcol: %v\n", err)
		os.Exit(1)
	}

	l, a, bb := rgbToLab(r, g, b)
	name, hex, dist := nearest(l, a, bb)

	fmt.Printf("input:  #%02x%02x%02x  (L=%.1f a=%.1f b=%.1f)\n", r, g, b, l, a, bb)
	fmt.Printf("match:  %s (%s)\n", name, hex)
	fmt.Printf("deltaE: %.2f  %s\n", dist, describe(dist))
}

// describe gives a rough plain-English sense of a CIE76 deltaE value.
// The thresholds are the commonly cited rules of thumb, not a precise
// perceptual model (that would need CIEDE2000).
func describe(dist float64) string {
	switch {
	case dist < 1:
		return "(imperceptible difference)"
	case dist < 2.5:
		return "(barely noticeable to a trained eye)"
	case dist < 10:
		return "(noticeable at a glance)"
	case dist < 25:
		return "(clearly a different colour)"
	default:
		return "(not a good match)"
	}
}

// parseHex accepts "#rgb", "#rrggbb", "rgb" or "rrggbb".
func parseHex(s string) (r, g, b uint8, err error) {
	s = strings.TrimPrefix(strings.TrimSpace(s), "#")
	switch len(s) {
	case 3:
		s = string([]byte{s[0], s[0], s[1], s[1], s[2], s[2]})
	case 6:
		// already full length
	default:
		return 0, 0, 0, fmt.Errorf("expected a 3 or 6 digit hex colour, got %q", s)
	}

	v, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid hex colour %q: %w", s, err)
	}
	return uint8(v >> 16), uint8(v >> 8), uint8(v), nil
}

// rgbToLab converts 8-bit sRGB to CIE L*a*b* under a D65 illuminant.
func rgbToLab(r, g, b uint8) (l, a, bb float64) {
	x, y, z := rgbToXYZ(r, g, b)
	return xyzToLab(x, y, z)
}

func rgbToXYZ(r, g, b uint8) (x, y, z float64) {
	rl := srgbToLinear(float64(r) / 255)
	gl := srgbToLinear(float64(g) / 255)
	bl := srgbToLinear(float64(b) / 255)

	// sRGB -> XYZ matrix, D65 white point.
	x = rl*0.4124564 + gl*0.3575761 + bl*0.1804375
	y = rl*0.2126729 + gl*0.7151522 + bl*0.0721750
	z = rl*0.0193339 + gl*0.1191920 + bl*0.9503041
	return x, y, z
}

func srgbToLinear(c float64) float64 {
	if c <= 0.04045 {
		return c / 12.92
	}
	return math.Pow((c+0.055)/1.055, 2.4)
}

// D65 reference white, 2 degree observer.
const (
	whiteX = 0.95047
	whiteY = 1.00000
	whiteZ = 1.08883
)

func xyzToLab(x, y, z float64) (l, a, b float64) {
	fx := labF(x / whiteX)
	fy := labF(y / whiteY)
	fz := labF(z / whiteZ)

	l = 116*fy - 16
	a = 500 * (fx - fy)
	b = 200 * (fy - fz)
	return l, a, b
}

func labF(t float64) float64 {
	const (
		delta = 6.0 / 29.0
	)
	if t > delta*delta*delta {
		return math.Cbrt(t)
	}
	return t/(3*delta*delta) + 4.0/29.0
}

// deltaE76 is the CIE76 colour difference: plain Euclidean distance in
// Lab space. It's not as perceptually uniform as CIEDE2000, but it's
// simple and good enough for "which named colour is this closest to".
func deltaE76(l1, a1, b1, l2, a2, b2 float64) float64 {
	dl := l1 - l2
	da := a1 - a2
	db := b1 - b2
	return math.Sqrt(dl*dl + da*da + db*db)
}
