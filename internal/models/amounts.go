package models

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strings"
)

type Amount struct {
	big.Rat
}

type PrecisionError struct {
	message string
}

func (p PrecisionError) Error() string {
	return p.message
}

// Whole returns the integral portion of the Amount
func (amount Amount) Whole() (int64, error) {
	var whole big.Int
	whole.Quo(amount.Num(), amount.Denom())
	if whole.IsInt64() {
		return whole.Int64(), nil
	}
	return 0, PrecisionError{"integral portion of Amount cannot be represented as an int64"}
}

// Fractional returns the fractional portion of the Amount, multiplied by
// 10^precision
func (amount Amount) Fractional(precision uint64) (int64, error) {
	if precision < amount.Precision() {
		return 0, PrecisionError{"Fractional portion of Amount cannot be represented with the given precision"}
	}

	// Reduce the fraction to its simplest form
	var r, gcd, d, n big.Int
	r.Rem(amount.Num(), amount.Denom())
	gcd.GCD(nil, nil, &r, amount.Denom())
	if gcd.Sign() != 0 {
		n.Quo(&r, &gcd)
		d.Quo(amount.Denom(), &gcd)
	} else {
		n.Set(&r)
		d.Set(amount.Denom())
	}

	// Figure out what we need to multiply the numerator by to get the
	// denominator to be 10^precision
	var prec, multiplier big.Int
	prec.SetUint64(precision)
	multiplier.SetInt64(10)
	multiplier.Exp(&multiplier, &prec, nil)
	multiplier.Quo(&multiplier, &d)

	n.Mul(&n, &multiplier)
	if n.IsInt64() {
		return n.Int64(), nil
	}
	return 0, fmt.Errorf("Fractional portion of Amount does not fit in int64 with given precision")
}

// FromParts re-assembles an Amount from the results from previous calls to
// Whole and Fractional
func (amount *Amount) FromParts(whole, fractional int64, precision uint64) {
	var fracnum, fracdenom, power big.Int
	fracnum.SetInt64(fractional)
	fracdenom.SetInt64(10)
	power.SetUint64(precision)
	fracdenom.Exp(&fracdenom, &power, nil)

	var fracrat big.Rat
	fracrat.SetFrac(&fracnum, &fracdenom)
	amount.Rat.SetInt64(whole)
	amount.Rat.Add(&amount.Rat, &fracrat)
}

// Round rounds the given Amount to the given precision
func (amount *Amount) Round(precision uint64) {
	// This probably isn't exactly the most efficient way to do this...
	amount.SetString(amount.FloatString(int(precision)))
}

func (amount Amount) String() string {
	return amount.FloatString(int(amount.Precision()))
}

func (amount *Amount) UnmarshalJSON(bytes []byte) error {
	var value string
	if err := json.Unmarshal(bytes, &value); err != nil {
		return err
	}
	value = strings.TrimSpace(value)

	if _, ok := amount.SetString(value); !ok {
		return fmt.Errorf("Failed to parse '%s' into Amount", value)
	}
	return nil
}

func (amount Amount) MarshalJSON() ([]byte, error) {
	return json.Marshal(amount.String())
}

// Precision returns the minimum positive integer p such that if you multiplied
// this Amount by 10^p, it would become an integer
func (amount Amount) Precision() uint64 {
	if amount.IsInt() || amount.Sign() == 0 {
		return 0
	}

	// Find d, the denominator of the reduced fractional portion of 'amount'
	var r, gcd, d big.Int
	r.Rem(amount.Num(), amount.Denom())
	gcd.GCD(nil, nil, &r, amount.Denom())
	if gcd.Sign() != 0 {
		d.Quo(amount.Denom(), &gcd)
	} else {
		d.Set(amount.Denom())
	}
	d.Abs(&d)

	var power, result big.Int
	one := big.NewInt(1)
	ten := big.NewInt(10)

	// Estimate an initial power
	if d.IsUint64() {
		power.SetInt64(int64(math.Log10(float64(d.Uint64()))))
	} else {

		// If the simplified denominator wasn't a uint64, its > 10^19
		power.SetInt64(19)
	}

	// If the initial estimate was too high, bring it down
	result.Exp(ten, &power, nil)
	for result.Cmp(&d) > 0 {
		power.Sub(&power, one)
		result.Exp(ten, &power, nil)
	}
	// If it was too low, bring it up
	for result.Cmp(&d) < 0 {
		power.Add(&power, one)
		result.Exp(ten, &power, nil)
	}

	if !power.IsUint64() {
		panic("Unable to represent Amount's precision as a uint64")
	}
	return power.Uint64()
}
