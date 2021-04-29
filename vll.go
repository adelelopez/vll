package main

import (
	"fmt"
	"image/color"
	"math"
	"strings"
	"time"
	"vll/page"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
)

const (
	width  = 1024
	height = 640
)

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// Now that MLL is basically working, what should we focus on next?
// Pretty buggy still, should get some testing done
// Implement deleting and adding loops
// Implement proof vs creative mode
// Implement multi-highlighting
// Implement hover highlighting (so you can see where you're dropping something)
//  do this with a map so it's fast
// To look okay, we need to prevent bubbles from ever getting split. Not sure how to enforce this.
func update(pg *page.Page) {
	pg.Root.Iterate(func(b *page.Bubble) {
		for i := 0; i < len(b.Children); i++ {
			for j := i + 1; j < len(b.Children); j++ {
				if page.Distance(b.Children[i], b.Children[j]) < float64(20.0*(b.Children[i].Height+b.Children[j].Height)+100) &&
					b.Children[i] != pg.Grabbed && b.Children[j] != pg.Grabbed {
					dx := 0
					dy := 0
					for dx*dy == 0 {
						dx = int(2*math.Atan(float64(b.Children[i].X-b.Children[j].X))) + page.Random(-2, 2)
						dy = int(2*math.Atan(float64(b.Children[i].Y-b.Children[j].Y))) + page.Random(-2, 2)
					}
					b.Children[i].X += dx
					b.Children[i].Y += dy
					b.Children[j].X -= dx
					b.Children[j].Y -= dy
				}
			}
		}
	})
}

