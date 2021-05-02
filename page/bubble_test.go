package page

import (
	"testing"

	"gotest.tools/assert"
)

func TestHeights(t *testing.T) {
	a := newBubble(0, 0, "A", WHITE)
	b := newBubble(0, 0, "B", BLACK)
	c := newBubble(0, 0, "C", WHITE)
	d := newBubble(0, 0, "D", WHITE)
	a.Insert(b)
	b.Insert(c)
	assert.Equal(t, a.Height, 2)
	assert.Equal(t, b.Height, 1)
	b.Detach(c)
	assert.Equal(t, a.Height, 1)
	assert.Equal(t, b.Height, 0)
	b.Insert(c)
	b.Insert(d)
	assert.Equal(t, a.Height, 2)
	assert.Equal(t, b.Height, 1)
	b.Detach(c)
	assert.Equal(t, a.Height, 2)
	assert.Equal(t, b.Height, 1)
}
