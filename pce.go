package minigo

import (
	"sync"
	"time"
)

const MaxSubPerMinute = 1

// Imported from 'linux/lib/crc7.c'
//
// Table for CRC-7 (polynomial x^7 + x^3 + 1).
// This is a big-endian CRC (msbit is highest power of x),
// aligned so the msbit of the byte is the x^6 coefficient
// and the lsbit is not used.
var crc7SyndromeTable = []byte{
	0x00, 0x12, 0x24, 0x36, 0x48, 0x5a, 0x6c, 0x7e,
	0x90, 0x82, 0xb4, 0xa6, 0xd8, 0xca, 0xfc, 0xee,
	0x32, 0x20, 0x16, 0x04, 0x7a, 0x68, 0x5e, 0x4c,
	0xa2, 0xb0, 0x86, 0x94, 0xea, 0xf8, 0xce, 0xdc,
	0x64, 0x76, 0x40, 0x52, 0x2c, 0x3e, 0x08, 0x1a,
	0xf4, 0xe6, 0xd0, 0xc2, 0xbc, 0xae, 0x98, 0x8a,
	0x56, 0x44, 0x72, 0x60, 0x1e, 0x0c, 0x3a, 0x28,
	0xc6, 0xd4, 0xe2, 0xf0, 0x8e, 0x9c, 0xaa, 0xb8,
	0xc8, 0xda, 0xec, 0xfe, 0x80, 0x92, 0xa4, 0xb6,
	0x58, 0x4a, 0x7c, 0x6e, 0x10, 0x02, 0x34, 0x26,
	0xfa, 0xe8, 0xde, 0xcc, 0xb2, 0xa0, 0x96, 0x84,
	0x6a, 0x78, 0x4e, 0x5c, 0x22, 0x30, 0x06, 0x14,
	0xac, 0xbe, 0x88, 0x9a, 0xe4, 0xf6, 0xc0, 0xd2,
	0x3c, 0x2e, 0x18, 0x0a, 0x74, 0x66, 0x50, 0x42,
	0x9e, 0x8c, 0xba, 0xa8, 0xd6, 0xc4, 0xf2, 0xe0,
	0x0e, 0x1c, 0x2a, 0x38, 0x46, 0x54, 0x62, 0x70,
	0x82, 0x90, 0xa6, 0xb4, 0xca, 0xd8, 0xee, 0xfc,
	0x12, 0x00, 0x36, 0x24, 0x5a, 0x48, 0x7e, 0x6c,
	0xb0, 0xa2, 0x94, 0x86, 0xf8, 0xea, 0xdc, 0xce,
	0x20, 0x32, 0x04, 0x16, 0x68, 0x7a, 0x4c, 0x5e,
	0xe6, 0xf4, 0xc2, 0xd0, 0xae, 0xbc, 0x8a, 0x98,
	0x76, 0x64, 0x52, 0x40, 0x3e, 0x2c, 0x1a, 0x08,
	0xd4, 0xc6, 0xf0, 0xe2, 0x9c, 0x8e, 0xb8, 0xaa,
	0x44, 0x56, 0x60, 0x72, 0x0c, 0x1e, 0x28, 0x3a,
	0x4a, 0x58, 0x6e, 0x7c, 0x02, 0x10, 0x26, 0x34,
	0xda, 0xc8, 0xfe, 0xec, 0x92, 0x80, 0xb6, 0xa4,
	0x78, 0x6a, 0x5c, 0x4e, 0x30, 0x22, 0x14, 0x06,
	0xe8, 0xfa, 0xcc, 0xde, 0xa0, 0xb2, 0x84, 0x96,
	0x2e, 0x3c, 0x0a, 0x18, 0x66, 0x74, 0x42, 0x50,
	0xbe, 0xac, 0x9a, 0x88, 0xf6, 0xe4, 0xd2, 0xc0,
	0x1c, 0x0e, 0x38, 0x2a, 0x54, 0x46, 0x70, 0x62,
	0x8c, 0x9e, 0xa8, 0xba, 0xc4, 0xd6, 0xe0, 0xf2,
}

func crc7BeByte(crc, data byte) byte {
	return crc7SyndromeTable[crc^data]
}

func getCRC7(data []byte) byte {
	var crc byte
	for _, b := range data {
		crc = crc7BeByte(crc, b)
	}

	return crc >> 1
}

