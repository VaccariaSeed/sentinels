package command

import (
	"crypto/rand"
	"fmt"
	"io"
)

// 创建一个uuid
func generateUUID() string {
	uuid := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, uuid)
	if err != nil {
		panic(err)
	}

	// 设置版本号 (4)
	uuid[6] = (uuid[6] & 0x0f) | 0x40
	// 设置变体 (10)
	uuid[8] = (uuid[8] & 0x3f) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x",
		uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}
