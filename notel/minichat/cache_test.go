package minichat

import "testing"

func TestBottom(t *testing.T) {
	const value = 5

	c := NewCache()
	c.Bottom(value)

	if c.lines[len(c.lines)-1] != value {
		t.Fatal("values incoherent")
	}
}

func TestTop(t *testing.T) {
	const value = 5

	c := NewCache()
	c.Top(value)

	if c.lines[0] != value {
		t.Fatal("values incoherent")
	}
}
