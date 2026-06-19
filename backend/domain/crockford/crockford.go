package crockford

import "encoding/base32"

const alphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

var encoding = base32.NewEncoding(alphabet).WithPadding(base32.NoPadding)

func Encode(b []byte) string {
	return encoding.EncodeToString(b)
}

func Decode(s string) ([]byte, error) {
	return encoding.DecodeString(s)
}
