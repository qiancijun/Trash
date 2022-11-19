package jwtdemo

import (
	"crypto/md5"
	"encoding/hex"
)

func MD5Salt(data, salt string, iteration int) string {
	b, s := []byte(data), []byte(salt)
	h := md5.New()
	h.Write(s)
	h.Write(b)
	var ans []byte
	ans = h.Sum(nil)
	for i := 0; i < iteration - 1; i++ {
		h.Reset()
		h.Write(ans)
		ans = h.Sum(nil)
	}
	return hex.EncodeToString(ans)
}