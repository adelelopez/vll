package page

import (
	"container/list"
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"sort"
	"strings"
)

const (
	seed        = 123
	width       = 1024
	height      = 720
	sidebar     = 225
	circSquared = 900.0
	alphabet    = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

var (
	WHITE = Kind(color.White)
	BLACK = Kind(color.Black)
	BLUE  = Kind(color.RGBA{
		17,
		205,
		205,
		255,
	})
	RED = Kind(color.RGBA{
		238,
		50,
		50,
		255,
	})
	BACKGROUND = Kind(color.RGBA{
		151,
		151,
		184,
		255,
	})
)

type Kind color.Color

type Bubble struct {
	X              int
	Y              int
	VX             int
	VY             int
	Depth          int // leaves have depth 0
	Height         int
	Kind           Kind
	Variable       string
	Children       []*Bubble
	Parent         *Bubble
	AssumptionPair *Bubble
}

type ByDepth []*Bubble

func (a ByDepth) Len() int           { return len(a) }
func (a ByDepth) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByDepth) Less(i, j int) bool { return a[i].Depth < a[j].Depth }

func init() {
	rand.Seed(seed)
}

func (b *Bubble) OppositeKind() Kind {
	switch b.Kind {
	case BLACK:
		return WHITE
	case WHITE, BACKGROUND:
		return BLACK
	case RED:
		return BLUE
	case BLUE:
		return RED
	default:
		return BACKGROUND
	}
}

func (b *Bubble) IsMult() bool {
	return b.Kind == BLACK || b.Kind == WHITE
}

func (b *Bubble) OppositePolarity() Kind {
	switch b.Kind {
	case BLACK, RED:
		return WHITE
	case WHITE, BLUE, BACKGROUND:
		return BLACK

	default:
		return BACKGROUND
	}
}

// TODO: use more efficient algorithm, segmented tree is probably a good fit
func LCA(bubs ...*Bubble) *Bubble {
	if bubs == nil {
		return nil
	}
	if len(bubs) == 1 {
		return bubs[0]
	}
	if len(bubs) == 2 {
		var newb *Bubble
		if bubs[0].IsAbove(bubs[1]) {
			return bubs[0]
		} else if bubs[1].IsAbove(bubs[0]) {
			return bubs[1]
		} else if bubs[0].Depth < bubs[1].Depth {
			newb = bubs[0].Parent
			if newb == nil {
				return nil
			}
			return LCA(newb, bubs[1])
		} else {
			newb = bubs[1].Parent
			if newb == nil {
				return nil
			}
			return LCA(bubs[0], newb)
		}
	}
	return LCA(LCA(bubs[0:len(bubs)/2]...), LCA(bubs[len(bubs)/2:]...))
}

func (b *Bubble) Copy() *Bubble {
	// create a new bubble
	newb := newBubble(b.X, b.Y, b.Variable, b.Kind)
	// for each child of the original, create a copy of the bubble and append it
	for _, child := range b.Children {
		twin := child.Copy()
		newb.Insert(twin)
	}
	return newb
}

func Random(min, max int) int {
	return min + rand.Intn(max-min)
}

func Name(k Kind) string {
	switch k {
	case WHITE:
		return "White"
	case BLACK:
		return "Black"
	case BLUE:
		return "Blue"
	case RED:
		return "Red"
	case BACKGROUND:
		return "Root"
	default:
		return "Unknown"
	}
}

func newBubble(x, y int, v string, k Kind) *Bubble {
	return &Bubble{
		X:        x,
		Y:        y,
		Kind:     k,
		Variable: v,
	}
}

func (b *Bubble) Iterate(f func(*Bubble)) {
	if b != nil {
		for _, child := range b.Children {
			child.Iterate(f)
		}
	}
	f(b)
}

func (b *Bubble) Sprint() string {
	s := "Current tree:\n"
	b.bfs(func(bub *Bubble) {
		for i := 0; i < bub.Depth; i++ {
			s += "."
		}
		if bub.Variable != "" {
			s += bub.Variable
		} else {
			s += Name(bub.Kind)
		}
		s += fmt.Sprintf("(%v)", bub.Height)
		s += "\n"
	})
	return s
}

func (b *Bubble) Tolestra() string {
	if len(b.Children) == 0 {
		if b.Kind == WHITE {
			if b.Variable == "" {
				return "1"
			}
			return b.Variable
		}
		if b.Kind == BLACK {
			if b.Variable == "" {
				return "0"
			}
			return "~" + b.Variable
		}
	}

	childrenStrings := make([]string, 0, len(b.Children))

	for _, child := range b.Children {
		childrenStrings = append(childrenStrings, child.Tolestra())
	}
	sort.Sort(sort.StringSlice(childrenStrings))

	str := ""
	if len(b.Children) <= 1 {
		str = strings.Join(childrenStrings, "")
		if b.Kind == BLUE {
			str = "!" + strings.Join(childrenStrings, "")
		}
		if b.Kind == RED {
			str = "?" + strings.Join(childrenStrings, "")
		}
		return str
	}
	switch b.Kind {
	case WHITE:
		str = "(" + strings.Join(childrenStrings, " * ") + ")"
	case BLACK:
		str = "(" + strings.Join(childrenStrings, " + ") + ")"
	case BLUE:
		str = "!(" + strings.Join(childrenStrings, " * ") + ")"
	case RED:
		str = "?(" + strings.Join(childrenStrings, " + ") + ")"
	}
	return str
}

