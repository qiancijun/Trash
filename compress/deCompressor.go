package compress

import (
	"bufio"
	"os"
)

type DeCompressor struct {
	filename           string
	fre                []int
	tableLength        int
	totalCompressedBit int
	root               *Node
}

func NewDeCompressor(file string) *DeCompressor {
	return &DeCompressor{
		filename:           file,
		totalCompressedBit: 0,
	}
}

func (this *DeCompressor) readInt(in *bufio.Reader) int {
	ans, cnt, offset := 0, 0, 24
	for {
		b, err := in.ReadByte()
		if err != nil {
			break
		}
		cnt++
		val := int(b)
		if val < 0 {
			val += 256
		}
		ans = (ans | (val << offset))
		offset -= 8
		if cnt == 4 {
			break
		}
	}
	if cnt != 4 {
		panic("Read int error")
	}
	return ans
}

func (this *DeCompressor) deSerializeFreq(in *bufio.Reader) {
	tableSize := this.readInt(in)
	this.tableLength = tableSize
	fre := make([]int, tableSize)
	for i := 0; i < tableSize; i++ {
		fre[i] = this.readInt(in)
	}
	this.totalCompressedBit = this.readInt(in)
	Debug("Decompress table size: %d , totalByte: %d\n", this.tableLength, this.totalCompressedBit/8)
	this.root = NewHuffmanTree(fre)
	Debug("Generate Huffman Tree success\n")
}

func (this *DeCompressor) Decompress(destFileName string) {
	// 打开读写文件，带有缓冲区
	source, err := os.Open(this.filename)
	if err != nil {
		panic("Open file fail, " + err.Error())
	}
	dest, err := os.Create(destFileName)
	if err != nil {
		panic("Open file fail, " + err.Error())
	}

	br := bufio.NewReader(source)
	bw := bufio.NewWriter(dest)
	defer dest.Close()
	defer source.Close()
	// 解压之前的准备工作：
	// 1. 从文件中读取频率表，生成 Huffman 树
	this.deSerializeFreq(br)
	tmp := this.root
	processedBit := 0
	for {
		b, err := br.ReadByte()
		if err != nil {
			break
		}
		for i := 0; i < min(this.totalCompressedBit-processedBit, 8); i++ {
			if (b & (1 << (8 - i - 1))) == 0 {
				tmp = tmp.child[0]
			} else {
				tmp = tmp.child[1]
			}

			if tmp.isLeaf {
				bw.WriteByte(tmp.val)
				tmp = this.root
			}
		}
		processedBit += 8
	}
	bw.Flush()
}
