package compress

import "testing"

func TestCalcFrequence(t *testing.T) {
	comprossor := NewCompressor("./test.docx")
	comprossor.calcFrequence()

}

func TestHuffmanTree(t *testing.T) {
	compressor := NewCompressor("./compressor.go")
	compressor.calcFrequence()
	compressor.generateTable()
	for k, v := range compressor.encodeTable {
		Debug("%v:%d -> %s\n", k, compressor.fre[k], v)
	}
}

func TestCompress(t *testing.T) {
	compressor := NewCompressor("./bible.txt")
	compressor.Compress("./test.dat")
}

func TestDecompress(t *testing.T) {
	dec := NewDeCompressor("./test.dat")
	dec.Decompress("test2.txt")
}

func TestCompressAndDecompressor(t *testing.T) {
	compressor := NewCompressor("./bible.txt")
	compressor.Compress("./test.dat")
	dec := NewDeCompressor("./test.dat")
	dec.Decompress("test2.txt")

	m1, _ := FileMD5("./bible.txt")
	m2, _ := FileMD5("./test2.txt")
	Debug("%s\n%s\n%v", m1, m2, m1 == m2)
}