package integration_test

import (
	"testing"
)

func TestLuaDates(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		simpleLuaTest(t, d.clients[0], []LuaTest{
			{"Year", "return date.new('0009-01-03').Year", "9"},
			{"Month", "return date.new('3999-02-01').Month", "2"},
			{"Day", "return date.new('1997-12-31').Day", "31"},
			{"__tostring", "return date.new('0997-12-01')", "0997-12-01"},
			{"__tostring 2", "return date.new(997, 12, 1)", "0997-12-01"},
			{"__tostring 3", "return date.new({year=997, month=12, day=1})", "0997-12-01"},
			{"__eq", "return date.new('2017-10-05') == date.new(2017, 10, 5)", "true"},
			{"(not) __eq", "return date.new('0997-12-01') == date.new('1997-12-01')", "false"},
			{"__lt", "return date.new('0997-12-01') < date.new('1997-12-01')", "true"},
			{"(not) __lt", "return date.new('2001-12-01') < date.new('1997-12-01')", "false"},
			{"not __lt self", "return date.new('2001-12-01') < date.new('2001-12-01')", "false"},
			{"__le", "return date.new('0997-12-01') <= date.new('1997-12-01')", "true"},
			{"(not) __le", "return date.new(2001, 12, 1) <= date.new('1997-12-01')", "false"},
			{"__le self", "return date.new('2001-12-01') <= date.new(2001, 12, 1)", "true"},
			{"__add", "return date.new('2001-12-30') + date.new({year=0, month=0, day=1})", "2001-12-31"},
			{"__add", "return date.new('2001-12-30') + date.new({year=0, month=0, day=2})", "2002-01-01"},
			{"__sub", "return date.new('2001-12-30') - date.new({year=1, month=1, day=1})", "2000-11-29"},
			{"__sub", "return date.new('2058-03-01') - date.new({year=0, month=0, day=1})", "2058-02-28"},
			{"__sub", "return date.new('2058-03-31') - date.new({year=0, month=1, day=0})", "2058-02-28"},
		})
	})
}
