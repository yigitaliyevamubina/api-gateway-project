package etc

import (
	"crypto/rand"
	"io"
)

var (
	numbers = []byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}
)

func GenerateCode(max int, isDev ...bool) string {
	if len(isDev) == 1 {
		return "5555"
	}
	code := make([]byte, max)
	n, err := io.ReadAtLeast(rand.Reader, code, max)
	if n != max {
		panic(err)
	}
	for i := 0; i < len(code); i++ {
		code[i] = numbers[int(code[i])%len(numbers)]
	}

	return string(code)
}
