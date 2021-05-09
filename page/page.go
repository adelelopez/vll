package page

import (
	"fmt"

	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/font/basicfont"
)

const (
	pxSize = 5
)

type Pair struct {
	Positive *Bubble
	Negative *Bubble
}

type Page struct {
	Root          *Bubble
	Grabbed       *Bubble
	GrabbedAtX    int
	GrabbedAtY    int
	GrabbedParent *Bubble
	Highlighted   []*Bubble
	win           *pixelgl.Window
	Atlas         *text.Atlas

	Mode           string
	AssumptionMode bool
	AssumptionPair *Pair

	unprocessedBubbles []*Bubble
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

	page.Root = &Bubble{
		Kind: BACKGROUND,
	}
	page.Mode = "Create"
	return page
}

func (pg *Page) NewBubble(x, y int, v string, k Kind) *Bubble {
	newb := newBubble(x, y, v, k)
	pg.unprocessedBubbles = append(pg.unprocessedBubbles, newb)
	return newb
}

func (pg *Page) InAssumption(b *Bubble) bool {
	return pg.AssumptionPair.Positive.IsAbove(b) || pg.AssumptionPair.Negative.IsAbove(b)
}

func (pg *Page) IsHighlighted(b *Bubble) bool {
	for _, bub := range pg.Highlighted {
		if b == bub {
			return true
		}
	}
	return false
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
		case WHITE, BLUE:
			if pg.GrabbedParent.IsAbove(other) {

				between := other
				for !between.IsMult() {
					between = between.Parent
				}
				for between != pg.GrabbedParent {
					if !between.IsMult() {
						fmt.Println("between")
						return false
					}
					between = between.Parent
				}

				if other.Kind == WHITE {
					return true
				}
				if other.Kind == BLACK || other.Kind == RED {
					fmt.Println("almost")
					fmt.Println(pg.Grabbed.Tolestra())
					fmt.Println(other.Opposite())
					if pg.Grabbed.Tolestra() == other.Opposite() {
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
		case BLACK, RED:
			if other.IsAbove(pg.GrabbedParent) && other.Kind == BLACK {
				between := other
				for between != pg.GrabbedParent {
					if !between.IsMult() {
						return false
					}
					between = between.Parent
				}
				return true
			}
			return false
		}
	}
	return false
}

func (pg *Page) ReleaseInto(b *Bubble) {
	fmt.Println("releasing")
	if pg.Grabbed != nil && b != nil {
		parent := pg.GrabbedParent
		// if dropped into a variable, find a more appropriate parent to place it into
		if b.Variable != "" {
			if b.Parent.Kind == b.Kind {
				b = b.Parent
			} else {
				// place a buffer loop around the variable
				fmt.Println("buffer")
				loop := pg.NewBubble(b.X, b.Y, "", b.Kind)
				pg.Place(b.Parent, loop)
				pg.Delete(b)
				pg.Place(loop, b)
				b = loop
			}
		}
		if b != parent || pg.GrabbedParent != pg.Grabbed.Parent {
			if parent != nil {
				parent.Detach(pg.Grabbed)
			}
			fmt.Println("replace")
			pg.Place(b, pg.Grabbed)
		}
		pg.Highlighted = []*Bubble{pg.Grabbed}

	}
	pg.Grabbed = nil
}

func (pg *Page) Place(parent, b *Bubble) {
	fmt.Println("placing")
	if b.AssumptionPair != nil && parent.AssumptionPair != nil {
		parent.AssumptionPair.Insert(b.AssumptionPair)
	}
	parent.Insert(b)
}

func (pg *Page) Delete(b *Bubble) {
	fmt.Println("deleting")
	if b.AssumptionPair != nil {
		if b.AssumptionPair != pg.AssumptionPair.Positive && b.AssumptionPair != pg.AssumptionPair.Negative {
			b.AssumptionPair.Parent.Detach(b.AssumptionPair)
		}
	}
	b.Parent.Detach(b)
}

func (pg *Page) Loop(loopKind Kind, bubbles ...*Bubble) {
	// check that the  bubbles all have the same parent
	parent := bubbles[0].Parent
	for _, bubble := range bubbles {
		if bubble.Parent != parent {
			return
		}
	}

	var innerLoop *Bubble
	if len(bubbles) > 1 {
		innerLoop = pg.NewBubble(parent.X, parent.Y, "", parent.Kind)
	}

	outerLoop := pg.NewBubble(parent.X, parent.Y, "", loopKind)
	pg.ProcessNewBubbles()
	fmt.Println("has this", outerLoop.AssumptionPair)
	pg.Place(parent, outerLoop)
	if innerLoop != nil {
		pg.Place(outerLoop, innerLoop)
		for _, bubble := range bubbles {
			pg.Delete(bubble)
			pg.Place(innerLoop, bubble)
		}
	} else {
		for _, bubble := range bubbles {
			pg.Delete(bubble)
			pg.Place(outerLoop, bubble)
		}
	}

	outerLoop.CenterAroundChildren()
	pg.Highlighted = []*Bubble{outerLoop}
}

func (pg *Page) ProcessNewBubbles() {
	if pg.AssumptionMode {
		fmt.Println("need to process", len(pg.unprocessedBubbles), "bubbles")
		for _, b := range pg.unprocessedBubbles {
			if b != pg.AssumptionPair.Positive && b != pg.AssumptionPair.Negative {
				var bub *Bubble
				if pg.AssumptionPair.Positive.IsAbove(b) {
					bub = newBubble(pg.AssumptionPair.Negative.X, pg.AssumptionPair.Negative.Y, b.Variable, b.OppositeKind())
				} else {
					bub = newBubble(pg.AssumptionPair.Positive.X, pg.AssumptionPair.Positive.Y, b.Variable, b.OppositeKind())
				}
				if b.Parent != nil {
					bub.Parent = b.Parent.AssumptionPair
				}
				bub.AssumptionPair = b
				b.AssumptionPair = bub
				bub.Parent.Insert(bub)
			}
		}
	}
	pg.unprocessedBubbles = nil
}

func (pg *Page) Execute(f func()) {
	// if in assumption mode, disable othe actions
	if len(pg.Highlighted) > 0 {
		if pg.AssumptionMode && !pg.InAssumption(pg.Highlighted[0]) && !pg.InAssumption(pg.GrabbedParent) {
			return
		}
	}

	f()
	pg.ProcessNewBubbles()
	pg.NormalizeHeight()
}

func (pg *Page) ExitAssumptionMode() {
	pg.AssumptionPair = nil
	pg.AssumptionMode = false
	pg.Root.Iterate(func(b *Bubble) {
		b.AssumptionPair = nil
	})
}
