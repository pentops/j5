package decimal_j5t

import (
	"fmt"

	"github.com/shopspring/decimal"
)

func (d *Decimal) ToShop() (decimal.Decimal, error) {
	return decimal.NewFromString(d.Value)
}

func (d *Decimal) Decimal() decimal.Decimal {
	sd, err := d.ToShop()
	if err != nil {
		panic(err)
	}
	return sd
}

func (d *Decimal) Neg() *Decimal {
	sd := d.Decimal().Neg()
	return FromShop(sd)
}

func (d *Decimal) Add(d2 *Decimal) *Decimal {
	sd := d.Decimal().Add(d2.Decimal())
	return FromShop(sd)
}

func (d *Decimal) Sub(d2 *Decimal) *Decimal {
	sd := d.Decimal().Sub(d2.Decimal())
	return FromShop(sd)
}

func (d *Decimal) Mul(d2 *Decimal) (*Decimal, error) {
	s1, err := d.ToShop()
	if err != nil {
		return nil, fmt.Errorf("error converting %v to decimal", s1.String())
	}

	s2, err := d2.ToShop()
	if err != nil {
		return nil, fmt.Errorf("error converting %v to decimal", s2.String())
	}

	return FromShop(s1.Mul(s2)), nil
}

func (d *Decimal) Div(d2 *Decimal) (*Decimal, error) {
	s1, err := d.ToShop()
	if err != nil {
		return nil, fmt.Errorf("error converting %v to decimal", s1.String())
	}

	s2, err := d2.ToShop()
	if err != nil {
		return nil, fmt.Errorf("error converting %v to decimal", s2.String())
	}

	if s2.IsZero() {
		return nil, fmt.Errorf("error divide by zero")
	}

	return FromShop(s1.Div(s2)), nil
}

func FromShop(d decimal.Decimal) *Decimal {
	return &Decimal{
		Value: d.String(),
	}
}

func (d *Decimal) ToString() string {
	if d == nil {
		return "null"
	}
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

type Testing interface {
	Errorf(format string, args ...any)
}

func AssertEqual(t Testing, want string, d1 *Decimal, name string) {
	dWant, err := decimal.NewFromString(want)
	if err != nil {
		t.Errorf("%s: error converting want %s to decimal: %v", name, want, err)
		return
	}

	if !d1.Decimal().Equal(dWant) {
		t.Errorf("%s: expected %s, got %s", name, want, d1.ToString())
	}
}
