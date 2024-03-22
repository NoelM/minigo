package minigo

import (
	"sync"
	"time"
)

// 1 byte, is 8 symbols (data) and 2 symbols (start and stop)
// 1 byte = 10 symbols
const ByteDurAt1200Bd = 8333 * time.Microsecond

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

	group *sync.WaitGroup

	Recv chan byte
	Send chan []byte
}

func NewNetwork(conn Connector, parity bool, group *sync.WaitGroup, source string) *Network {
	return &Network{
		conn:     conn,
		parity:   parity,
		source:   source,
		pceStack: make(chan []byte, 256),
		pceCache: NewCache(),
		group:    group,
		Recv:     make(chan byte, 1024),
		Send:     make(chan []byte, 256),
	}
}

func (n *Network) Serve() {
	n.group.Add(2)

	go n.recvLoop()
	go n.sendLoop()
}

func (n *Network) recvLoop() {
	pceAck := make([]byte, 0)

	for n.conn.Connected() {

		inBytes, readErr := n.conn.Read()
		if readErr != nil {
			warnLog.Printf("[%s] listen: stop loop: lost connection: %s\n", n.source, readErr.Error())

			// If the connection is lost, we send the stop signal to application:
			// Connexion/Fin
			n.Recv <- 0x13
			n.Recv <- 0x49

			break
		} else if len(inBytes) == 0 {
			// No data read, we continue
			continue
		}

		// Read all bytes received
		for _, b := range inBytes {

			// If parity enabled, check the bytes parity
			if n.parity {
				var parityErr error

				if b, parityErr = ValidAndRemoveParity(b); parityErr != nil {
					errorLog.Printf("[%s] listen: wrong parity ignored key=%x\n", n.source, b)
					continue
				}
			}

			if b == Sub {
				// SUB desginates parity error, we increment counter wetherer
				// the PCE starts or not
				warnLog.Printf("[%s] listen: minitel bad parity with SUB\n", n.source)
				n.incSub()

				// The byte is not sent to the minitel
				continue
			}

			if b == Nack {
				// The minitel requested a repetition, noted:
				// NACK, X, with X the block to be repeated
				if !n.pce {
					errorLog.Printf("[%s] listen: minitel request sync (NACK) but PCE is OFF\n", n.source)

					// The byte is not sent to the minitel loop
					continue
				}

				infoLog.Printf("[%s] listen: minitel request synchronization NACK frame\n", n.source)
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

					infoLog.Printf("[%s] listen: receieved NACK block=%x\n", n.source, b)
				} else {
					errorLog.Printf("[%s] listen: receieved invalid NACK block=%x\n", n.source, b)
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
					infoLog.Printf("[%s] listen: minitel acknowledged the PCE\n", n.source)

					n.pce = true
					n.pcePending = false
					pceAck = []byte{}

				} else if !next {
					// If next is at false, without an ack, this means that the suite
					// of bytes does not correspond to an ack
					warnLog.Printf("[%s] listen: unable to acknowledge PCE, invalid message\n", n.source)

					n.pcePending = false
					pceAck = []byte{}
				}

				// The protocol bytes are also sent to the minitel loop
			}

			// Push byte to the In chan, listened by the minitel loop
			n.Recv <- b
		}
	}

	n.group.Done()
}

func (n *Network) incSub() {
	if time.Since(n.subTime) < time.Minute {
		// If the counter has been reset less than a miniute ago, we increment
		n.subCnt += 1
		infoLog.Printf("[%s] inc-sub: increment SUB counter=%d since=%.0f sec\n", n.source, n.subCnt, time.Since(n.subTime).Seconds())

		if n.subCnt > MaxSubPerMinute && !n.pce && !n.pcePending {
			// If the limit is reached, and the PCE not active nor Pending (waiting for the ack)
			// the module asks for PCE ON
			warnLog.Printf("[%s] inc-sub: the activation of PCE has been transmitted\n", n.source)

			n.send(GetRequestPCE())
			n.pcePending = true
		}
	} else {
		// The last reset was more than a minute ago! So we reset the counters
		n.subTime = time.Now()
		n.subCnt = 1

		infoLog.Printf("[%s] inc-sub: reset SUB counter %d\n", n.source, n.subCnt)
	}
}

func (n *Network) sendLoop() {
	for n.conn.Connected() {
		if n.nackBlock || n.pcePending {
			// Some commands block all the Send data:
			// * NACK, waits for a blockId
			// * PCE ON, waits for the ack
			time.Sleep(100 * time.Millisecond)
			continue

		}

		if n.nackSynSend {
			// If the block id has been received, we try to repeat the last block

			if n.pceCache.Has(n.nackBlockId) {
				// The block is in cache, well done, we send it
				n.send(GetSynFrame(n.nackBlockId, n.pceCache.Get(n.nackBlockId)))
				infoLog.Printf("[%s] send: the SYN frame has been sent for block=%x\n", n.source, n.nackBlockId+0x40)

				// When NACK is sent, the minitel starts a counter of 1140ms (NackTimer),
				// we release the loop when the time is elapsed, in order to avoid
				// other messages sent while the minitel waits for the SYN frame
				time.Sleep(time.Since(n.nackTime) - NackTimer)
				n.nackSynSend = false

				// The Syn Block is sent, we go back to the begining
				continue

			}

			select {
			case blk := <-n.pceStack:
				// The block is not in cache, we get the fist block to send
				// It should only be the block 0x40
				n.pceCache.Add(blk)
				n.send(GetSynFrame(n.nackBlockId, blk))

				if n.nackBlockId == 0 {
					infoLog.Printf("[%s] send: the SYN frame has been sent with the first PCE\n", n.source)
				} else {
					warnLog.Printf("[%s] send: the SYN frame has been sent with the first PCE (block=%x)\n", n.source, n.nackBlockId+0x40)
				}

				// When NACK is sent, the minitel starts a counter of 1140ms (NackTimer),
				// we release the loop when the time is elapsed, in order to avoid
				// other messages sent while the minitel waits for the SYN frame
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
		case msg := <-n.Send:
			if n.pce {
				// when the PCE is active, we compute blocks
				// but we always send them later, to prevent any
				// NACK requested after a block
				for _, blk := range ApplyPCE(msg) {
					n.pceStack <- blk
				}

			} else {
				// otherwise we just send it!
				if n.parity {
					n.send(ApplyParity(msg))
				} else {
					n.send(msg)
				}
				continue
			}
		default:
			// no message? to avoid infinite loop, we'll wait a bit
			time.Sleep(100 * time.Millisecond)
		}

		// PCE send blocks from pceStack
		if n.pce && !n.nackSynSend {
			// The PCE is ON and no NACK/SYN from has been requested
			// We pop the first-inserted block, and send it if exists
			select {
			case blk := <-n.pceStack:
				n.send(blk)
				n.pceCache.Add(blk)
			default:
				time.Sleep(100 * time.Millisecond)
			}
		}
	}

	n.group.Done()
}

func (n *Network) send(data []byte) {
	n.conn.Write(data)
	time.Sleep(time.Duration(len(data)) * ByteDurAt1200Bd)
}

func (n *Network) Connected() bool {
	return n.conn.Connected()
}
