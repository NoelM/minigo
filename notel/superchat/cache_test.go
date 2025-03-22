package superchat

import "testing"

func TestFulfill(t *testing.T) {
	cache := NewCache()
	cache.Init()

	for i := 0; i < nbRows; i += 1 {
		cache.AppendTop(Message, i, 0)
	}

	firstRow := cache.FirstRow()
	if firstRow.msgId != nbRows-1 {
		t.Fatalf("first row contains msgId=%d [expected=%d]", firstRow.msgId, nbRows-1)
	}

	rowId, lastRow := cache.LastRow()
	if lastRow.msgId != 0 || rowId != nbRows-1 {
		t.Fatalf("last row contains msgId=%d [expected=%d] at rowId=%d [expected=%d]", lastRow.msgId, 0, rowId, nbRows-1)
	}
}

func TestEmpty(t *testing.T) {
	cache := NewCache()
	cache.Init()

	firstRow := cache.FirstRow()
	if firstRow.kind != Null || firstRow.msgId != 0 {
		t.Fatalf("first row contains kind=%d [expected=%d] msgId=%d [expected=%d]", firstRow.kind, Null, firstRow.msgId, 0)
	}

	rowId, lastRow := cache.LastRow()
	if lastRow.kind != Null || lastRow.msgId != 0 || rowId != 0 {
		t.Fatalf("last row contains kind=%d [expected=%d] msgId=%d [expected=%d] at rowId=%d [expected=%d]", firstRow.kind, Null, lastRow.msgId, 0, rowId, nbRows-1)
	}
}

func TestParialFill(t *testing.T) {
	cache := NewCache()
	cache.Init()

	for i := 0; i < msgStopRow; i += 1 {
		cache.AppendTop(Message, i, 0)
	}

	firstRow := cache.FirstRow()
	if firstRow.msgId != msgStopRow-1 {
		t.Fatalf("first row contains msgId=%d [expected=%d]", firstRow.msgId, msgStopRow-1)
	}

	rowId, lastRow := cache.LastRow()
	if lastRow.msgId != 0 || rowId != msgStopRow-1 {
		t.Fatalf("last row contains msgId=%d [expected=%d] at rowId=%d [expected=%d]", lastRow.msgId, 0, rowId, msgStopRow-1)
	}
}

func TestExtraFill(t *testing.T) {
	cache := NewCache()
	cache.Init()

	for i := 0; i < nbRows+1; i += 1 {
		cache.AppendTop(Message, i, 0)
	}

	firstRow := cache.FirstRow()
	if firstRow.msgId != nbRows {
		t.Fatalf("first row contains msgId=%d [expected=%d]", firstRow.msgId, nbRows)
	}

	rowId, lastRow := cache.LastRow()
	if lastRow.msgId != 1 || rowId != nbRows-1 {
		t.Fatalf("last row contains msgId=%d [expected=%d] at rowId=%d [expected=%d]", lastRow.msgId, 1, rowId, nbRows-1)
	}
}

func TestExtraFillWithAppendBottom(t *testing.T) {
	cache := NewCache()
	cache.Init()

	for i := 0; i < nbRows+1; i += 1 {
		cache.AppendTop(Message, i, 0)
	}

	lastMsgId := 256
	cache.AppendBottom(Message, lastMsgId, 0)

	firstRow := cache.FirstRow()
	if firstRow.msgId != nbRows-1 {
		t.Fatalf("first row contains msgId=%d [expected=%d]", firstRow.msgId, nbRows-1)
	}

	rowId, lastRow := cache.LastRow()
	if lastRow.msgId != lastMsgId || rowId != nbRows-1 {
		t.Fatalf("last row contains msgId=%d [expected=%d] at rowId=%d [expected=%d]", lastRow.msgId, lastMsgId, rowId, nbRows-1)
	}
}
