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

// how should exponentials work?
// we could have a new type of bubble: a blue (or red) bubble
// these bubbles could only be single loops
// that doesn't seem so hard
// characters ? and ! could work like tab for creating them
//   note that the assumption pair code would need to be more complete for this to work

// okay, how about additives?
// and additive bubble consists of an outershell, along with a
// I think it will need to be radial and

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
			for j := 0; j < len(b.Children); j++ {
				b.Children[j].Iterate(func(nibling *page.Bubble) {
					if page.Distance(b.Children[i], nibling) < float64(30.0*(b.Children[i].Height+nibling.Height)+85) &&
						b.Children[i] != pg.Grabbed && nibling != pg.Grabbed && i != j {
						dx := 0
						dy := 0
						for dx*dy == 0 {
							dx = int(2*math.Atan(float64(b.Children[i].X-nibling.X))) + page.Random(-2, 2)
							dy = int(2*math.Atan(float64(b.Children[i].Y-nibling.Y))) + page.Random(-2, 2)
						}
						b.Children[i].X += dx
						b.Children[i].Y += dy
						nibling.X -= dx
						nibling.Y -= dy
					}
				})
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
					owner = pg.NewBubble(x, y, "", page.WHITE)
					pg.Root.Insert(owner)
					pg.Grab(owner, x, y)
				}
			} else {
				if win.JustPressed(pixelgl.KeyLeftShift) || win.JustPressed(pixelgl.KeyLeftShift) {
					pg.Highlighted = append(pg.Highlighted, owner)
				} else {
					pg.Grab(owner, x, y)
				}
			}
		}

		// Yank bubbles out of their parents (if change in velocity is sufficiently high)
		if pg.Grabbed != nil && pg.Grabbed.Parent != nil && pg.Grabbed.Parent != pg.Root {
			if !pg.AssumptionMode || pg.Grabbed.AssumptionPair != nil {
				if pg.Grabbed.VX*pg.Grabbed.VX+pg.Grabbed.VY*pg.Grabbed.VY > 400 {
					pg.Execute(func() { pg.Delete(pg.Grabbed) })
				}
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
		fmt.Fprintln(basicTxt, x, y, page.Name(b.Kind), b.Variable)
		fmt.Fprintln(basicTxt)
		fmt.Fprintln(basicTxt, pg.Root.Sprint())
		fmt.Fprintln(basicTxt, "Assumption Mode:\n", pg.AssumptionMode)
		basicTxt.Draw(win, pixel.IM.Scaled(basicTxt.Orig, 2))

		win.SetTitle(pg.Root.String() + " | Mode: " + pg.Mode)

		switch pg.Mode {
		case "Create":
			// New bubbles with variable names are created when text is typed
			if str := win.Typed(); strings.TrimSpace(str) != "" || win.JustPressed(pixelgl.KeySpace) {
				if len(pg.Highlighted) == 1 && !pg.IsHighlighted(pg.Root) {
					highlighted := pg.Highlighted[0]
					if highlighted.Variable == "" {
						highlighted.Insert(pg.NewBubble(grabbedX, grabbedY, str, highlighted.Kind))
					} else {
						highlighted.Variable += str
					}
					pg.NormalizeHeight()
					continue // don't try to interpret letters typed as variables as commands
				}
			}

			if win.JustReleased(pixelgl.MouseButtonLeft) {
				owner := pg.NearestAlternative(x, y)
				pg.Execute(func() { pg.ReleaseInto(owner) })
			}

			// Right click creates new multiplicative units
			if win.JustPressed(pixelgl.MouseButtonRight) {
				// Insert a new bubble
				owner := pg.BelongsTo(x, y)
				pg.Grab(pg.NewBubble(x, y, "", owner.OppositeKind()), x, y)
				pg.Execute(func() { pg.ReleaseInto(owner) })
			}

			if win.JustReleased(pixelgl.MouseButtonRight) {
				pg.Grabbed = nil
				pg.Highlighted = nil
			}

			if win.JustPressed(pixelgl.KeyTab) {
				// insert a loop around highlighted bubbles
				if len(pg.Highlighted) > 0 && pg.Grabbed == nil {
					parent := pg.Highlighted[0].Parent
					loopKind := parent.OppositeKind()
					if len(pg.Highlighted) == 1 {
						loopKind = pg.Highlighted[0].OppositeKind()
					}
					pg.Execute(func() { pg.Loop(loopKind, pg.Highlighted...) })
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
					pg.Execute(func() {
						highlighted := pg.Highlighted[0]
						highlighted.Insert(pg.NewBubble(grabbedX, grabbedY, "", highlighted.Kind))
					})
				}
			}
			if win.JustPressed(pixelgl.KeyBackspace) || win.JustPressed(pixelgl.KeyDelete) {
				// delete a loop in proof mode
				if pg.Grabbed == nil {
					pg.Execute(func() {
						for _, highlighted := range pg.Highlighted {
							if pg.AssumptionPair != nil && (highlighted == pg.AssumptionPair.Positive || highlighted == pg.AssumptionPair.Negative) {
								return
							}
							if len(highlighted.Children) == 1 && highlighted.Variable == "" && highlighted.Parent != nil {
								child := highlighted.Children[0]
								newParent := highlighted.Parent
								pg.Delete(highlighted)
								pg.Place(newParent, child)
							}
							// allow deletion of empty bubbles with a parent of the same color
							if len(highlighted.Children) == 0 && highlighted.Variable == "" {
								if highlighted.Parent != nil && highlighted.Parent.Kind == highlighted.Kind {
									pg.Delete(highlighted)
								}
							}
						}
					})
				}
			}
			// Place grabbed bubble to new location, if possible
			if win.JustReleased(pixelgl.MouseButtonLeft) {
				owner := pg.NearestAlternative(x, y)
				// if this is logically allowed, then do the required operations
				if pg.CanPlaceAt(owner) {
					fmt.Println("can place")
					pg.Execute(func() { pg.ReleaseInto(owner) })
				} else {
					// otherwise, just give it back to its original parent
					fmt.Println("not allowed")
					pg.Execute(func() { pg.ReleaseInto(pg.GrabbedParent) })
				}
				pg.Grabbed = nil
			}

			if str := win.Typed(); strings.TrimSpace(str) != "" || win.JustPressed(pixelgl.KeySpace) {
				if pg.AssumptionMode && len(pg.Highlighted) == 1 {
					subject := pg.Highlighted[0]
					if subject.Variable == "" {
						pg.Execute(func() {
							pg.Grab(pg.NewBubble(subject.X, subject.Y, str, subject.Kind), subject.X, subject.Y)
							pg.ReleaseInto(subject)
						})
					} else {
						pg.Execute(func() {
							subject.Variable += str
							if subject.AssumptionPair != nil {
								subject.AssumptionPair.Variable += str
							}
						})
					}
					continue // don't try to interpret letters typed as variables as commands
				}
			}

			if win.JustPressed(pixelgl.MouseButtonRight) {
				owner := pg.BelongsTo(x, y)
				if !pg.AssumptionMode && pg.AssumptionPair == nil {
					// Right click grabs things from "the void"
					if owner.Kind == page.BLACK {
						pg.AssumptionPair = &page.Pair{Negative: owner}
						pg.GrabbedAtX, pg.GrabbedAtY = x, y
					} else if owner.Kind == page.WHITE {
						pg.AssumptionPair = &page.Pair{Positive: owner}
						pg.GrabbedAtX, pg.GrabbedAtY = x, y
					}
				} else if pg.InAssumption(owner) {
					pg.Execute(func() {
						newb := pg.NewBubble(x, y, "", owner.OppositeKind())
						pg.Grab(newb, newb.X, newb.Y)
						pg.ReleaseInto(owner)
					})
				} else {
					pg.AssumptionPair = nil
					pg.AssumptionMode = false
				}
			}

			if win.JustReleased(pixelgl.MouseButtonRight) {
				owner := pg.BelongsTo(x, y)
				if !pg.AssumptionMode && pg.AssumptionPair != nil {
					if owner.Kind == page.BLACK {
						if pg.AssumptionPair.Positive != nil && pg.AssumptionPair.Positive.Parent == owner {
							// create new bubbles for assumption pair
							pg.AssumptionPair.Negative = owner
							newPositive := pg.NewBubble(pg.GrabbedAtX, pg.GrabbedAtY, "", page.WHITE)
							pg.AssumptionPair.Positive.Insert(newPositive)
							pg.AssumptionPair.Positive = newPositive
							newNegative := pg.NewBubble(x, y, "", page.BLACK)
							pg.AssumptionPair.Negative.Insert(newNegative)
							pg.AssumptionPair.Negative = newNegative
							pg.AssumptionPair.Positive.AssumptionPair = pg.AssumptionPair.Negative
							pg.AssumptionPair.Negative.AssumptionPair = pg.AssumptionPair.Positive
							pg.AssumptionMode = true

						}
					}
					if owner.Kind == page.WHITE {
						if pg.AssumptionPair.Negative != nil && pg.AssumptionPair.Negative == owner.Parent {
							// create new bubbles for assumption pair
							pg.AssumptionPair.Positive = owner
							newPositive := pg.NewBubble(x, y, "", page.WHITE)
							pg.AssumptionPair.Positive.Insert(newPositive)
							pg.AssumptionPair.Positive = newPositive
							newNegative := pg.NewBubble(pg.GrabbedAtX, pg.GrabbedAtY, "", page.BLACK)
							pg.AssumptionPair.Negative.Insert(newNegative)
							pg.AssumptionPair.Negative = newNegative
							pg.AssumptionPair.Positive.AssumptionPair = pg.AssumptionPair.Negative
							pg.AssumptionPair.Negative.AssumptionPair = pg.AssumptionPair.Positive
							pg.AssumptionMode = true
						}
					}
				}
				pg.Grabbed = nil
				pg.Highlighted = nil
			}

			if win.JustPressed(pixelgl.KeyTab) {
				// insert a loop around highlighted bubbles
				if len(pg.Highlighted) > 0 && pg.Grabbed == nil {
					subject := pg.Highlighted[0]
					if subject != pg.AssumptionPair.Positive && subject != pg.AssumptionPair.Negative {
						loopKind := subject.Parent.OppositeKind()
						if len(pg.Highlighted) == 1 {
							loopKind = subject.OppositeKind()
						}
						pg.Execute(func() { pg.Loop(loopKind, pg.Highlighted...) })
					}
				}
			}
		}
	}
}

func main() {
	pixelgl.Run(run)
}