func ComputePCEBlock(buf []byte) []byte {
	// make sure we a have the good length
	inner := make([]byte, 15)
	copy(inner, buf)

	crc := getCRC7(inner)

	inner = append(inner, GetByteWithParity(crc), 0)
	return inner
}

type PCEManager struct {
	status  bool
	blocks  *Stack
	remains [][]byte

	subTs  time.Time
	subCnt int

	nackRcv bool
	blockId int
	pendSyn bool

	conn     Connector
	writeMtx *sync.Mutex

	source string
}

func NewPCEManager(conn Connector, writeMtx *sync.Mutex, source string) *PCEManager {
	return &PCEManager{
		conn:     conn,
		writeMtx: writeMtx,
		blockId:  -1,
		blocks:   NewStack(16),
		source:   source,
	}
}

func (p *PCEManager) On() {
	p.status = true
	p.blocks.Reset()

	p.writeMtx.Unlock()
}

func (p *PCEManager) Off() {
	p.status = false
}

func (p *PCEManager) Status() bool {
	return p.status
}

func (p *PCEManager) IncSub() bool {
	// The enablement of PCE is restricted to a rate of 10 SUB per minutes
	if time.Since(p.subTs) < time.Minute {
		p.subCnt += 1
		infoLog.Printf("[%s] listen: recv SUB, first=%.0fs cnt=%d pce=%t\n", p.source, time.Since(p.subTs).Seconds(), p.subCnt, p.status)

		if p.subCnt > MaxSubPerMinute && !p.status {
			infoLog.Printf("[%s] listen: too many SUB cnt=%d pce=%t: activate PCE\n", p.source, p.subCnt, p.status)

			p.startPCE()
			return true
		}

	} else {
		p.subCnt = 1
		infoLog.Printf("[%s] listen: recv SUB, first=%.0fs cnt=%d pce=%t\n", p.source, time.Since(p.subTs).Seconds(), p.subCnt, p.status)

		p.subTs = time.Now()
	}

	return false
}

func (p *PCEManager) resetNack() {
	p.nackRcv = false
	p.blockId = -1
	p.pendSyn = false
}

func (p *PCEManager) GotNack() {
	p.writeMtx.Lock()
	p.nackRcv = true
}

func (p *PCEManager) WaitForBlockId() bool {
	return p.nackRcv && p.blockId < 0
}

func (p *PCEManager) GotBlockId(blockId int32) {
	defer p.writeMtx.Unlock()

	p.blockId = p.blockId - 0x40
	if p.blockId < 0 || p.blockId > 15 {
		p.resetNack()
	}

	if p.blocks.Empty() {
		p.pendSyn = true
		return
	}

	p.synSend()
}

func (p *PCEManager) Send(msg []byte) {
	p.remains = append(p.remains, ApplyPCE(msg, true)...)
}

func (p *PCEManager) SendNext() {
	var buf []byte

	if p.pendSyn {
		buf = ApplyParity([]byte{Syn, Syn, 0x40 + byte(p.blockId)})
	}

	if len(p.remains) > 0 {
		blk := p.popRemains()

		if p.pendSyn {
			buf = append(buf, p.popRemains()...)
			p.resetNack()
		} else {
			buf = blk
			p.blocks.Add(blk)
		}

		p.writeAndLock(buf)
	}
}

func (p *PCEManager) writeAndLock(b []byte) {
	p.writeMtx.Lock()
	defer p.writeMtx.Unlock()
	p.conn.Write(b)
}

func (p *PCEManager) popRemains() []byte {
	b := make([]byte, len(p.remains[0]))
	copy(b, p.remains[0])

	if len(p.remains) > 1 {
		p.remains = p.remains[1:]
	} else {
		p.remains = make([][]byte, 0)
	}

	return b
}

func (p *PCEManager) synSend() {
	synMsg := ApplyParity([]byte{Syn, Syn, 0x40 + byte(p.blockId)})
	synMsg = append(synMsg, p.blocks.Get(p.blockId)...)

	p.conn.Write(synMsg)
}

func (p *PCEManager) startPCE() error {
	p.writeMtx.Lock()

	buf, _ := GetProCode(Pro2)
	buf = append(buf, Start, PCE)

	return p.conn.Write(ApplyParity(buf))
}
