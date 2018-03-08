package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// NonceStr 生成一个长度为 2*n 的随机字符串
func NonceStr(n int) string {
	buf := make([]byte, n)
	_, err := rand.Read(buf)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(buf)
}
