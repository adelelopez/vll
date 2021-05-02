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

func (b *Bubble) String() string {
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
		childrenStrings = append(childrenStrings, child.String())
	}
	sort.Sort(sort.StringSlice(childrenStrings))

	if len(b.Children) <= 1 {
		return strings.Join(childrenStrings, "")
	}
	if b.Kind == BLACK {
		return "(" + strings.Join(childrenStrings, " + ") + ")"
	}
	return "(" + strings.Join(childrenStrings, " * ") + ")"
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
		childrenStrings = append(childrenStrings, child.String())
	}
	sort.Sort(sort.StringSlice(childrenStrings))

	if len(b.Children) <= 1 {
		return strings.Join(childrenStrings, "")
	}
	if b.Kind == WHITE {
		return "(" + strings.Join(childrenStrings, " + ") + ")"
	}
	return "(" + strings.Join(childrenStrings, " * ") + ")"
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
	ancestor := other

	for ancestor != nil {
		if ancestor == b {
			return true
		}
		ancestor = ancestor.Parent
	}
	return false
}
