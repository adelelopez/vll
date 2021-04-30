package page

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/font/basicfont"
)

const (
	pxSize = 5
)

type Pair struct {
	Negative *Bubble
	Positive *Bubble
}

type Page struct {
	Root           *Bubble
	Grabbed        *Bubble
	GrabbedAtX     int
	GrabbedAtY     int
	GrabbedParent  *Bubble
	Highlighted    []*Bubble
	win            *pixelgl.Window
	Atlas          *text.Atlas
	AssumptionPair *Pair

	Mode string
}

func init() {
}

// first goal: get MLL to display correctly

// idea
// keep terms in the same bubble within a fixed distance of each other
// terms above some terms should be within a moderate range of distances
// terms move automatically in order to keep these constraints

// UI
// There should be a "Create" mode and a "Prove" mood (and "Fun" mode, which makes and manipulates tautologies automatically)
// Create
// Click on a bubble should select it
// Click and drag a bubble should let you move it around
// Right-click on a bubble should give drop-down (but a cute and cool looking one)
// 	options should be to
//  create/remove black/white loop
//

// Next goal
// get drag out and drag in to look good and feel right
// grabbed velocity determines whether or not a bubble gets pulled out or just moved around
// maybe that works for getting dragged in too?

// Now what?
// Allow multiple bubbles to be highlighted
// Bubbles with variables are leaves, and can't have stuff added to them.
// Backspace or Delete on a highlighted bubble removes it
// Variables are only added if one bubble is highlighted
// If a leaf is highlighted, then typing appends to the existing variable.
// Space on highlighted bubbles creates a new loop around it if they are siblings with a parent of the same kind

// When dragging a bubble in proof mode:
//   If it's a white bubble, then it stays part of its parent if the nearest alternative is "above" it
//     and it only becomes detached if the nearest alternative is a white region below its parent.
//     If it's dropped over a black bubble below it, it checks if the bubble is its opposite.
//       If so, both are removed. (ideally, there should be a way to make them "attract" while it's still grabbed)
//   If it's a black bubble, then it stays part of its parent if the nearest alternative is "below" it
//     and it only becomes detached if the nearest alternative is a black region above its parent.
// In proof mode, a right click on a white bubble in a black bubble creates two bubbles of opposite polarity.
//   While both of these bubbles are highlighted, they can be edited as in creative mode, with the same changes
//   happening in each, but with opposite polarities. (cursor should change type to indicate this)

func NewPage(win *pixelgl.Window) *Page {
	basicAtlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)

	page := &Page{win: win, Atlas: basicAtlas}
	c := color.RGBA{
		151,
		151,
		184,
		255,
	}

	page.Root = &Bubble{
		color: c,
		Kind:  BACKGROUND,
	}
	page.Mode = "Create"
	return page
}

func (pg *Page) childrenBoundary(b *Bubble, x, y int) float64 {
	squaredSum, closestD2 := 0.0, math.MaxFloat64

	b.Iterate(func(bub *Bubble) {
		sign := -1.0
		// I really like this repel effect, but it takes too long to compute
		// need to find a better way
		// ancestor := b.Parent
		// for ancestor != nil {
		// 	if ancestor == bub {
		// 		sign = 0
		// 		break
		// 		// return
		// 	}
		// 	ancestor = ancestor.Parent
		// }
		// descendent := bub
		// for descendent != nil {
		// 	if descendent == b {
		// 		sign = 1
		// 		break
		// 	}
		// 	descendent = descendent.Parent
		// }
		sign = 1

		// find squared distance of pixel from the circle
		var dx, dy float64
		dx = float64(x - bub.X)
		dy = float64(y - bub.Y)
		// make them ellipsoid if they have lots of text
		if len(bub.Variable) > 2 {
			dx /= float64(len(bub.Variable)) * 0.36
		}

		d2 := dx*dx + dy*dy

		if b.Variable == "" && b.Depth > 1 && b.Height < 2 {
			squaredSum += 900.0 / d2

			// keep track of the color and distance of the closest circle
			if d2 < closestD2 && b != nil {
				closestD2 = d2
			}
		} else {
			if sign > 0 {
				squaredSum += circSquared / d2

				// keep track of the color and distance of the closest circle
				if d2 < closestD2 && b != nil {
					closestD2 = d2
				}
			} else if sign < 0 {
				squaredSum -= circSquared / d2
			}
		}
	})

	// the sum is the L^p norm of the distances from the pixel to the boundary of each circle
	// sum := math.Sqrt(squaredSum)
	return squaredSum
}

