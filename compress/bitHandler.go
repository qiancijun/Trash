package compress

import (
	"bufio"
	"os"
)

const MOD int = 8

type BitHandler struct {
	fileName       string
	pos            int
	cur            byte
	totalBit       int
	totalBitOffset int
	bytesCount	   int
	dest           *os.File
	bw             *bufio.Writer
}

func NewBitHandler(fileName string) *BitHandler {
	dest, err := os.Create(fileName)
	if err != nil {
		panic("Create file fail, " + err.Error())
	}
	return &BitHandler{
		dest:     dest,
		fileName: fileName,
		cur:      0,
		pos:      0,
		totalBit: 0,
		bw:       bufio.NewWriter(dest),
	}
}

func (this *BitHandler) WriteInt(val int) {
	buf := []byte{
		byte((val & 0xFF000000) >> 24),
		byte((val & 0x00FF0000) >> 16),
		byte((val & 0x0000FF00) >> 8),
		byte(val & 0x000000FF),
	}
	this.bw.Write(buf)
}

func (this *BitHandler) SerializeFreq(fre []int) {
	this.WriteInt(len(fre))
	for _, i := range fre {
		this.WriteInt(i)
	}
	// 1 是因为预先写了一个频率表的长度
	// 频率表的长度在此处一定是256，但是为了可扩展性，写在这里
	this.totalBitOffset = (1 + len(fre)) * 4
	// 写这个 0 是为了记录 TotalBit 预留位置
	// 这就表明我们压缩的文件不得超过2^32 bit，也就是 4GB
	this.WriteInt(0)
	this.bw.Flush()
}

func (this *BitHandler) WriteBit(flag byte) {
	this.cur = ((this.cur << 1) | flag)
	this.pos++
	this.totalBit++

	// 凑够了一个字节，写入缓冲区
	// 缓冲区默认长度 4096
	if this.pos == MOD {
		this.pos = 0
		this.bw.WriteByte(this.cur)
		this.bytesCount++
		if this.bytesCount == 4096 {
			this.bw.Flush()
			this.bytesCount = 0
		}
		this.cur = 0
	}
}

func (this *BitHandler) WriteTotalBit() {
	f := this.dest
	f.Seek(int64(this.totalBitOffset), 0)
	buf := []byte{
		byte((this.totalBit & 0xFF000000) >> 24),
		byte((this.totalBit & 0x00FF0000) >> 16),
		byte((this.totalBit & 0x0000FF00) >> 8),
		byte(this.totalBit & 0x000000FF),
	}
	f.Write(buf)
}

func (this *BitHandler) Flush() {
	_ = this.bw.Flush()
}

func (this *BitHandler) Close() {
	if this.pos > 0 {
		for this.pos != MOD {
			this.cur <<= 1
			this.pos++
		}
		this.bw.WriteByte(this.cur)
		this.bw.Flush()
	}
}