func (b *Bubble) Opposite() string {
	if len(b.Children) == 0 {
		if b.Kind == BLACK {
			if b.Variable == "" {
				return "1"
			}
			return b.Variable
		}
		if b.Kind == WHITE {
			if b.Variable == "" {
				return "0"
			}
			return "~" + b.Variable
		}
	}

	childrenStrings := make([]string, 0, len(b.Children))

	for _, child := range b.Children {
		childrenStrings = append(childrenStrings, child.Opposite())
	}
	sort.Sort(sort.StringSlice(childrenStrings))

	str := ""
	if len(b.Children) <= 1 {
		str = strings.Join(childrenStrings, "")
		if b.Kind == BLUE {
			str = "?" + strings.Join(childrenStrings, "")
		}
		if b.Kind == RED {
			str = "!" + strings.Join(childrenStrings, "")
		}
		return str
	}

	switch b.Kind {
	case WHITE:
		str = "(" + strings.Join(childrenStrings, " + ") + ")"
	case BLACK:
		str = "(" + strings.Join(childrenStrings, " * ") + ")"
	case BLUE:
		str = "?(" + strings.Join(childrenStrings, " + ") + ")"
	case RED:
		str = "!(" + strings.Join(childrenStrings, " * ") + ")"
	}
	return str
}

func (b *Bubble) bfs(f func(*Bubble)) {
	queue := list.New()
	queue.PushFront(b)
	for queue.Len() > 0 {
		nextBub := queue.Remove(queue.Front()).(*Bubble)
		f(nextBub)
		for _, child := range nextBub.Children {
			queue.PushFront(child)
		}
	}
}

func (b *Bubble) Insert(child *Bubble) *Bubble {
	if b == nil {
		return nil
	}
	if child.IsAbove(b) {
		return nil
		panic("cyclic tree formed!")
	}
	for _, kiddo := range b.Children {
		if kiddo == child {
			return child
		}
	}
	child.Depth = b.Depth + 1
	child.Parent = b
	b.Children = append(b.Children, child)

	return child
}

func (b *Bubble) normalizeHeight() {
	if len(b.Children) == 0 {
		b.Height = 0
		return
	}

	height := 0
	for _, child := range b.Children {
		if child.Height > height {
			height = child.Height
		}
	}
	b.Height = height + 1
}

func (b *Bubble) Detach(child *Bubble) {
	fmt.Println("detaching")
	if b == nil {
		return
	}
	b.Iterate(func(b *Bubble) {
		for i := 0; i < len(b.Children); i++ {
			if b.Children[i] == child {
				// fast delete
				b.Children[len(b.Children)-1], b.Children[i] = b.Children[i], b.Children[len(b.Children)-1]
				b.Children = b.Children[:len(b.Children)-1]

				child.Parent = nil
				return
			}
		}
	})
}

func (b *Bubble) CenterOfMass() (x int, y int) {
	n := 0
	b.Iterate(func(bub *Bubble) {
		x += bub.X
		y += bub.Y
		n++
	})
	x /= n
	y /= n
	return
}

func (b *Bubble) CenterAroundChildren() {
	var x, y int
	n := 0
	b.Iterate(func(bub *Bubble) {
		if bub != b {
			x += bub.X
			y += bub.Y
			n++
		}
	})
	x /= n
	y /= n
	b.X, b.Y = x, y
	return
}

func Distance(a, b *Bubble) float64 {
	xDist := a.X - b.X
	yDist := a.Y - b.Y
	return math.Sqrt(float64(xDist*xDist + yDist*yDist))
}

func (b *Bubble) MoveBy(dx, dy int) {
	b.Iterate(func(bub *Bubble) {
		bub.X -= dx
		bub.Y -= dy
	})
	b.VX = -dx
	b.VY = -dy

	ancestor := b.Parent
	for ancestor != nil && ancestor.Depth != 0 {
		ancestor.X, ancestor.Y = ancestor.CenterOfMass()
		ancestor = ancestor.Parent
	}
}

func (b *Bubble) Siblings() []*Bubble {
	if b.Parent == nil {
		return nil
	}
	siblings := make([]*Bubble, 0, len(b.Parent.Children)-1)
	for _, child := range b.Parent.Children {
		// you're not your own sibling
		if child != b {
			siblings = append(siblings, child)
		}
	}
	return siblings
}

func (b *Bubble) IsAbove(other *Bubble) bool {
	if other == nil {
		return false
	}
	ancestor := other

	for ancestor != nil {
		if ancestor == b {
			return true
		}
		ancestor = ancestor.Parent
	}
	return false
}
