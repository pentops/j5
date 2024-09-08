package decimal_j5t

import (
	"fmt"

	"github.com/shopspring/decimal"
)

func (d *Decimal) ToShop() (decimal.Decimal, error) {
	return decimal.NewFromString(d.Value)
}

func FromShop(d decimal.Decimal) *Decimal {
	return &Decimal{
		Value: d.String(),
	}
}

func (d *Decimal) ToString() string {
	return d.Value
}

func FromString(s string) *Decimal {
	return &Decimal{
		Value: s,
	}
}

func FromFloat(f float64) *Decimal {
	d := decimal.NewFromFloat(f)
	return &Decimal{
		Value: d.String(),
	}
}

func FromInt(i int64) *Decimal {
	d := decimal.NewFromInt(i)
	return &Decimal{
		Value: d.String(),
	}
}

func Zero() *Decimal {
	return &Decimal{
		Value: "0",
	}
}

// Scan implements sql.Scanner
func (d *Decimal) Scan(src any) error {
	switch src := src.(type) {
	case string:
		d.Value = src
		return nil
	default:
		return fmt.Errorf("unsupported type %T", src)
	}
}
