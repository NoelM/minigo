package minigo

const CacheSize = 16

type Cache struct {
	curId     int
	maxId     int
	container [][]byte
}

func NewCache() *Cache {
	return &Cache{
		curId:     0,
		container: make([][]byte, CacheSize),
	}
}

func (c *Cache) Reset() {
	c.curId = 0
	c.maxId = -1
}

func (c *Cache) Add(msg []byte) {
	if c.curId > c.maxId {
		c.maxId = c.curId
	}

	c.container[c.curId] = make([]byte, len(msg))
	copy(c.container[c.curId], msg)

	c.curId += 1
	if c.curId == CacheSize {
		c.curId = 0
	}
}

func (c *Cache) Get(id byte) []byte {
	return c.container[id]
}

func (c *Cache) Empty() bool {
	return c.maxId < 0
}

func (c *Cache) Has(id byte) bool {
	return id <= byte(c.maxId)
}
