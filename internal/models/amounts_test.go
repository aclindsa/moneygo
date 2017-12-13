package models_test

import (
	"github.com/aclindsa/moneygo/internal/models"
	"testing"
)

func expectedPrecision(t *testing.T, amount *models.Amount, precision uint64) {
	t.Helper()
	if amount.Precision() != precision {
		t.Errorf("Expected precision %d for %s, found %d", precision, amount.String(), amount.Precision())
	}
}

func TestAmountPrecision(t *testing.T) {
	var a models.Amount
	a.SetString("1.1928712")
	expectedPrecision(t, &a, 7)
	a.SetString("0")
	expectedPrecision(t, &a, 0)
	a.SetString("-0.7")
	expectedPrecision(t, &a, 1)
	a.SetString("-1.1837281037509137509173049173052130957210361309572047598275398265926351231426357130289523647634895285603247284245928712")
	expectedPrecision(t, &a, 118)
	a.SetInt64(1050)
	expectedPrecision(t, &a, 0)
}

func TestAmountRound(t *testing.T) {
	var a models.Amount
	tests := []struct {
		String   string
		RoundTo  uint64
		Expected string
	}{
		{"0", 5, "0"},
		{"929.92928", 2, "929.93"},
		{"-105.499999", 4, "-105.5"},
		{"0.5111111", 1, "0.5"},
		{"0.5111111", 0, "1"},
		{"9.876456", 3, "9.876"},
	}

	for _, test := range tests {
		a.SetString(test.String)
		a.Round(test.RoundTo)
		if a.String() != test.Expected {
			t.Errorf("Expected '%s' after Round(%d) to be %s intead of %s\n", test.String, test.RoundTo, test.Expected, a.String())
		}
	}
}

func TestAmountString(t *testing.T) {
	var a models.Amount
	for _, s := range []string{
		"1.1928712",
		"0",
		"-0.7",
		"-1.1837281037509137509173049173052130957210361309572047598275398265926351231426357130289523647634895285603247284245928712",
		"1050",
	} {
		a.SetString(s)
		if s != a.String() {
			t.Errorf("Expected '%s', found '%s'", s, a.String())
		}
	}

	a.SetString("+182.27")
	if "182.27" != a.String() {
		t.Errorf("Expected '182.27', found '%s'", a.String())
	}
	a.SetString("-0")
	if "0" != a.String() {
		t.Errorf("Expected '0', found '%s'", a.String())
	}
}

func TestWhole(t *testing.T) {
	var a models.Amount
	tests := []struct {
		String string
		Whole  int64
	}{
		{"0", 0},
		{"-0", 0},
		{"181.1293871230", 181},
		{"-0.1821", 0},
		{"99992737.9", 99992737},
		{"-7380.000009", -7380},
		{"4108740192740912741", 4108740192740912741},
	}

	for _, test := range tests {
		a.SetString(test.String)
		val, err := a.Whole()
		if err != nil {
			t.Errorf("Unexpected error: %s\n", err)
		} else if val != test.Whole {
			t.Errorf("Expected '%s'.Whole() to return %d intead of %d\n", test.String, test.Whole, val)
		}
	}

	a.SetString("81367662642302823790328492349823472634926342")
	_, err := a.Whole()
	if err == nil {
		t.Errorf("Expected error for overflowing int64")
	}
}

func TestFractional(t *testing.T) {
	var a models.Amount
	tests := []struct {
		String     string
		Precision  uint64
		Fractional int64
	}{
		{"0", 5, 0},
		{"181.1293871230", 9, 129387123},
		{"181.1293871230", 10, 1293871230},
		{"181.1293871230", 15, 129387123000000},
		{"1828.37", 7, 3700000},
		{"-0.748", 5, -74800},
		{"-9", 5, 0},
		{"-9.9", 1, -9},
	}

	for _, test := range tests {
		a.SetString(test.String)
		val, err := a.Fractional(test.Precision)
		if err != nil {
			t.Errorf("Unexpected error: %s\n", err)
		} else if val != test.Fractional {
			t.Errorf("Expected '%s'.Fractional(%d) to return %d intead of %d\n", test.String, test.Precision, test.Fractional, val)
		}
	}
}

func TestFromParts(t *testing.T) {
	var a models.Amount
	tests := []struct {
		Whole      int64
		Fractional int64
		Precision  uint64
		Result     string
	}{
		{839, 9080, 4, "839.908"},
		{-10, 0, 5, "-10"},
		{0, 900, 10, "0.00000009"},
		{9128713621, 87272727, 20, "9128713621.00000000000087272727"},
		{89, 1, 0, "90"}, // Not sure if this should really be supported, but it is
	}

	for _, test := range tests {
		a.FromParts(test.Whole, test.Fractional, test.Precision)
		if a.String() != test.Result {
			t.Errorf("Expected Amount.FromParts(%d, %d, %d) to return %s intead of %s\n", test.Whole, test.Fractional, test.Precision, test.Result, a.String())
		}
	}
}
