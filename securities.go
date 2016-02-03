package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

const (
	Currency int64 = 1
	Stock          = 2
)

type Security struct {
	SecurityId  int64
	Name        string
	Description string
	Symbol      string
	// Number of decimal digits (to the right of the decimal point) this
	// security is precise to
	Precision int
	Type      int64
	// AlternateId is CUSIP for Type=Stock
	AlternateId string
}

type SecurityList struct {
	Securities *[]*Security `json:"securities"`
}

var security_map = map[int64]*Security{
	// ISO 4217, from http://www.currency-iso.org/en/home/tables/table-a1.html
	8: &Security{
		SecurityId:  8,
		Name:        "ALL",
		Description: "Lek",
		Symbol:      "ALL",
		Precision:   2,
		Type:        Banknote},
	12: &Security{
		SecurityId:  12,
		Name:        "DZD",
		Description: "Algerian Dinar",
		Symbol:      "DZD",
		Precision:   2,
		Type:        Banknote},
	32: &Security{
		SecurityId:  32,
		Name:        "ARS",
		Description: "Argentine Peso",
		Symbol:      "ARS",
		Precision:   2,
		Type:        Banknote},
	36: &Security{
		SecurityId:  36,
		Name:        "AUD",
		Description: "Australian Dollar",
		Symbol:      "AUD",
		Precision:   2,
		Type:        Banknote},
	44: &Security{
		SecurityId:  44,
		Name:        "BSD",
		Description: "Bahamian Dollar",
		Symbol:      "BSD",
		Precision:   2,
		Type:        Banknote},
	48: &Security{
		SecurityId:  48,
		Name:        "BHD",
		Description: "Bahraini Dinar",
		Symbol:      "BHD",
		Precision:   3,
		Type:        Banknote},
	50: &Security{
		SecurityId:  50,
		Name:        "BDT",
		Description: "Taka",
		Symbol:      "BDT",
		Precision:   2,
		Type:        Banknote},
	51: &Security{
		SecurityId:  51,
		Name:        "AMD",
		Description: "Armenian Dram",
		Symbol:      "AMD",
		Precision:   2,
		Type:        Banknote},
	52: &Security{
		SecurityId:  52,
		Name:        "BBD",
		Description: "Barbados Dollar",
		Symbol:      "BBD",
		Precision:   2,
		Type:        Banknote},
	60: &Security{
		SecurityId:  60,
		Name:        "BMD",
		Description: "Bermudian Dollar",
		Symbol:      "BMD",
		Precision:   2,
		Type:        Banknote},
	64: &Security{
		SecurityId:  64,
		Name:        "BTN",
		Description: "Ngultrum",
		Symbol:      "BTN",
		Precision:   2,
		Type:        Banknote},
	68: &Security{
		SecurityId:  68,
		Name:        "BOB",
		Description: "Boliviano",
		Symbol:      "BOB",
		Precision:   2,
		Type:        Banknote},
	72: &Security{
		SecurityId:  72,
		Name:        "BWP",
		Description: "Pula",
		Symbol:      "BWP",
		Precision:   2,
		Type:        Banknote},
	84: &Security{
		SecurityId:  84,
		Name:        "BZD",
		Description: "Belize Dollar",
		Symbol:      "BZD",
		Precision:   2,
		Type:        Banknote},
	90: &Security{
		SecurityId:  90,
		Name:        "SBD",
		Description: "Solomon Islands Dollar",
		Symbol:      "SBD",
		Precision:   2,
		Type:        Banknote},
	96: &Security{
		SecurityId:  96,
		Name:        "BND",
		Description: "Brunei Dollar",
		Symbol:      "BND",
		Precision:   2,
		Type:        Banknote},
	104: &Security{
		SecurityId:  104,
		Name:        "MMK",
		Description: "Kyat",
		Symbol:      "MMK",
		Precision:   2,
		Type:        Banknote},
	108: &Security{
		SecurityId:  108,
		Name:        "BIF",
		Description: "Burundi Franc",
		Symbol:      "BIF",
		Precision:   0,
		Type:        Banknote},
	116: &Security{
		SecurityId:  116,
		Name:        "KHR",
		Description: "Riel",
		Symbol:      "KHR",
		Precision:   2,
		Type:        Banknote},
	124: &Security{
		SecurityId:  124,
		Name:        "CAD",
		Description: "Canadian Dollar",
		Symbol:      "CAD",
		Precision:   2,
		Type:        Banknote},
	132: &Security{
		SecurityId:  132,
		Name:        "CVE",
		Description: "Cabo Verde Escudo",
		Symbol:      "CVE",
		Precision:   2,
		Type:        Banknote},
	136: &Security{
		SecurityId:  136,
		Name:        "KYD",
		Description: "Cayman Islands Dollar",
		Symbol:      "KYD",
		Precision:   2,
		Type:        Banknote},
	144: &Security{
		SecurityId:  144,
		Name:        "LKR",
		Description: "Sri Lanka Rupee",
		Symbol:      "LKR",
		Precision:   2,
		Type:        Banknote},
	152: &Security{
		SecurityId:  152,
		Name:        "CLP",
		Description: "Chilean Peso",
		Symbol:      "CLP",
		Precision:   0,
		Type:        Banknote},
	156: &Security{
		SecurityId:  156,
		Name:        "CNY",
		Description: "Yuan Renminbi",
		Symbol:      "CNY",
		Precision:   2,
		Type:        Banknote},
	170: &Security{
		SecurityId:  170,
		Name:        "COP",
		Description: "Colombian Peso",
		Symbol:      "COP",
		Precision:   2,
		Type:        Banknote},
	174: &Security{
		SecurityId:  174,
		Name:        "KMF",
		Description: "Comoro Franc",
		Symbol:      "KMF",
		Precision:   0,
		Type:        Banknote},
	188: &Security{
		SecurityId:  188,
		Name:        "CRC",
		Description: "Costa Rican Colon",
		Symbol:      "CRC",
		Precision:   2,
		Type:        Banknote},
	191: &Security{
		SecurityId:  191,
		Name:        "HRK",
		Description: "Kuna",
		Symbol:      "HRK",
		Precision:   2,
		Type:        Banknote},
	192: &Security{
		SecurityId:  192,
		Name:        "CUP",
		Description: "Cuban Peso",
		Symbol:      "CUP",
		Precision:   2,
		Type:        Banknote},
	203: &Security{
		SecurityId:  203,
		Name:        "CZK",
		Description: "Czech Koruna",
		Symbol:      "CZK",
		Precision:   2,
		Type:        Banknote},
	208: &Security{
		SecurityId:  208,
		Name:        "DKK",
		Description: "Danish Krone",
		Symbol:      "DKK",
		Precision:   2,
		Type:        Banknote},
	214: &Security{
		SecurityId:  214,
		Name:        "DOP",
		Description: "Dominican Peso",
		Symbol:      "DOP",
		Precision:   2,
		Type:        Banknote},
	222: &Security{
		SecurityId:  222,
		Name:        "SVC",
		Description: "El Salvador Colon",
		Symbol:      "SVC",
		Precision:   2,
		Type:        Banknote},
	230: &Security{
		SecurityId:  230,
		Name:        "ETB",
		Description: "Ethiopian Birr",
		Symbol:      "ETB",
		Precision:   2,
		Type:        Banknote},
	232: &Security{
		SecurityId:  232,
		Name:        "ERN",
		Description: "Nakfa",
		Symbol:      "ERN",
		Precision:   2,
		Type:        Banknote},
	238: &Security{
		SecurityId:  238,
		Name:        "FKP",
		Description: "Falkland Islands Pound",
		Symbol:      "FKP",
		Precision:   2,
		Type:        Banknote},
	242: &Security{
		SecurityId:  242,
		Name:        "FJD",
		Description: "Fiji Dollar",
		Symbol:      "FJD",
		Precision:   2,
		Type:        Banknote},
	262: &Security{
		SecurityId:  262,
		Name:        "DJF",
		Description: "Djibouti Franc",
		Symbol:      "DJF",
		Precision:   0,
		Type:        Banknote},
	270: &Security{
		SecurityId:  270,
		Name:        "GMD",
		Description: "Dalasi",
		Symbol:      "GMD",
		Precision:   2,
		Type:        Banknote},
	292: &Security{
		SecurityId:  292,
		Name:        "GIP",
		Description: "Gibraltar Pound",
		Symbol:      "GIP",
		Precision:   2,
		Type:        Banknote},
	320: &Security{
		SecurityId:  320,
		Name:        "GTQ",
		Description: "Quetzal",
		Symbol:      "GTQ",
		Precision:   2,
		Type:        Banknote},
	324: &Security{
		SecurityId:  324,
		Name:        "GNF",
		Description: "Guinea Franc",
		Symbol:      "GNF",
		Precision:   0,
		Type:        Banknote},
	328: &Security{
		SecurityId:  328,
		Name:        "GYD",
		Description: "Guyana Dollar",
		Symbol:      "GYD",
		Precision:   2,
		Type:        Banknote},
	332: &Security{
		SecurityId:  332,
		Name:        "HTG",
		Description: "Gourde",
		Symbol:      "HTG",
		Precision:   2,
		Type:        Banknote},
	340: &Security{
		SecurityId:  340,
		Name:        "HNL",
		Description: "Lempira",
		Symbol:      "HNL",
		Precision:   2,
		Type:        Banknote},
	344: &Security{
		SecurityId:  344,
		Name:        "HKD",
		Description: "Hong Kong Dollar",
		Symbol:      "HKD",
		Precision:   2,
		Type:        Banknote},
	348: &Security{
		SecurityId:  348,
		Name:        "HUF",
		Description: "Forint",
		Symbol:      "HUF",
		Precision:   2,
		Type:        Banknote},
	352: &Security{
		SecurityId:  352,
		Name:        "ISK",
		Description: "Iceland Krona",
		Symbol:      "ISK",
		Precision:   0,
		Type:        Banknote},
	356: &Security{
		SecurityId:  356,
		Name:        "INR",
		Description: "Indian Rupee",
		Symbol:      "INR",
		Precision:   2,
		Type:        Banknote},
	360: &Security{
		SecurityId:  360,
		Name:        "IDR",
		Description: "Rupiah",
		Symbol:      "IDR",
		Precision:   2,
		Type:        Banknote},
	364: &Security{
		SecurityId:  364,
		Name:        "IRR",
		Description: "Iranian Rial",
		Symbol:      "IRR",
		Precision:   2,
		Type:        Banknote},
	368: &Security{
		SecurityId:  368,
		Name:        "IQD",
		Description: "Iraqi Dinar",
		Symbol:      "IQD",
		Precision:   3,
		Type:        Banknote},
	376: &Security{
		SecurityId:  376,
		Name:        "ILS",
		Description: "New Israeli Sheqel",
		Symbol:      "ILS",
		Precision:   2,
		Type:        Banknote},
	388: &Security{
		SecurityId:  388,
		Name:        "JMD",
		Description: "Jamaican Dollar",
		Symbol:      "JMD",
		Precision:   2,
		Type:        Banknote},
	392: &Security{
		SecurityId:  392,
		Name:        "JPY",
		Description: "Yen",
		Symbol:      "JPY",
		Precision:   0,
		Type:        Banknote},
	398: &Security{
		SecurityId:  398,
		Name:        "KZT",
		Description: "Tenge",
		Symbol:      "KZT",
		Precision:   2,
		Type:        Banknote},
	400: &Security{
		SecurityId:  400,
		Name:        "JOD",
		Description: "Jordanian Dinar",
		Symbol:      "JOD",
		Precision:   3,
		Type:        Banknote},
	404: &Security{
		SecurityId:  404,
		Name:        "KES",
		Description: "Kenyan Shilling",
		Symbol:      "KES",
		Precision:   2,
		Type:        Banknote},
	408: &Security{
		SecurityId:  408,
		Name:        "KPW",
		Description: "North Korean Won",
		Symbol:      "KPW",
		Precision:   2,
		Type:        Banknote},
	410: &Security{
		SecurityId:  410,
		Name:        "KRW",
		Description: "Won",
		Symbol:      "KRW",
		Precision:   0,
		Type:        Banknote},
	414: &Security{
		SecurityId:  414,
		Name:        "KWD",
		Description: "Kuwaiti Dinar",
		Symbol:      "KWD",
		Precision:   3,
		Type:        Banknote},
	417: &Security{
		SecurityId:  417,
		Name:        "KGS",
		Description: "Som",
		Symbol:      "KGS",
		Precision:   2,
		Type:        Banknote},
	418: &Security{
		SecurityId:  418,
		Name:        "LAK",
		Description: "Kip",
		Symbol:      "LAK",
		Precision:   2,
		Type:        Banknote},
	422: &Security{
		SecurityId:  422,
		Name:        "LBP",
		Description: "Lebanese Pound",
		Symbol:      "LBP",
		Precision:   2,
		Type:        Banknote},
	426: &Security{
		SecurityId:  426,
		Name:        "LSL",
		Description: "Loti",
		Symbol:      "LSL",
		Precision:   2,
		Type:        Banknote},
	430: &Security{
		SecurityId:  430,
		Name:        "LRD",
		Description: "Liberian Dollar",
		Symbol:      "LRD",
		Precision:   2,
		Type:        Banknote},
	434: &Security{
		SecurityId:  434,
		Name:        "LYD",
		Description: "Libyan Dinar",
		Symbol:      "LYD",
		Precision:   3,
		Type:        Banknote},
	446: &Security{
		SecurityId:  446,
		Name:        "MOP",
		Description: "Pataca",
		Symbol:      "MOP",
		Precision:   2,
		Type:        Banknote},
	454: &Security{
		SecurityId:  454,
		Name:        "MWK",
		Description: "Kwacha",
		Symbol:      "MWK",
		Precision:   2,
		Type:        Banknote},
	458: &Security{
		SecurityId:  458,
		Name:        "MYR",
		Description: "Malaysian Ringgit",
		Symbol:      "MYR",
		Precision:   2,
		Type:        Banknote},
	462: &Security{
		SecurityId:  462,
		Name:        "MVR",
		Description: "Rufiyaa",
		Symbol:      "MVR",
		Precision:   2,
		Type:        Banknote},
	478: &Security{
		SecurityId:  478,
		Name:        "MRO",
		Description: "Ouguiya",
		Symbol:      "MRO",
		Precision:   2,
		Type:        Banknote},
	480: &Security{
		SecurityId:  480,
		Name:        "MUR",
		Description: "Mauritius Rupee",
		Symbol:      "MUR",
		Precision:   2,
		Type:        Banknote},
	484: &Security{
		SecurityId:  484,
		Name:        "MXN",
		Description: "Mexican Peso",
		Symbol:      "MXN",
		Precision:   2,
		Type:        Banknote},
	496: &Security{
		SecurityId:  496,
		Name:        "MNT",
		Description: "Tugrik",
		Symbol:      "MNT",
		Precision:   2,
		Type:        Banknote},
	498: &Security{
		SecurityId:  498,
		Name:        "MDL",
		Description: "Moldovan Leu",
		Symbol:      "MDL",
		Precision:   2,
		Type:        Banknote},
	504: &Security{
		SecurityId:  504,
		Name:        "MAD",
		Description: "Moroccan Dirham",
		Symbol:      "MAD",
		Precision:   2,
		Type:        Banknote},
	512: &Security{
		SecurityId:  512,
		Name:        "OMR",
		Description: "Rial Omani",
		Symbol:      "OMR",
		Precision:   3,
		Type:        Banknote},
	516: &Security{
		SecurityId:  516,
		Name:        "NAD",
		Description: "Namibia Dollar",
		Symbol:      "NAD",
		Precision:   2,
		Type:        Banknote},
	524: &Security{
		SecurityId:  524,
		Name:        "NPR",
		Description: "Nepalese Rupee",
		Symbol:      "NPR",
		Precision:   2,
		Type:        Banknote},
	532: &Security{
		SecurityId:  532,
		Name:        "ANG",
		Description: "Netherlands Antillean Guilder",
		Symbol:      "ANG",
		Precision:   2,
		Type:        Banknote},
	533: &Security{
		SecurityId:  533,
		Name:        "AWG",
		Description: "Aruban Florin",
		Symbol:      "AWG",
		Precision:   2,
		Type:        Banknote},
	548: &Security{
		SecurityId:  548,
		Name:        "VUV",
		Description: "Vatu",
		Symbol:      "VUV",
		Precision:   0,
		Type:        Banknote},
	554: &Security{
		SecurityId:  554,
		Name:        "NZD",
		Description: "New Zealand Dollar",
		Symbol:      "NZD",
		Precision:   2,
		Type:        Banknote},
	558: &Security{
		SecurityId:  558,
		Name:        "NIO",
		Description: "Cordoba Oro",
		Symbol:      "NIO",
		Precision:   2,
		Type:        Banknote},
	566: &Security{
		SecurityId:  566,
		Name:        "NGN",
		Description: "Naira",
		Symbol:      "NGN",
		Precision:   2,
		Type:        Banknote},
	578: &Security{
		SecurityId:  578,
		Name:        "NOK",
		Description: "Norwegian Krone",
		Symbol:      "NOK",
		Precision:   2,
		Type:        Banknote},
	586: &Security{
		SecurityId:  586,
		Name:        "PKR",
		Description: "Pakistan Rupee",
		Symbol:      "PKR",
		Precision:   2,
		Type:        Banknote},
	590: &Security{
		SecurityId:  590,
		Name:        "PAB",
		Description: "Balboa",
		Symbol:      "PAB",
		Precision:   2,
		Type:        Banknote},
	598: &Security{
		SecurityId:  598,
		Name:        "PGK",
		Description: "Kina",
		Symbol:      "PGK",
		Precision:   2,
		Type:        Banknote},
	600: &Security{
		SecurityId:  600,
		Name:        "PYG",
		Description: "Guarani",
		Symbol:      "PYG",
		Precision:   0,
		Type:        Banknote},
	604: &Security{
		SecurityId:  604,
		Name:        "PEN",
		Description: "Nuevo Sol",
		Symbol:      "PEN",
		Precision:   2,
		Type:        Banknote},
	608: &Security{
		SecurityId:  608,
		Name:        "PHP",
		Description: "Philippine Peso",
		Symbol:      "PHP",
		Precision:   2,
		Type:        Banknote},
	634: &Security{
		SecurityId:  634,
		Name:        "QAR",
		Description: "Qatari Rial",
		Symbol:      "QAR",
		Precision:   2,
		Type:        Banknote},
	643: &Security{
		SecurityId:  643,
		Name:        "RUB",
		Description: "Russian Ruble",
		Symbol:      "RUB",
		Precision:   2,
		Type:        Banknote},
	646: &Security{
		SecurityId:  646,
		Name:        "RWF",
		Description: "Rwanda Franc",
		Symbol:      "RWF",
		Precision:   0,
		Type:        Banknote},
	654: &Security{
		SecurityId:  654,
		Name:        "SHP",
		Description: "Saint Helena Pound",
		Symbol:      "SHP",
		Precision:   2,
		Type:        Banknote},
	678: &Security{
		SecurityId:  678,
		Name:        "STD",
		Description: "Dobra",
		Symbol:      "STD",
		Precision:   2,
		Type:        Banknote},
	682: &Security{
		SecurityId:  682,
		Name:        "SAR",
		Description: "Saudi Riyal",
		Symbol:      "SAR",
		Precision:   2,
		Type:        Banknote},
	690: &Security{
		SecurityId:  690,
		Name:        "SCR",
		Description: "Seychelles Rupee",
		Symbol:      "SCR",
		Precision:   2,
		Type:        Banknote},
	694: &Security{
		SecurityId:  694,
		Name:        "SLL",
		Description: "Leone",
		Symbol:      "SLL",
		Precision:   2,
		Type:        Banknote},
	702: &Security{
		SecurityId:  702,
		Name:        "SGD",
		Description: "Singapore Dollar",
		Symbol:      "SGD",
		Precision:   2,
		Type:        Banknote},
	704: &Security{
		SecurityId:  704,
		Name:        "VND",
		Description: "Dong",
		Symbol:      "VND",
		Precision:   0,
		Type:        Banknote},
	706: &Security{
		SecurityId:  706,
		Name:        "SOS",
		Description: "Somali Shilling",
		Symbol:      "SOS",
		Precision:   2,
		Type:        Banknote},
	710: &Security{
		SecurityId:  710,
		Name:        "ZAR",
		Description: "Rand",
		Symbol:      "ZAR",
		Precision:   2,
		Type:        Banknote},
	728: &Security{
		SecurityId:  728,
		Name:        "SSP",
		Description: "South Sudanese Pound",
		Symbol:      "SSP",
		Precision:   2,
		Type:        Banknote},
	748: &Security{
		SecurityId:  748,
		Name:        "SZL",
		Description: "Lilangeni",
		Symbol:      "SZL",
		Precision:   2,
		Type:        Banknote},
	752: &Security{
		SecurityId:  752,
		Name:        "SEK",
		Description: "Swedish Krona",
		Symbol:      "SEK",
		Precision:   2,
		Type:        Banknote},
	756: &Security{
		SecurityId:  756,
		Name:        "CHF",
		Description: "Swiss Franc",
		Symbol:      "CHF",
		Precision:   2,
		Type:        Banknote},
	760: &Security{
		SecurityId:  760,
		Name:        "SYP",
		Description: "Syrian Pound",
		Symbol:      "SYP",
		Precision:   2,
		Type:        Banknote},
	764: &Security{
		SecurityId:  764,
		Name:        "THB",
		Description: "Baht",
		Symbol:      "THB",
		Precision:   2,
		Type:        Banknote},
	776: &Security{
		SecurityId:  776,
		Name:        "TOP",
		Description: "Pa’anga",
		Symbol:      "TOP",
		Precision:   2,
		Type:        Banknote},
	780: &Security{
		SecurityId:  780,
		Name:        "TTD",
		Description: "Trinidad and Tobago Dollar",
		Symbol:      "TTD",
		Precision:   2,
		Type:        Banknote},
	784: &Security{
		SecurityId:  784,
		Name:        "AED",
		Description: "UAE Dirham",
		Symbol:      "AED",
		Precision:   2,
		Type:        Banknote},
	788: &Security{
		SecurityId:  788,
		Name:        "TND",
		Description: "Tunisian Dinar",
		Symbol:      "TND",
		Precision:   3,
		Type:        Banknote},
	800: &Security{
		SecurityId:  800,
		Name:        "UGX",
		Description: "Uganda Shilling",
		Symbol:      "UGX",
		Precision:   0,
		Type:        Banknote},
	807: &Security{
		SecurityId:  807,
		Name:        "MKD",
		Description: "Denar",
		Symbol:      "MKD",
		Precision:   2,
		Type:        Banknote},
	818: &Security{
		SecurityId:  818,
		Name:        "EGP",
		Description: "Egyptian Pound",
		Symbol:      "EGP",
		Precision:   2,
		Type:        Banknote},
	826: &Security{
		SecurityId:  826,
		Name:        "GBP",
		Description: "Pound Sterling",
		Symbol:      "GBP",
		Precision:   2,
		Type:        Banknote},
	834: &Security{
		SecurityId:  834,
		Name:        "TZS",
		Description: "Tanzanian Shilling",
		Symbol:      "TZS",
		Precision:   2,
		Type:        Banknote},
	840: &Security{
		SecurityId:  840,
		Name:        "USD",
		Description: "US Dollar",
		Symbol:      "USD",
		Precision:   2,
		Type:        Banknote},
	858: &Security{
		SecurityId:  858,
		Name:        "UYU",
		Description: "Peso Uruguayo",
		Symbol:      "UYU",
		Precision:   2,
		Type:        Banknote},
	860: &Security{
		SecurityId:  860,
		Name:        "UZS",
		Description: "Uzbekistan Sum",
		Symbol:      "UZS",
		Precision:   2,
		Type:        Banknote},
	882: &Security{
		SecurityId:  882,
		Name:        "WST",
		Description: "Tala",
		Symbol:      "WST",
		Precision:   2,
		Type:        Banknote},
	886: &Security{
		SecurityId:  886,
		Name:        "YER",
		Description: "Yemeni Rial",
		Symbol:      "YER",
		Precision:   2,
		Type:        Banknote},
	901: &Security{
		SecurityId:  901,
		Name:        "TWD",
		Description: "New Taiwan Dollar",
		Symbol:      "TWD",
		Precision:   2,
		Type:        Banknote},
	931: &Security{
		SecurityId:  931,
		Name:        "CUC",
		Description: "Peso Convertible",
		Symbol:      "CUC",
		Precision:   2,
		Type:        Banknote},
	932: &Security{
		SecurityId:  932,
		Name:        "ZWL",
		Description: "Zimbabwe Dollar",
		Symbol:      "ZWL",
		Precision:   2,
		Type:        Banknote},
	934: &Security{
		SecurityId:  934,
		Name:        "TMT",
		Description: "Turkmenistan New Manat",
		Symbol:      "TMT",
		Precision:   2,
		Type:        Banknote},
	936: &Security{
		SecurityId:  936,
		Name:        "GHS",
		Description: "Ghana Cedi",
		Symbol:      "GHS",
		Precision:   2,
		Type:        Banknote},
	937: &Security{
		SecurityId:  937,
		Name:        "VEF",
		Description: "Bolivar",
		Symbol:      "VEF",
		Precision:   2,
		Type:        Banknote},
	938: &Security{
		SecurityId:  938,
		Name:        "SDG",
		Description: "Sudanese Pound",
		Symbol:      "SDG",
		Precision:   2,
		Type:        Banknote},
	940: &Security{
		SecurityId:  940,
		Name:        "UYI",
		Description: "Uruguay Peso en Unidades Indexadas (URUIURUI)",
		Symbol:      "UYI",
		Precision:   0,
		Type:        Banknote},
	941: &Security{
		SecurityId:  941,
		Name:        "RSD",
		Description: "Serbian Dinar",
		Symbol:      "RSD",
		Precision:   2,
		Type:        Banknote},
	943: &Security{
		SecurityId:  943,
		Name:        "MZN",
		Description: "Mozambique Metical",
		Symbol:      "MZN",
		Precision:   2,
		Type:        Banknote},
	944: &Security{
		SecurityId:  944,
		Name:        "AZN",
		Description: "Azerbaijanian Manat",
		Symbol:      "AZN",
		Precision:   2,
		Type:        Banknote},
	946: &Security{
		SecurityId:  946,
		Name:        "RON",
		Description: "Romanian Leu",
		Symbol:      "RON",
		Precision:   2,
		Type:        Banknote},
	947: &Security{
		SecurityId:  947,
		Name:        "CHE",
		Description: "WIR Euro",
		Symbol:      "CHE",
		Precision:   2,
		Type:        Banknote},
	948: &Security{
		SecurityId:  948,
		Name:        "CHW",
		Description: "WIR Franc",
		Symbol:      "CHW",
		Precision:   2,
		Type:        Banknote},
	949: &Security{
		SecurityId:  949,
		Name:        "TRY",
		Description: "Turkish Lira",
		Symbol:      "TRY",
		Precision:   2,
		Type:        Banknote},
	950: &Security{
		SecurityId:  950,
		Name:        "XAF",
		Description: "CFA Franc BEAC",
		Symbol:      "XAF",
		Precision:   0,
		Type:        Banknote},
	951: &Security{
		SecurityId:  951,
		Name:        "XCD",
		Description: "East Caribbean Dollar",
		Symbol:      "XCD",
		Precision:   2,
		Type:        Banknote},
	952: &Security{
		SecurityId:  952,
		Name:        "XOF",
		Description: "CFA Franc BCEAO",
		Symbol:      "XOF",
		Precision:   0,
		Type:        Banknote},
	953: &Security{
		SecurityId:  953,
		Name:        "XPF",
		Description: "CFP Franc",
		Symbol:      "XPF",
		Precision:   0,
		Type:        Banknote},
	955: &Security{
		SecurityId:  955,
		Name:        "XBA",
		Description: "Bond Markets Unit European Composite Unit (EURCO)",
		Symbol:      "XBA",
		Precision:   0,
		Type:        Banknote},
	956: &Security{
		SecurityId:  956,
		Name:        "XBB",
		Description: "Bond Markets Unit European Monetary Unit (E.M.U.-6)",
		Symbol:      "XBB",
		Precision:   0,
		Type:        Banknote},
	957: &Security{
		SecurityId:  957,
		Name:        "XBC",
		Description: "Bond Markets Unit European Unit of Account 9 (E.U.A.-9)",
		Symbol:      "XBC",
		Precision:   0,
		Type:        Banknote},
	958: &Security{
		SecurityId:  958,
		Name:        "XBD",
		Description: "Bond Markets Unit European Unit of Account 17 (E.U.A.-17)",
		Symbol:      "XBD",
		Precision:   0,
		Type:        Banknote},
	959: &Security{
		SecurityId:  959,
		Name:        "XAU",
		Description: "Gold",
		Symbol:      "XAU",
		Precision:   0,
		Type:        Banknote},
	960: &Security{
		SecurityId:  960,
		Name:        "XDR",
		Description: "SDR (Special Drawing Right)",
		Symbol:      "XDR",
		Precision:   0,
		Type:        Banknote},
	961: &Security{
		SecurityId:  961,
		Name:        "XAG",
		Description: "Silver",
		Symbol:      "XAG",
		Precision:   0,
		Type:        Banknote},
	962: &Security{
		SecurityId:  962,
		Name:        "XPT",
		Description: "Platinum",
		Symbol:      "XPT",
		Precision:   0,
		Type:        Banknote},
	963: &Security{
		SecurityId:  963,
		Name:        "XTS",
		Description: "Codes specifically reserved for testing purposes",
		Symbol:      "XTS",
		Precision:   0,
		Type:        Banknote},
	964: &Security{
		SecurityId:  964,
		Name:        "XPD",
		Description: "Palladium",
		Symbol:      "XPD",
		Precision:   0,
		Type:        Banknote},
	965: &Security{
		SecurityId:  965,
		Name:        "XUA",
		Description: "ADB Unit of Account",
		Symbol:      "XUA",
		Precision:   0,
		Type:        Banknote},
	967: &Security{
		SecurityId:  967,
		Name:        "ZMW",
		Description: "Zambian Kwacha",
		Symbol:      "ZMW",
		Precision:   2,
		Type:        Banknote},
	968: &Security{
		SecurityId:  968,
		Name:        "SRD",
		Description: "Surinam Dollar",
		Symbol:      "SRD",
		Precision:   2,
		Type:        Banknote},
	969: &Security{
		SecurityId:  969,
		Name:        "MGA",
		Description: "Malagasy Ariary",
		Symbol:      "MGA",
		Precision:   2,
		Type:        Banknote},
	970: &Security{
		SecurityId:  970,
		Name:        "COU",
		Description: "Unidad de Valor Real",
		Symbol:      "COU",
		Precision:   2,
		Type:        Banknote},
	971: &Security{
		SecurityId:  971,
		Name:        "AFN",
		Description: "Afghani",
		Symbol:      "AFN",
		Precision:   2,
		Type:        Banknote},
	972: &Security{
		SecurityId:  972,
		Name:        "TJS",
		Description: "Somoni",
		Symbol:      "TJS",
		Precision:   2,
		Type:        Banknote},
	973: &Security{
		SecurityId:  973,
		Name:        "AOA",
		Description: "Kwanza",
		Symbol:      "AOA",
		Precision:   2,
		Type:        Banknote},
	974: &Security{
		SecurityId:  974,
		Name:        "BYR",
		Description: "Belarussian Ruble",
		Symbol:      "BYR",
		Precision:   0,
		Type:        Banknote},
	975: &Security{
		SecurityId:  975,
		Name:        "BGN",
		Description: "Bulgarian Lev",
		Symbol:      "BGN",
		Precision:   2,
		Type:        Banknote},
	976: &Security{
		SecurityId:  976,
		Name:        "CDF",
		Description: "Congolese Franc",
		Symbol:      "CDF",
		Precision:   2,
		Type:        Banknote},
	977: &Security{
		SecurityId:  977,
		Name:        "BAM",
		Description: "Convertible Mark",
		Symbol:      "BAM",
		Precision:   2,
		Type:        Banknote},
	978: &Security{
		SecurityId:  978,
		Name:        "EUR",
		Description: "Euro",
		Symbol:      "EUR",
		Precision:   2,
		Type:        Banknote},
	979: &Security{
		SecurityId:  979,
		Name:        "MXV",
		Description: "Mexican Unidad de Inversion (UDI)",
		Symbol:      "MXV",
		Precision:   2,
		Type:        Banknote},
	980: &Security{
		SecurityId:  980,
		Name:        "UAH",
		Description: "Hryvnia",
		Symbol:      "UAH",
		Precision:   2,
		Type:        Banknote},
	981: &Security{
		SecurityId:  981,
		Name:        "GEL",
		Description: "Lari",
		Symbol:      "GEL",
		Precision:   2,
		Type:        Banknote},
	984: &Security{
		SecurityId:  984,
		Name:        "BOV",
		Description: "Mvdol",
		Symbol:      "BOV",
		Precision:   2,
		Type:        Banknote},
	985: &Security{
		SecurityId:  985,
		Name:        "PLN",
		Description: "Zloty",
		Symbol:      "PLN",
		Precision:   2,
		Type:        Banknote},
	986: &Security{
		SecurityId:  986,
		Name:        "BRL",
		Description: "Brazilian Real",
		Symbol:      "BRL",
		Precision:   2,
		Type:        Banknote},
	990: &Security{
		SecurityId:  990,
		Name:        "CLF",
		Description: "Unidad de Fomento",
		Symbol:      "CLF",
		Precision:   4,
		Type:        Banknote},
	994: &Security{
		SecurityId:  994,
		Name:        "XSU",
		Description: "Sucre",
		Symbol:      "XSU",
		Precision:   0,
		Type:        Banknote},
	997: &Security{
		SecurityId:  997,
		Name:        "USN",
		Description: "US Dollar (Next day)",
		Symbol:      "USN",
		Precision:   2,
		Type:        Banknote},
	999: &Security{
		SecurityId:  999,
		Name:        "XXX",
		Description: "The codes assigned for transactions where no currency is involved",
		Symbol:      "XXX",
		Precision:   0,
		Type:        Banknote},

	// Securities
	1000: &Security{
		SecurityId:  1000,
		Name:        "SPY",
		Description: "S&P 500 ETF Fund",
		Symbol:      "SPY",
		Precision:   5,
		Type:        Stock},
}

