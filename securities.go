package main

type SecurityType int64

const (
	Banknote   SecurityType = 1
	Bond                    = 2
	Stock                   = 3
	MutualFund              = 4
)

type Security struct {
	SecurityId int64
	Name       string
	// Number of decimal digits (to the right of the decimal point) this
	// security is precise to
	Precision int64
	Type      SecurityType
}
