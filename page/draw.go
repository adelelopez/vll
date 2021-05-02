package page

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/text"
)

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

func (pg *Page) colorBubble(b *Bubble, x, y int) color.Color {
	if b == nil {
		return color.Black
	}
	clr := b.Kind
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
	} else if pg.AssumptionMode {
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

func (pg *Page) drawLabel(b *Bubble) {
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
}

func (pg *Page) Label() {
	pg.Root.Iterate(func(b *Bubble) {
		pg.drawLabel(b)
	})
	if pg.Grabbed != nil {
		pg.Grabbed.Iterate(func(b *Bubble) {
			pg.drawLabel(b)
		})
	}
}
