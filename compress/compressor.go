package compress

import (
	"bufio"
	"io"
	"os"
)

// 根究字节来压缩，所以一个频率表的长度最大是256
const TABLE_LENGTH = 256

type Compressor struct {
	fileName          string
	encodeTable       map[byte]string
	decodeTable       map[string]byte
	// 频率计数
	fre               []int
	// 原始文件的总字节数
	totalOriginByte   int
	// 压缩后文件的总字节数
	totalCompressByte int
	// 已经读取的文件总字节数
	alreadyReadByte   int
	// 哈夫曼树根节点
	root              *Node
	bitWriter 		  *BitHandler
}

func NewCompressor(file string) *Compressor {
	
	compressor := &Compressor{
		fileName:          file,
		encodeTable:       make(map[byte]string),
		decodeTable:       make(map[string]byte),
		fre:               make([]int, 256),
		totalOriginByte:   0,
		totalCompressByte: 0,
		root:              nil,
	}
	compressor.generateTable()
	return compressor
}

// 计算文件中每个字节出现的频率
func (compressor *Compressor) calcFrequence() {
	file, err := os.Open(compressor.fileName)
	if err != nil {
		panic("Open file fail")
	}
	reader := bufio.NewReader(file)
	buf := make([]byte, 1024)
	total := 0
	for {
		n, err := reader.Read(buf);
		if err == io.EOF || n == 0 {
			break
		}
		if err != nil {
			panic("Read file fail, " + err.Error())
		}
		for i := 0; i < n; i++ {
			compressor.fre[buf[i]]++
		}
		total += n
	}
	compressor.totalOriginByte = total
	Debug("Origin file total byte: %d bytes\n", compressor.totalOriginByte)
}



// 将 Huffman 树生成的结构添加到映射表中
func (compressor *Compressor) generateTable() {
	compressor.calcFrequence()
	root := NewHuffmanTree(compressor.fre)
	compressor.root = root
	Debug("huffmanTree generate success\n")
	var dfs func(*Node, string)
	dfs = func(node *Node, tmp string) {
		if node == nil {
			return
		}
		if node.isLeaf {
			compressor.encodeTable[node.val] = tmp
			compressor.decodeTable[tmp] = byte(node.val)
		}
		dfs(node.child[0], tmp + "0")
		dfs(node.child[1], tmp + "1")
	}
	dfs(root, "")
}

func (compressor *Compressor) Compress(dest string) {
	compressor.compressPreHandle(dest)
	// 每次读取 4KB 的数据
	// 为什么选取 4KB，和 Golang 底层的 bufio 有关
	// 默认创建的 bufio.Writer 的缓冲区大小是 4096
	// buf := make([]byte, 4096) 
	source, err := os.Open(compressor.fileName)
	if err != nil {
		panic("Open file fail, " + err.Error())
	}
	defer source.Close()
	br := bufio.NewReader(source)
	bitW := compressor.bitWriter
	for {
		b, err := br.ReadByte()
		if err != nil {
			break
		}
		compressor.alreadyReadByte++
		code := compressor.encodeTable[b]
		for _, c := range code {
			if c == '1' {
				bitW.WriteBit(1)
			} else {
				bitW.WriteBit(0)
			}
		}
	}
	// 补充最后的字节
	bitW.Flush()
	bitW.Close()
	bitW.WriteTotalBit()
	Debug("Compressed byte: %d, Compressed rate: %.6f\n", bitW.totalBit / 8, float64((1.0 * bitW.totalBit / 8)) / float64((1.0 * compressor.totalOriginByte)))
	Debug("Read Byte: %d\n", compressor.alreadyReadByte)
}

func (compressor *Compressor) compressPreHandle(dest string) {
	bitWriter := NewBitHandler(dest)
	compressor.bitWriter = bitWriter
	// 将 Huffman 树序列化到文件中，只需要记录频率表，Huffman 树可以根据频率表重新生成，减少空间
	fre := compressor.fre
	// 先写入频率表的长度，其实这里一直都是 256，为了可扩展性
	bitWriter.SerializeFreq(fre)
}