func run() {
	win, err := pixelgl.NewWindow(pixelgl.WindowConfig{
		Bounds:    pixel.R(0, 0, float64(width), float64(height)),
		VSync:     true,
		Resizable: false,
	})
	if err != nil {
		panic(err)
	}

	pg := page.NewPage(win)

	go func() {
		for {
			update(pg)
			time.Sleep(60 * time.Millisecond)
		}
	}()

	grabbedX := 0
	grabbedY := 0

	for !win.Closed() {
		win.Update()

		p := pg.DrawPicture()
		s := pixel.NewSprite(p, p.Bounds())
		s.Draw(win, pixel.IM.Moved(win.Bounds().Center()))

		// TODO: figure out why these offsets are needed for things to line up properly
		x := int(win.MousePosition().X) - 1
		y := int(height-win.MousePosition().Y) + 38

		pg.Label()

		basicTxt := text.New(pixel.V(0, height-20), pg.Atlas)

		if win.JustPressed(pixelgl.KeyEscape) {
			return
		}
		if win.JustPressed(pixelgl.KeyH) {
			// toggle help screen with all the commands
		}
		// Left click has drag and drop behavior
		if win.JustPressed(pixelgl.MouseButtonLeft) {
			owner := pg.BelongsTo(x, y)
			grabbedX = x
			grabbedY = y
			if owner == pg.Root {
				pg.Highlighted = nil
				if len(pg.Root.Children) == 0 {
					owner = page.NewBubble(x, y, "", page.WHITE)
					pg.Root.Insert(owner)
					pg.Grab(owner, x, y)
				}
			} else {
				pg.Grab(owner, x, y)
			}
		}

		// Yank bubbles out of their parents (if change in velocity is sufficiently high)
		if pg.Grabbed != nil && pg.Grabbed.Parent != nil && pg.Grabbed.Parent != pg.Root {
			if pg.Grabbed.VX*pg.Grabbed.VX+pg.Grabbed.VY*pg.Grabbed.VY > 400 {
				pg.Grabbed.Parent.Detach(pg.Grabbed)
			}
		}

		// The drag part of drag-and-drop behavior for a grabbed bubble
		if pg.Grabbed != nil && grabbedX != 0 && grabbedY != 0 {
			dx := int(win.MousePreviousPosition().X - win.MousePosition().X)
			dy := int(win.MousePosition().Y - win.MousePreviousPosition().Y)
			pg.Grabbed.MoveBy(dx, dy)
		}

		// Print sidebar info
		basicTxt.Color = color.White
		b := pg.BelongsTo(x, y)
		fmt.Fprintln(basicTxt, pg.Mode+" mode:")
		fmt.Fprintln(basicTxt, x, y, b.Kind, b.Variable)
		fmt.Fprintln(basicTxt, pg.Root.Sprint())
		basicTxt.Draw(win, pixel.IM.Scaled(basicTxt.Orig, 2))

		win.SetTitle(pg.Root.String() + " | Mode: " + pg.Mode)

		switch pg.Mode {
		case "Create":
			// New bubbles with variable names are created when text is typed
			if str := win.Typed(); strings.TrimSpace(str) != "" || win.JustPressed(pixelgl.KeySpace) {
				if len(pg.Highlighted) == 1 && !pg.IsHighlighted(pg.Root) {
					highlighted := pg.Highlighted[0]
					if highlighted.Variable == "" {
						highlighted.Insert(page.NewBubble(grabbedX, grabbedY, str, highlighted.Kind))
					} else {
						highlighted.Variable += str
					}
					pg.NormalizeHeight()
					continue // don't try to interpret letters typed as variables as commands
				}
			}

			if win.JustReleased(pixelgl.MouseButtonLeft) {
				owner := pg.NearestAlternative(x, y)
				pg.ReleaseInto(owner)
			}

			// Right click creates new multiplicative units
			if win.JustPressed(pixelgl.MouseButtonRight) {
				// Insert a new bubble
				owner := pg.BelongsTo(x, y)
				pg.Grab(page.NewBubble(x, y, "", page.OppositeKind(owner.Kind)), x, y)
				pg.ReleaseInto(owner)
			}

			if win.JustReleased(pixelgl.MouseButtonRight) {
				pg.Grabbed = nil
				pg.Highlighted = nil
			}

			if win.JustPressed(pixelgl.KeyTab) {
				// insert a loop around highlighted bubbles
				if len(pg.Highlighted) > 0 && pg.Grabbed == nil {
					// check that the highlighted bubbles all have the same parent
					parent := pg.Highlighted[0].Parent
					for _, highlighted := range pg.Highlighted {
						if highlighted.Parent != parent {
							continue // skip to next frame
						}
					}

					var innerLoop *page.Bubble
					if len(pg.Highlighted) > 1 {
						loopKind := parent.Kind
						innerLoop = page.NewBubble(parent.X, parent.Y, "", loopKind)
					}

					loopKind := page.OppositeKind(parent.Kind)
					if len(pg.Highlighted) == 1 {
						loopKind = page.OppositeKind(pg.Highlighted[0].Kind)
					}
					outerLoop := page.NewBubble(parent.X, parent.Y, "", loopKind)
					parent.Insert(outerLoop)
					if innerLoop != nil {
						outerLoop.Insert(innerLoop)
						for _, highlighted := range pg.Highlighted {
							parent.Detach(highlighted)
							innerLoop.Insert(highlighted)
						}
					} else {
						for _, highlighted := range pg.Highlighted {
							parent.Detach(highlighted)
							outerLoop.Insert(highlighted)
						}
					}

					outerLoop.CenterAroundChildren()
					pg.Highlighted = []*page.Bubble{outerLoop}
					pg.NormalizeHeight()
				}
			}

			if win.JustPressed(pixelgl.KeyBackspace) || win.JustPressed(pixelgl.KeyDelete) {
				// delete a bubble in create mode
				if pg.Grabbed == nil {
					for _, highlighted := range pg.Highlighted {
						highlighted.Parent.Detach(highlighted)
					}
				}
				pg.NormalizeHeight()
			}

			if win.JustPressed(pixelgl.KeyEnter) {
				pg.Mode = "Proof"
			}
		case "Proof":
			// Allow user to insert multiplicative units at will
			if win.JustPressed(pixelgl.KeySpace) {
				if len(pg.Highlighted) == 1 && !pg.IsHighlighted(pg.Root) {
					highlighted := pg.Highlighted[0]
					highlighted.Insert(page.NewBubble(grabbedX, grabbedY, "", highlighted.Kind))
				}
			}
			if win.JustPressed(pixelgl.KeyBackspace) || win.JustPressed(pixelgl.KeyDelete) {
				// delete a loop in proof mode
				if pg.Grabbed == nil {
					for _, highlighted := range pg.Highlighted {
						if len(highlighted.Children) == 1 && highlighted.Variable == "" {
							child := highlighted.Children[0]
							newParent := highlighted.Parent
							newParent.Detach(highlighted)
							newParent.Insert(child)
						}
						// allow deletion of empty bubbles with a parent of the same color
						if len(highlighted.Children) == 0 && highlighted.Variable == "" {
							if highlighted.Parent != nil && highlighted.Parent.Kind == highlighted.Kind {
								highlighted.Parent.Detach(highlighted)
							}
						}
					}
					pg.NormalizeHeight()
				}
			}
			// Place grabbed bubble to new location, if possible
			if win.JustReleased(pixelgl.MouseButtonLeft) {
				owner := pg.NearestAlternative(x, y)

				// if this is logically allowed, then do the required operations
				if pg.CanPlaceAt(owner) {
					pg.ReleaseInto(owner)
				} else {
					// otherwise, just give it back to its original parent
					pg.ReleaseInto(pg.GrabbedParent)
				}
			}

			if str := win.Typed(); strings.TrimSpace(str) != "" || win.JustPressed(pixelgl.KeySpace) {
				if pg.InAssumptionMode() {
					if pg.AssumptionPair.Positive.Variable == "" {
						pg.AssumptionPair.Positive.Insert(page.NewBubble(pg.AssumptionPair.Positive.X, pg.AssumptionPair.Positive.Y, str, page.WHITE))
						pg.AssumptionPair.Negative.Insert(page.NewBubble(pg.AssumptionPair.Negative.X, pg.AssumptionPair.Negative.Y, str, page.BLACK))
					} else {
						pg.AssumptionPair.Positive.Variable += str
						pg.AssumptionPair.Negative.Variable += str
					}
					continue // don't try to interpret letters typed as variables as commands
				}
			}

			if win.JustPressed(pixelgl.MouseButtonRight) {
				pg.AssumptionPair = nil
				// Right click grabs things from "the void"
				owner := pg.BelongsTo(x, y)
				if owner.Kind == page.BLACK {
					pg.AssumptionPair = &page.Pair{Negative: owner}
					pg.GrabbedAtX, pg.GrabbedAtY = x, y
				} else if owner.Kind == page.WHITE {
					pg.AssumptionPair = &page.Pair{Positive: owner}
					pg.GrabbedAtX, pg.GrabbedAtY = x, y
				}
			}

			if win.JustReleased(pixelgl.MouseButtonRight) {
				owner := pg.BelongsTo(x, y)
				if owner.Kind == page.BLACK {
					if pg.AssumptionPair.Positive != nil && pg.AssumptionPair.Positive.Parent == owner {
						// create new bubbles for assumption pair
						pg.AssumptionPair.Negative = owner
						newPositive := page.NewBubble(pg.GrabbedAtX, pg.GrabbedAtY, "", page.WHITE)
						pg.AssumptionPair.Positive.Insert(newPositive)
						pg.AssumptionPair.Positive = newPositive
						newNegative := page.NewBubble(x, y, "", page.BLACK)
						pg.AssumptionPair.Negative.Insert(newNegative)
						pg.AssumptionPair.Negative = newNegative
					}
				}
				if owner.Kind == page.WHITE {
					if pg.AssumptionPair.Negative != nil && pg.AssumptionPair.Negative == owner.Parent {
						// create new bubbles for assumption pair
						pg.AssumptionPair.Positive = owner
						newPositive := page.NewBubble(x, y, "", page.WHITE)
						pg.AssumptionPair.Positive.Insert(newPositive)
						pg.AssumptionPair.Positive = newPositive
						newNegative := page.NewBubble(pg.GrabbedAtX, pg.GrabbedAtY, "", page.BLACK)
						pg.AssumptionPair.Negative.Insert(newNegative)
						pg.AssumptionPair.Negative = newNegative
					}
				}
				pg.Grabbed = nil
				pg.Highlighted = nil
			}

			if win.JustPressed(pixelgl.KeyTab) {
				// insert a loop around highlighted bubbles
				if len(pg.Highlighted) > 0 && pg.Grabbed == nil {
					// check that the highlighted bubbles all have the same parent
					parent := pg.Highlighted[0].Parent
					for _, highlighted := range pg.Highlighted {
						if highlighted.Parent != parent {
							continue // skip to next frame
						}
					}

					var innerLoop *page.Bubble
					if len(pg.Highlighted) > 1 {
						loopKind := parent.Kind
						innerLoop = page.NewBubble(parent.X, parent.Y, "", loopKind)
					}

					loopKind := page.OppositeKind(parent.Kind)
					if len(pg.Highlighted) == 1 {
						loopKind = page.OppositeKind(pg.Highlighted[0].Kind)
					}
					outerLoop := page.NewBubble(parent.X, parent.Y, "", loopKind)
					parent.Insert(outerLoop)
					if innerLoop != nil {
						outerLoop.Insert(innerLoop)
						for _, highlighted := range pg.Highlighted {
							parent.Detach(highlighted)
							innerLoop.Insert(highlighted)
						}
					} else {
						for _, highlighted := range pg.Highlighted {
							parent.Detach(highlighted)
							outerLoop.Insert(highlighted)
						}
					}

					outerLoop.CenterAroundChildren()
					pg.Highlighted = []*page.Bubble{outerLoop}
					pg.NormalizeHeight()
				}
			}
		}
	}
}

func main() {
	pixelgl.Run(run)
}
