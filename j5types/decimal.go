package j5types

import "math/big"

func (d *Decimal) AsBig() *big.Rat {
	val, ok := new(big.Rat).SetString(d.String_)
	if !ok {
		panic("Invalid Decimal value")
	}
	return val
}
