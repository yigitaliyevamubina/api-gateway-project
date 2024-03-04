package etc

import "math/rand"

var (
	chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func RandomCode(max int) string {
	code := make([]byte, max)
	for i := 0; i < max; i++ {
		code[i] = chars[rand.Int()%len(chars)]
	}

	return string(code)
}
