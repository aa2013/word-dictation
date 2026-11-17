package util

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"regexp"
	"strings"
)

func RandomDigits(n int) string {
	out := make([]byte, n)
	for i := range out {
		v, _ := rand.Int(rand.Reader, big.NewInt(10))
		out[i] = byte('0' + v.Int64())
	}
	return string(out)
}

func RandomHex(n int) string {
	buf := make([]byte, n)
	_, _ = rand.Read(buf)
	return fmt.Sprintf("%x", buf)
}

var (
	emailPattern  = regexp.MustCompile(`^\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*$`)
	base64Pattern = regexp.MustCompile(`^(data:[^;]+;base64,)?([A-Za-z0-9+/]{4})*([A-Za-z0-9+/]{4}|[A-Za-z0-9+/]{3}=|[A-Za-z0-9+/]{2}==)$`)
)

func IsEmail(s string) bool { return s != "" && emailPattern.MatchString(s) }

func IsBase64(s string) bool { return s != "" && base64Pattern.MatchString(s) }

func StripBase64Prefix(s string) string {
	if strings.HasPrefix(s, "data:") {
		if i := strings.Index(s, ","); i != -1 {
			return s[i+1:]
		}
	}
	return s
}
