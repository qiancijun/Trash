package codec

import (
	"bufio"
	"encoding/gob"
	"io"
)


/**
* conn 是由构建函数传入，通常是通过 TCP 或者 Unix 建立 socket 时得到的链接实例
* dec 和 enc 对应 gob 的 Decoder 和 Encoder
* buf 是为了防止阻塞而创建的带缓冲的 Writer，一般这么做能提升性能。
*/
type GobCodec struct {
	conn io.ReadWriteCloser
	buf  *bufio.Writer
	dec  *gob.Decoder
	enc  *gob.Encoder
}

var _ Codec = (*GobCodec)(nil)
