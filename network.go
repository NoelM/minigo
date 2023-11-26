package minigo

import (
	"time"
)

type Network struct {
	conn   Connector
	parity bool
	source string

	subTime time.Time
	subCnt  int

	nackTime    time.Time
	nackBlock   bool
	nackBlockId byte
	nackSynSend bool

	pce        bool
	pcePending bool
	pceStack   chan []byte
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
		pceStack: make(chan []byte, 256),
		pceCache: NewCache(),
		msgStack: NewStack(),
		In:       make(chan byte, 1024),
		Out:      make(chan []byte, 256),
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

			if b == Sub {
				// SUB desginates parity error, we increment counter wetherer
				// the PCE starts or not
				n.IncSub()

				// The byte is not sent to the minitel
				continue
			}

			if b == Nack {
				// The minitel requested a repetition, noted:
				// NACK, X, with X the block to be repeated
				n.nackTime = time.Now()
				n.nackBlock = true

				// The byte is not sent to the minitel loop
				continue
			}

			if n.nackBlock {
				// If previously on receieved a NACK, now one waits for a block
				if b >= 0x40 && b <= 0x4F {
					// Irrelevant block ID, bye, bye
					n.nackBlockId = b - byte(0x40)
					n.nackSynSend = true
				}
				n.nackBlock = false

				// The byte is not sent to the minitel loop
				continue
			}

			// Check if the PCE has been requested earlier, by IncSub
			if n.pcePending {
				// The message of ack is 4-byte long, so one appends bytes
				pceAck = append(pceAck, b)

				// The function AckPCE, returns:
				// * ack if the nack has been validated
				// * next if it needs another byte, set to false if the
				//   suite of bytes are irrelevant
				if ack, next := AckPCE(pceAck); ack {
					n.pce = true
					n.pcePending = false
					pceAck = []byte{}

				} else if !next {
					// If next is at false, without an ack, this means that the suite
					// of bytes does not correspond to an ack
					n.pcePending = false
					pceAck = []byte{}
				}

				// The protocol bytes are also sent to the minitel loop
			}

			// Push byte to the In chan, listened by the minitel loop
			n.In <- b
		}
	}
}

func (n *Network) IncSub() {
	if time.Since(n.subTime) < time.Minute {
		// If the counter has been reset less than a miniute ago, we increment
		n.subCnt += 1

		if n.subCnt > MaxSubPerMinute && !n.pce && !n.pcePending {
			// If the limit is reached, and the PCE not active nor Pending (waiting for the ack)
			// the module asks for PCE ON
			n.send(GetRequestPCE())
			n.pcePending = true
		}
	} else {
		// The last reset was more than a minute ago! So we reset the counters
		n.subTime = time.Now()
		n.subCnt = 1
	}
}

func (n *Network) SendLoop() {
	for n.conn.Connected() {
		if n.nackBlock || n.pcePending {
			// Some commands block all the Send commands:
			// * NACK, waits for a blockId
			// * PCE ON, waits for the ack
			time.Sleep(100 * time.Millisecond)
			continue

		}

		if n.nackSynSend {
			// If the block id has been received, we try to repeat the last block

			if n.pceCache.Has(n.nackBlockId) {
				// If the block is in cache, well done, we send it
				n.send(GetSynFrame(n.nackBlockId, n.pceCache.Get(n.nackBlockId)))

				time.Sleep(time.Since(n.nackTime) - NackTimer)
				n.nackSynSend = false

				// The Syn Block is sent, we go back to the begining
				continue

			}

			select {
			case blk := <-n.pceStack:
				n.pceCache.Add(blk)
				n.send(GetSynFrame(n.nackBlockId, blk))

				time.Sleep(time.Since(n.nackTime) - NackTimer)
				n.nackSynSend = false

				// The Syn Block is sent, we go back to the begining
				continue

			default:
				// Nothing to send, let's check the Out chan
			}
		}

		// The Out chan must not be blocking, so we handle it on a 'select'
		select {
		case msg := <-n.Out:
			if n.pce {
				// when the PCE is active, we compute blocks
				for _, blk := range ApplyPCE(msg) {
					n.pceStack <- blk
				}

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
