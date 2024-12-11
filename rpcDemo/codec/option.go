package codec

import "errors"

const MagicNumber = 0x3bef5c

/**
* 客户端与服务端的通信需要协商一些内容，比如 HTTP 分为 header 和 body 两部分
* 对于 RPC 协议来说，这部分需要自主设计，为了提升性能，一般在报文的最开始规划固定的字节
* 目前唯一需要协商的内容是消息的编解码方式
*
* 采用的编码方式为：
*   采用 JSON 编码 Option
*   后续的 header 和 body 的编码方式由 Option 中的 CodecType
*
* 报文的形式：
*   Option | Header1 | Body1 | Header2 | Body2 | ...
 */

type Option struct {
	MagicNumber int
	CodecType   Type
}

var DefaultOption = &Option{
	MagicNumber: MagicNumber,
	CodecType:   GobType,
}

func ParseOptions(opts ...*Option) (*Option, error) {
	if len(opts) == 0 || opts[0] == nil {
		return DefaultOption, nil
	}
	if len(opts) != 1 {
		return nil, errors.New("number of options is more than 1")
	}
	opt := opts[0]
	opt.MagicNumber = DefaultOption.MagicNumber
	if opt.CodecType == "" {
		opt.CodecType = DefaultOption.CodecType
	}
	return opt, nil
}