func (pg *Page) IsHighlighted(b *Bubble) bool {
	for _, bub := range pg.Highlighted {
		if b == bub {
			return true
		}
	}
	return false
}

func (pg *Page) InAssumptionMode() bool {
	return pg.AssumptionPair != nil && pg.AssumptionPair.Positive != nil && pg.AssumptionPair.Negative != nil
}

func (pg *Page) colorBubble(b *Bubble, x, y int) color.Color {
	if b == nil {
		return color.Black
	}
	clr := b.color
	if b.Kind == WHITE {
		clr = color.White
	}
	if b.Kind == BLACK {
		clr = color.Black
	}
	if pg.IsHighlighted(b) {
		if b.Kind == WHITE && (x/pxSize-y/pxSize)%2 == 0 {
			clr = color.RGBA{
				R: 255,
				G: 255,
				B: 100,
				A: 255,
			}
		}
		if b.Kind == BLACK && (x/pxSize-y/pxSize)%2 == 0 {
			clr = color.RGBA{
				R: 50,
				G: 0,
				B: 135,
				A: 255,
			}
		}
	} else if pg.InAssumptionMode() {
		if pg.AssumptionPair.Positive.IsAbove(b) || pg.AssumptionPair.Negative.IsAbove(b) {
			if b.Kind == WHITE {
				clr = color.RGBA{
					R: 255,
					G: 255,
					B: 100,
					A: 255,
				}
			}
			if b.Kind == BLACK {
				clr = color.RGBA{
					R: 50,
					G: 0,
					B: 135,
					A: 255,
				}
			}
		}
	}
	return clr
}

func (p *Page) BelongsTo(x, y int) *Bubble {
	var owner *Bubble
	owner = p.Root

	p.Root.bfs(func(bub *Bubble) {
		dist := p.childrenBoundary(bub, x, y)
		for i := bub.Depth; i > 0; i-- {
			n := (bub.Depth - i)
			if bub.Variable == "" {
				n++
			}
			if dist > thresh(n-bub.Height) {
				owner = bub
			}
		}
	})
	return owner
}

func (p *Page) BelongsToGrabbed(x, y int) *Bubble {
	var owner *Bubble
	owner = p.Root

	p.Grabbed.bfs(func(bub *Bubble) {
		dist := p.childrenBoundary(bub, x, y)
		for i := bub.Depth; i > 0; i-- {
			n := (bub.Depth - i)
			if bub.Variable == "" {
				n++
			}
			if dist > thresh(n-bub.Height) {
				owner = bub
			}
		}
	})
	return owner
}

func (p *Page) NearestAlternative(x, y int) *Bubble {
	var owner *Bubble
	owner = p.Root

	p.Root.bfs(func(bub *Bubble) {
		if bub == p.Root {
			return
		}
		dist := p.childrenBoundary(bub, x, y)

		for i := bub.Depth; i > 0; i-- {
			n := (bub.Depth - i)
			if bub.Variable == "" {
				n++
			}
			if dist > thresh(n-bub.Height) {
				excluded := false
				if p.Grabbed != nil {
					p.Grabbed.Iterate(func(ex *Bubble) {
						if bub == ex {
							excluded = true
						}
					})
				}
				if !excluded {
					owner = bub
				}
			}
		}
	})
	return owner
}

func thresh(n int) float64 {
	return 0.5 * math.Pow(1.311, float64(n))
}

