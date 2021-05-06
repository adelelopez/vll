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
// if blue ring and something inside it are highlighted, delete will delete the whole thing
// should red rings just be freely editable inside? -- definitely not, that's logically incorrect
// ? creates a red loop around things, which always works
// if it's around a unit, you enter contingency mode, which lets you freely edit the interior until you exit it
// ! requires copying
// i think it makes sense to implement copying with double-clicks, a double click in a blue loop copies the contents, and it is grabbed on the second click
// maybe it also makes sense to implement ctrl-c ctrl-v copy-paste controls too -- this would also allow for copying multiple bubbles at once

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
	clickTime := time.Now()
	var clickOwner *page.Bubble

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

		// Yank bubbles out of their parents (if change in velocity is sufficiently high)
		if pg.Grabbed != nil && pg.Grabbed.Parent != nil && pg.Grabbed.Parent != pg.Root {
			if !pg.AssumptionMode || pg.Grabbed.AssumptionPair != nil && pg.Grabbed.Parent.Kind != page.RED && pg.Grabbed.Parent.Kind != page.BLUE {
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
				switch str {
				case "!":
					if len(pg.Highlighted) > 0 && pg.Grabbed == nil {
						subject := pg.Highlighted[0]
						if pg.AssumptionPair == nil || (subject != pg.AssumptionPair.Positive && subject != pg.AssumptionPair.Negative) {
							loopKind := page.BLUE
							pg.Execute(func() { pg.Loop(loopKind, pg.Highlighted...) })
						}
					}
				case "?":
					if len(pg.Highlighted) > 0 && pg.Grabbed == nil {
						// is it confusing if this doesn't put a loop around things?
						subject := pg.Highlighted[0]
						if subject.Kind != page.BLUE && subject.Kind != page.RED {
							pg.Execute(func() {
								newb := pg.NewBubble(subject.X, subject.Y, "", subject.Kind)
								pg.Grab(newb, subject.X, subject.Y)
								pg.ReleaseInto(subject)
								pg.Loop(page.RED, newb)
							})
						}
					}
				default:
					if len(pg.Highlighted) == 1 && !pg.IsHighlighted(pg.Root) {
						highlighted := pg.Highlighted[0]
						if highlighted.Kind != page.BLUE && highlighted.Kind != page.RED {
							pg.Execute(func() {
								if highlighted.Variable == "" {
									highlighted.Insert(pg.NewBubble(grabbedX, grabbedY, str, highlighted.Kind))
								} else {
									highlighted.Variable += str
								}
							})
							continue // don't try to interpret letters typed as variables as commands
						}
					}
				}
			}
			// Left click has drag and drop behavior
			if win.JustPressed(pixelgl.MouseButtonLeft) {
				owner := pg.BelongsTo(x, y)
				grabbedX = x
				grabbedY = y
				if owner == clickOwner && time.Now().Sub(clickTime) < time.Duration(350*time.Millisecond) {
					fmt.Println("doubleclick")
					newb := owner.Copy()
					pg.Place(owner.Parent, newb)
					pg.Grab(newb, x, y)
				} else {
					clickTime = time.Now()
					clickOwner = owner
					pg.Grab(owner, x, y)
				}
				if owner == pg.Root {
					pg.Highlighted = nil
					if len(pg.Root.Children) == 0 {
						owner = pg.NewBubble(x, y, "", page.WHITE)
						pg.Root.Insert(owner)
						pg.Grab(owner, x, y)
					}
				} else {
					if win.Pressed(pixelgl.KeyLeftShift) || win.Pressed(pixelgl.KeyLeftShift) {
						fmt.Println("shifty")
						pg.Highlighted = append(pg.Highlighted, owner)
					} else {
						pg.Grab(owner, x, y)
					}
				}

			}

			if win.JustReleased(pixelgl.MouseButtonLeft) {
				owner := pg.NearestAlternative(x, y)
				if owner.Kind != page.RED && owner.Kind != page.BLUE {
					pg.Execute(func() { pg.ReleaseInto(owner) })
				}
				pg.Grabbed = nil
			}

			// Right click creates new multiplicative units
			if win.JustPressed(pixelgl.MouseButtonRight) {
				// Insert a new bubble
				owner := pg.BelongsTo(x, y)
				if owner.Kind != page.RED && owner.Kind != page.BLUE {
					pg.Grab(pg.NewBubble(x, y, "", owner.OppositePolarity()), x, y)
					pg.Execute(func() { pg.ReleaseInto(owner) })
				}
			}

			if win.JustReleased(pixelgl.MouseButtonRight) {
				pg.Grabbed = nil
				pg.Highlighted = nil
			}

			if win.JustPressed(pixelgl.KeyTab) {
				// insert a loop around highlighted bubbles
				if len(pg.Highlighted) > 0 && pg.Grabbed == nil {
					parent := pg.Highlighted[0].Parent
					loopKind := parent.OppositePolarity()
					if len(pg.Highlighted) == 1 {
						loopKind = pg.Highlighted[0].OppositePolarity()
					} else if parent == pg.Root {
						continue
					}
					pg.Execute(func() { pg.Loop(loopKind, pg.Highlighted...) })
				}
			}

			if win.JustPressed(pixelgl.KeyBackspace) || win.JustPressed(pixelgl.KeyDelete) {
				// delete a bubble in create mode
				// delete a loop in proof mode
				if pg.Grabbed == nil {
					pg.Execute(func() {
						for _, highlighted := range pg.Highlighted {
							newParent := highlighted.Parent
							for _, child := range highlighted.Children {
								pg.Delete(highlighted)
								pg.Place(newParent, child)
							}
						}
					})
				}
			}

			if win.JustPressed(pixelgl.KeyEnter) {
				pg.Mode = "Proof"
			}
		case "Proof":
			if win.JustPressed(pixelgl.KeyBackspace) || win.JustPressed(pixelgl.KeyDelete) {
				// delete a loop in proof mode
				if pg.Grabbed == nil {
					pg.Execute(func() {
						for _, highlighted := range pg.Highlighted {
							if pg.AssumptionPair != nil && (highlighted == pg.AssumptionPair.Positive || highlighted == pg.AssumptionPair.Negative) {
								return
							}

							// TODO: if a blue loop is highlighted, along with any of its descendents, delete the entire bubble

							if len(highlighted.Children) == 1 && highlighted.Variable == "" && highlighted.Parent != nil && highlighted.Kind != page.RED {
								child := highlighted.Children[0]
								newParent := highlighted.Parent
								pg.Delete(highlighted)
								pg.Place(newParent, child)
							}

							// allow deletion of empty bubbles with a parent of the same color
							if len(highlighted.Children) == 0 && highlighted.Variable == "" {
								if highlighted.Parent != nil && highlighted.Parent.Kind == highlighted.Kind && highlighted.Kind != page.RED {
									pg.Delete(highlighted)
								}
							}
						}
						// only delete a red loop if it's child is black and its grandkids are red loops
						if len(pg.Highlighted) == 1 && pg.Highlighted[0].Kind == page.RED {
							subject := pg.Highlighted[0]
							if len(subject.Children) == 1 && subject.Children[0].Kind == page.BLACK {
								child := subject.Children[0]
								allred := true
								for _, grandkid := range child.Children {
									if grandkid.Kind != page.RED {
										allred = false
									}
								}
								if allred {
									newParent := subject.Parent
									pg.Delete(subject)
									pg.Place(newParent, child)
								}
							}
						}
					})
				}
			}
			// Left click has drag and drop behavior
			if win.JustPressed(pixelgl.MouseButtonLeft) {
				owner := pg.BelongsTo(x, y)
				grabbedX = x
				grabbedY = y

				if win.Pressed(pixelgl.KeyLeftShift) || win.Pressed(pixelgl.KeyLeftShift) {
					fmt.Println("shifty")
					pg.Highlighted = append(pg.Highlighted, owner)
				} else {
					if owner == clickOwner && time.Now().Sub(clickTime) < time.Duration(350*time.Millisecond) {
						fmt.Println("doubleclick")
						if owner.Kind == page.BLUE {
							newb := owner.Copy()
							pg.Place(owner.Parent, newb)
							pg.Grab(newb, x, y)
						}
					} else {
						clickTime = time.Now()
						clickOwner = owner
						pg.Grab(owner, x, y)
					}
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
				switch str {
				case "!":
					if len(pg.Highlighted) > 0 && pg.Grabbed == nil {
						subject := pg.Highlighted[0]
						if pg.AssumptionPair == nil || (subject != pg.AssumptionPair.Positive && subject != pg.AssumptionPair.Negative) {
							loopKind := page.BLUE
							for _, highlighted := range pg.Highlighted {
								if highlighted.Kind != page.BLUE && !(highlighted.Variable == "" &&
									len(highlighted.Children) == 0 && highlighted.Kind == page.WHITE) {
									continue
								}
							}
							pg.Execute(func() { pg.Loop(loopKind, pg.Highlighted...) })
						}
					}
				case "?":
					if len(pg.Highlighted) > 0 && pg.Grabbed == nil {
						subject := pg.Highlighted[0]
						if pg.AssumptionPair == nil || (subject != pg.AssumptionPair.Positive && subject != pg.AssumptionPair.Negative) {
							loopKind := page.RED
							for _, highlighted := range pg.Highlighted {
								if highlighted.Kind != page.BLUE && !(highlighted.Variable == "" &&
									len(highlighted.Children) == 0 && highlighted.Kind == page.WHITE) {
									continue
								}
							}
							pg.Execute(func() { pg.Loop(loopKind, pg.Highlighted...) })
							if len(pg.Highlighted) == 1 && pg.Highlighted[0].Variable == "" {
								// TODO: Enter contingency mode
							}
						}
					}
				default:
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
						if owner.Kind != page.RED && owner.Kind != page.BLUE {
							newb := pg.NewBubble(x, y, "", owner.OppositePolarity())
							pg.Grab(newb, newb.X, newb.Y)
							pg.ReleaseInto(owner)
						}
					})
				} else {
					pg.ExitAssumptionMode()
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
					if pg.AssumptionPair == nil || (subject != pg.AssumptionPair.Positive && subject != pg.AssumptionPair.Negative) {
						loopKind := subject.Parent.OppositePolarity()
						if len(pg.Highlighted) == 1 {
							loopKind = subject.OppositePolarity()
						} else if subject.Parent == pg.Root {
							continue
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
