package compress

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

var DebugEnabled = true

func Debug(format string, a ...interface{}) (n int, err error) {
	if DebugEnabled {
		n, err = fmt.Printf(format, a...)
	}
	return
}

func FileMD5(filePath string) (string, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return "", err
    }
    hash := md5.New()
    _, _ = io.Copy(hash, file)
    return hex.EncodeToString(hash.Sum(nil)), nil
}

func max(a, b int) int {
	if a > b { return a }
	return b
}

func min(a, b int) int {
	if a > b { return b }
	return a
}