# nearcol

You have a hex colour - maybe pulled from a design tool, a screenshot, or
a CSS file someone else wrote - and you want to know one thing: what is
this *close to*, in words a person would use, and how close?

`nearcol` answers exactly that question. Give it a colour, it converts to
CIE L\*a\*b\* and reports the nearest named colour along with the
perceptual distance (CIE76 deltaE) between them.

```
$ nearcol '#3366ff'
input:  #3366ff  (L=42.9 a=27.4 b=-73.6)
match:  royalblue (#4169e1)
deltaE: 11.87  (clearly a different colour)

$ nearcol c0c0c0
input:  #c0c0c0  (L=77.7 a=0.0 b=0.0)
match:  silver (#c0c0c0)
deltaE: 0.00  (imperceptible difference)
```

Hex input can be 3 or 6 digits, with or without a leading `#`.

## why Lab, not RGB

Comparing colours by raw RGB distance is misleading: it treats every
channel as equally important to the eye, which isn't true. Converting to
CIE L\*a\*b\* first puts the comparison in a space that was designed to
track perceived difference, so "closest colour" actually means something
close to what a person would call closest.

## deltaE, roughly

- < 1: no visible difference
- 1-2.5: visible only to a trained eye, side by side
- 2.5-10: visible at a glance
- 10-25: clearly a different colour
- \> 25: not a meaningful match

This is CIE76 (plain Euclidean distance in Lab), not CIEDE2000, so treat
the numbers as a rough guide rather than a calibrated measurement.

## build

```
go build -o nearcol .
```

Standard library only, no dependencies.

## the colour table

`colors.go` holds a working set of CSS/X11 colour keywords, not the full
147. If the match you get back feels too coarse, that's the table to
extend - add a name and a 6-digit hex value.
