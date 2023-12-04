package minichat

const nbLines = 24

type Cache struct {
	lines []int
}

func NewCache() *Cache {
	return &Cache{
		lines: make([]int, nbLines),
	}
}

func (c *Cache) Bottom(i int) {
	copy(c.lines[:nbLines-1], c.lines[1:])
	c.lines[nbLines-1] = i
}

func (c *Cache) MultBottom(i, mult int) {
	for k := 0; k < mult; k += 1 {
		c.Bottom(i)
	}
}

func (c *Cache) Top(i int) {
	copy(c.lines[1:], c.lines[:nbLines-1])
	c.lines[0] = i
}

func (c *Cache) MultTop(i, mult int) {
	for k := 0; k < mult; k += 1 {
		c.Top(i)
	}
}
