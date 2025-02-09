package superchat

const nbRows = 24

type RowKind int

const (
	Null RowKind = iota
	Date
	Message
)

type Row struct {
	kind   RowKind
	msgId  int
	lineId int
}

type Cache struct {
	rows []Row
}

func NewCache() *Cache {
	return &Cache{
		rows: make([]Row, nbRows),
	}
}

func (c *Cache) Init() {
	for i := 0; i < nbRows; i += 1 {
		c.rows[i] = Row{Null, 0, 0}
	}
}

func (c *Cache) AppendBottom(kind RowKind, msgId, lineId int) {
	// We append a line at the screen's bottom
	// The content of row number `nbRows-1` will be modified
	//
	// LINE ID
	// 0         1
	// 1         2
	// 2         3
	// ...       ...
	// nbRows-1  [NEW LINE]
	//
	copy(c.rows, c.rows[1:])
	c.rows[nbRows-1] = Row{kind, msgId, lineId}
}

func (c *Cache) AppendTop(kind RowKind, msgId, lineId int) {
	// We append a line at the screen's top
	// The content of row number `0` will be modified
	//
	// LINE ID
	// 0         [NEW LINE]
	// 1         0
	// 2         1
	// ...       ...
	// nbRows-1  nbRows-2
	//
	copy(c.rows[1:], c.rows)
	c.rows[0] = Row{kind, msgId, lineId}
}

func (c *Cache) FirstRow() Row {
	return c.rows[0]
}

func (c *Cache) LastRow() (int, Row) {
	for i := nbRows - 1; i >= 0; i -= 1 {
		if c.rows[i].kind != Null {
			return i, c.rows[i]
		}
	}

	return 0, Row{Null, 0, 0}
}
