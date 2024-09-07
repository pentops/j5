package id62

import (
	"encoding/base64"
	"fmt"
	"math/big"
	"regexp"

	"github.com/google/uuid"
)

var PatternString = `^[0-9A-Za-z]{22}$`
var Pattern = regexp.MustCompile(PatternString)

type UUID [16]byte

func New() UUID {
	id := uuid.Must(uuid.NewV7())
	return UUID(id[:])
}

func NewString() string {
	return New().String()
}

func (id UUID) String() string {
	return base62String(id[:])
}

func (id UUID) UUIDString() string {
	return uuid.UUID(id).String()
}

func (id UUID) Base64String() string {
	return base64.StdEncoding.EncodeToString(id[:])
}

func Parse(s string) (UUID, error) {
	var uuid UUID
	if err := parseBase62(s, uuid[:]); err != nil {
		return UUID{}, err
	}
	return uuid, nil
}

func base62String(id []byte) string {
	var i big.Int
	i.SetBytes(id)
	str := i.Text(62)
	if len(str) < 22 {
		str = fmt.Sprintf("%022s", str)
	} else if len(str) > 22 {
		panic("base62 value is too large")
	}
	return str
}

func parseBase62(s string, into []byte) error {
	var i big.Int
	_, ok := i.SetString(s, 62)
	if !ok {
		return fmt.Errorf("cannot parse base62: %q", s)
	}
	valBytes := i.Bytes()
	if len(valBytes) > len(into) {
		return fmt.Errorf("base62 value is too large: %d > %d", len(valBytes), len(into))

	} else if len(valBytes) < len(into) { // left pad with zeros
		copy(into[len(into)-len(valBytes):], valBytes)
	} else {
		copy(into, valBytes)
	}
	return nil
}
