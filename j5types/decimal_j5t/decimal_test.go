package decimal_j5t

import (
	"testing"
)

func TestDecimal(t *testing.T) {
	df := FromFloat(1.33)
	di := FromInt(2)

	product, err := df.Mul(FromFloat(2.66))
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if product.ToString() != "3.5378" {
		t.Errorf("expected string to be 3.5378, got %v", product.Value)
	}

	product, err = di.Mul(FromFloat(1.33))
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if product.ToString() != "2.66" {
		t.Errorf("expected string to be 2.66, got %v", product.Value)
	}

	product, err = di.Mul(FromInt(0))
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if product.ToString() != "0" {
		t.Errorf("expected string to be 0, got %v", product.Value)
	}
}