func (pg *Page) DrawPicture() *pixel.PictureData {
	m := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(m, m.Bounds(), &image.Uniform{color.Black}, image.ZP, draw.Src)

	for x := sidebar; x < width; x += pxSize {
		for y := 0; y < height; y += pxSize {
			var clr color.Color

			b := pg.BelongsTo(x, y)
			clr = pg.colorBubble(b, x, y)

			rect := image.Rect(x, y, x+pxSize, y+pxSize)
			draw.Draw(m, rect, &image.Uniform{clr}, image.ZP, draw.Src)

			if pg.Grabbed != nil {
				b = pg.BelongsToGrabbed(x, y)
				if b != pg.Root {
					clr = pg.colorBubble(b, x, y)
					draw.Draw(m, rect, &image.Uniform{clr}, image.ZP, draw.Src)
				}
			}
		}
	}
	return pixel.PictureDataFromImage(m)
}

func (pg *Page) Label() {
	pg.Root.Iterate(func(b *Bubble) {
		centerX := float64(b.X + 3)
		if len(b.Variable) >= 1 {
			centerX -= float64(14 * len(b.Variable))
		}
		basicTxt := text.New(pixel.V(centerX, height-float64(b.Y+55)), pg.Atlas)
		if b.Kind == WHITE {
			basicTxt.Color = color.Black
		} else {
			basicTxt.Color = color.White
		}
		fmt.Fprintln(basicTxt, b.Variable)
		basicTxt.Draw(pg.win, pixel.IM.Scaled(basicTxt.Orig, 4))
	})
}

func (pg *Page) NormalizeHeight() {
	pg.Root.Iterate(func(b *Bubble) {
		b.normalizeHeight()
	})
}

func (pg *Page) Grab(bub *Bubble, x, y int) {
	pg.Grabbed = bub
	pg.GrabbedParent = bub.Parent
	pg.GrabbedAtX = x
	pg.GrabbedAtY = y
	pg.Highlighted = []*Bubble{bub}
}

//   If it's in a white bubble, then it stays part of its parent if the nearest alternative is "above" it
//     and it only becomes detached if the nearest alternative is a white region below its parent.
//     If it's dropped over a black bubble below it, it checks if the bubble is its opposite.
//       If so, both are removed. (ideally, there should be a way to make them "attract" while it's still grabbed)
//   If it's a black bubble, then it stays part of its parent if the nearest alternative is "below" it
//     and it only becomes detached if the nearest alternative is a black region above its parent.
func (pg *Page) CanPlaceAt(other *Bubble) bool {
	if pg.Grabbed != nil && pg.Grabbed.Parent == other {
		return true
	}
	if pg.GrabbedParent != nil {
		switch pg.GrabbedParent.Kind {
		case WHITE:
			if pg.GrabbedParent.IsAbove(other) {
				if other.Kind == WHITE {
					return true
				}
				if other.Kind == BLACK {
					if pg.Grabbed.String() == other.Opposite() {
						other.Parent.Detach(other)
						pg.GrabbedParent.Detach(pg.Grabbed)
						pg.Grabbed = nil
						pg.GrabbedParent = nil
						pg.Highlighted = nil
						return true
					}
				}
			}
			return false
		case BLACK:
			if other.IsAbove(pg.GrabbedParent) && other.Kind == BLACK {
				return true
			}
			return false
		}
	}
	return false
}

func (pg *Page) ReleaseInto(b *Bubble) {
	if pg.Grabbed != nil {
		parent := pg.GrabbedParent
		// if dropped into a variable, find a more appropriate parent to place it into
		if b.Variable != "" {
			if b.Parent.Kind == b.Kind {
				b = b.Parent
			} else {
				// place a buffer loop around the variable
				loop := NewBubble(b.X, b.Y, "", b.Kind)
				b.Parent.Insert(loop)
				b.Parent.Detach(b)
				loop.Insert(b)
				b = loop
			}
		}
		if b != parent || pg.GrabbedParent != pg.Grabbed.Parent {
			if parent != nil {
				parent.Detach(pg.Grabbed)
			}
			b.Insert(pg.Grabbed)
		}
		pg.Highlighted = []*Bubble{pg.Grabbed}

	}
	pg.Grabbed = nil
	pg.NormalizeHeight()
}