var security_list []*Security

func init() {
	for _, value := range security_map {
		security_list = append(security_list, value)
	}
}

func GetSecurity(securityid int64) *Security {
	s := security_map[securityid]
	if s != nil {
		return s
	}
	return nil
}

func GetSecurityByName(name string) (*Security, error) {
	for _, value := range security_map {
		if value.Name == name {
			return value, nil
		}
	}
	return nil, errors.New("Invalid Security Name")
}

func GetSecurities() []*Security {
	return security_list
}

func (s *Security) Write(w http.ResponseWriter) error {
	enc := json.NewEncoder(w)
	return enc.Encode(s)
}

func (sl *SecurityList) Write(w http.ResponseWriter) error {
	enc := json.NewEncoder(w)
	return enc.Encode(sl)
}

func SecurityHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		securityid, err := GetURLID(r.URL.Path)
		if err == nil {
			security := GetSecurity(securityid)
			if security == nil {
				WriteError(w, 3 /*Invalid Request*/)
				return
			}
			err := security.Write(w)
			if err != nil {
				WriteError(w, 999 /*Internal Error*/)
				log.Print(err)
				return
			}
		} else {
			var sl SecurityList
			securities := GetSecurities()
			sl.Securities = &securities
			err := (&sl).Write(w)
			if err != nil {
				WriteError(w, 999 /*Internal Error*/)
				log.Print(err)
				return
			}
		}
	} else {
		WriteError(w, 3 /*Invalid Request*/)
	}
}
