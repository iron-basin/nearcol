// nearcol answers one question: given an sRGB colour, which named colour
// is it perceptually closest to, and how far away is it?
package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

func main() {
	ciede2000 := flag.Bool("ciede2000", false, "use CIEDE2000 instead of CIE76 for deltaE")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: nearcol [-ciede2000] <hex-colour>   e.g. nearcol #3366ff")
	}
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		flag.Usage()
		os.Exit(1)
	}

	r, g, b, err := parseHex(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "nearcol: %v\n", err)
		os.Exit(1)
	}

	deltaE, label := deltaE76, "CIE76"
	if *ciede2000 {
		deltaE, label = deltaE2000, "CIEDE2000"
	}

	l, a, bb := rgbToLab(r, g, b)
	name, hex, dist := nearest(l, a, bb, deltaE)

	fmt.Printf("input:  #%02x%02x%02x  (L=%.1f a=%.1f b=%.1f)\n", r, g, b, l, a, bb)
	fmt.Printf("match:  %s (%s)\n", name, hex)
	fmt.Printf("deltaE: %.2f  %s  [%s]\n", dist, describe(dist), label)
}

// describe gives a rough plain-English sense of a deltaE value. The
// thresholds are the commonly cited rules of thumb; they were derived for
// CIE76 but stay roughly right for CIEDE2000 too since both scales are
// anchored to a just-noticeable-difference of about 1.
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

// deltaE2000 is the CIEDE2000 colour difference formula (Sharma, Wu, Dalal
// 2005). It corrects known non-uniformities in CIE76 - chroma and hue get
// their own weighting, plus a rotation term that fixes the blue region
// where CIE76 overstates distance. More faithful to how people actually
// judge closeness, at the cost of being unreadable at a glance.
func deltaE2000(l1, a1, b1, l2, a2, b2 float64) float64 {
	c1 := math.Hypot(a1, b1)
	c2 := math.Hypot(a2, b2)
	avgC := (c1 + c2) / 2

	g := 0.5 * (1 - math.Sqrt(math.Pow(avgC, 7)/(math.Pow(avgC, 7)+math.Pow(25, 7))))

	a1p := a1 * (1 + g)
	a2p := a2 * (1 + g)

	c1p := math.Hypot(a1p, b1)
	c2p := math.Hypot(a2p, b2)
	avgCp := (c1p + c2p) / 2

	h1p := atan2Deg(b1, a1p)
	h2p := atan2Deg(b2, a2p)

	var deltahp float64
	switch {
	case c1p*c2p == 0:
		deltahp = 0
	case math.Abs(h1p-h2p) <= 180:
		deltahp = h2p - h1p
	case h2p <= h1p:
		deltahp = h2p - h1p + 360
	default:
		deltahp = h2p - h1p - 360
	}

	deltaLp := l2 - l1
	deltaCp := c2p - c1p
	deltaHp := 2 * math.Sqrt(c1p*c2p) * math.Sin(degToRad(deltahp)/2)

	var avgHp float64
	switch {
	case c1p*c2p == 0:
		avgHp = h1p + h2p
	case math.Abs(h1p-h2p) <= 180:
		avgHp = (h1p + h2p) / 2
	case h1p+h2p < 360:
		avgHp = (h1p + h2p + 360) / 2
	default:
		avgHp = (h1p + h2p - 360) / 2
	}

	avgLp := (l1 + l2) / 2

	t := 1 - 0.17*math.Cos(degToRad(avgHp-30)) +
		0.24*math.Cos(degToRad(2*avgHp)) +
		0.32*math.Cos(degToRad(3*avgHp+6)) -
		0.20*math.Cos(degToRad(4*avgHp-63))

	deltaTheta := 30 * math.Exp(-math.Pow((avgHp-275)/25, 2))
	rc := 2 * math.Sqrt(math.Pow(avgCp, 7)/(math.Pow(avgCp, 7)+math.Pow(25, 7)))
	rt := -math.Sin(degToRad(2*deltaTheta)) * rc

	sl := 1 + (0.015*math.Pow(avgLp-50, 2))/math.Sqrt(20+math.Pow(avgLp-50, 2))
	sc := 1 + 0.045*avgCp
	sh := 1 + 0.015*avgCp*t

	const kl, kc, kh = 1, 1, 1

	dl := deltaLp / (kl * sl)
	dc := deltaCp / (kc * sc)
	dh := deltaHp / (kh * sh)

	return math.Sqrt(dl*dl + dc*dc + dh*dh + rt*dc*dh)
}

func degToRad(d float64) float64 { return d * math.Pi / 180 }

// atan2Deg is math.Atan2 in degrees, normalised to [0, 360).
func atan2Deg(y, x float64) float64 {
	if y == 0 && x == 0 {
		return 0
	}
	deg := math.Atan2(y, x) * 180 / math.Pi
	if deg < 0 {
		deg += 360
	}
	return deg
}
