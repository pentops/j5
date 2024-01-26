package decimal_j5t

import (
	"github.com/shopspring/decimal"
)

func (d *Decimal) ToShop() (decimal.Decimal, error) {
	return decimal.NewFromString(d.Value)
}

func FromShop(d *decimal.Decimal) *Decimal {
	return &Decimal{
		Value: d.String(),
	}
}
