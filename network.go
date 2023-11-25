package minigo

import "time"

type Network struct {
	conn   Connector
	parity bool
	source string

	subTs  time.Time
	subCnt int

	nackTs      time.Time
	nackBlock   bool
	nackBlockId byte
	nackSynSend bool

	pce        bool
	pcePending bool
	pceStack   *Stack
	pceCache   *Cache

	msgStack *Stack

	In  chan byte
	Out chan []byte
}

func NewNetwork(conn Connector, parity bool, source string) *Network {
	return &Network{
		conn:     conn,
		parity:   parity,
		source:   source,
		pceStack: NewStack(),
		pceCache: NewCache(),
		msgStack: NewStack(),
		In:       make(chan byte),
		Out:      make(chan []byte),
	}
}

func (n *Network) ListenLoop() {
	pceAck := make([]byte, 0)

	for n.conn.Connected() {

		inBytes, readErr := n.conn.Read()
		if readErr != nil {
			warnLog.Printf("[%s] listen: stop loop: lost connection: %s\n", n.source, readErr.Error())
			break
		} else if len(inBytes) == 0 {
			continue
		}

		// Read all bytes received
		for _, b := range inBytes {
			if n.parity {
				if b, parityErr := CheckByteParity(b); parityErr != nil {
					errorLog.Printf("[%s] listen: wrong parity ignored key=%x\n", n.source, b)
					continue
				}
			}

			if n.pcePending {
				pceAck = append(pceAck, b)
				if ack, next := AckPCE(pceAck); ack {
					n.pce = true
					n.pcePending = false
					pceAck = []byte{}

				} else if !next {
					n.pcePending = false
					pceAck = []byte{}
				}
			}

			if n.nackBlock {
				if b >= 0x40 && b <= 0x4F {
					n.nackBlockId = b - byte(0x40)
					n.nackSynSend = true
				}
				n.nackBlock = false

			} else if b == Sub {
				n.IncSub()

			} else if b == Nack {
				n.nackTs = time.Now()
				n.nackBlock = true

			} else {
				n.In <- b

			}
		}
	}
}

func (n *Network) IncSub() {
	if time.Since(n.subTs) < time.Minute {
		n.subCnt += 1
		if n.subCnt > MaxSubPerMinute && !n.pce && !n.pcePending {
			n.send(GetRequestPCE())
			n.pcePending = true
		}
	} else {
		n.subCnt = 1
	}
}

func (n *Network) SendLoop() {
	for n.conn.Connected() {
		if n.nackBlock || n.pcePending {
			// If NACK has been received, we pause of a while
			time.Sleep(100 * time.Millisecond)
			continue

		} else if n.nackSynSend {
			// If the block id has been received, we try to repeat the last block

			if !n.pceCache.Empty() {
				// If the block is in cache, well done, we send it
				n.send(GetSynFrame(n.nackBlockId, n.pceCache.Get(n.nackBlockId)))

				time.Sleep(time.Since(n.nackTs) - NackTimer)
				n.nackSynSend = false

			} else if !n.pceStack.Empty() {
				// If the block is not in cache we try the next available in stack
				blk := n.pceStack.Pop()
				n.pceCache.Add(blk)

				n.send(GetSynFrame(n.nackBlockId, blk))

				time.Sleep(time.Since(n.nackTs) - NackTimer)
				n.nackSynSend = false
			}
		}

		// The Out chan must not be blocking, so we handle it on a 'select'
		select {
		case msg := <-n.Out:
			if n.pce {
				// when the PCE is active, we compute blocks
				n.pceStack.Add(ApplyPCE(msg)...)
			} else {
				// otherwise we just send it!
				n.send(msg)
				continue
			}
		default:
			// no message? to avoid infinite loop, we'll wait a bit
			time.Sleep(100 * time.Millisecond)
		}

		// PCE send blocks from pceStack
		if n.pce && !n.nackSynSend {
			// We pop the first-inserted block, and send it if exists
			if blk := n.pceStack.Pop(); blk != nil {
				n.send(blk)
				n.pceCache.Add(blk)
			}
		}
	}
}

func (n *Network) send(data []byte) {
	n.conn.Write(data)
}

func (n *Network) Connected() bool {
	return n.conn.Connected()
}
