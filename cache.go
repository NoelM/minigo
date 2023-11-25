package minigo

const CacheSize = 16

type Cache struct {
	maxSize   int
	curId     int
	container [][]byte
}

func NewCache() *Cache {
	return &Cache{
		curId:     -1,
		container: make([][]byte, CacheSize),
	}
}

func (c *Cache) Reset() {
	c.curId = -1
}

func (c *Cache) Add(msg []byte) {
	if c.curId < 0 {
		c.curId = 0
	}

	c.container[c.curId] = make([]byte, len(msg))
	copy(c.container[c.curId], msg)

	c.curId += 1
	if c.curId == c.maxSize {
		c.curId = 0
	}
}

func (c *Cache) Last() []byte {
	return c.container[CacheSize-1]
}

func (c *Cache) Get(id byte) []byte {
	return c.container[id]
}

func (c *Cache) Empty() bool {
	return c.curId < 0
}
