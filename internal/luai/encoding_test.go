package luai

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

type testStruct struct {
	Field1 string
	Field2 int
	Field3 bool
}

func TestRegular(t *testing.T) {
	luaVm := lua.NewState()

	test := testStruct{
		Field1: "test",
		Field2: 1,
		Field3: true,
	}

	_table, err := Marshal(luaVm, &test)
	if err != nil {
		t.Fatal(err)
	}

	table := _table.(*lua.LTable)

	field1 := table.RawGetString("Field1")
	if field1.Type() != lua.LTString {
		t.Errorf("expected string, got %s", field1.Type())
	}

	if field1.String() != "test" {
		t.Errorf("expected 'test', got '%s'", field1.String())
	}

	field2 := table.RawGetString("Field2")
	if field2.Type() != lua.LTNumber {
		t.Errorf("expected number, got %s", field2.Type())
	}

	if field2.String() != "1" {
		t.Errorf("expected '1', got '%s'", field2.String())
	}

	field3 := table.RawGetString("Field3")
	if field3.Type() != lua.LTBool {
		t.Errorf("expected bool, got %s", field3.Type())
	}

	if field3.String() != "true" {
		t.Errorf("expected 'true', got '%s'", field3.String())
	}

	struct2 := testStruct{}
	err = Unmarshal(table, &struct2)
	if err != nil {
		t.Fatal(err)
	}

	if struct2.Field1 != "test" {
		t.Errorf("expected 'test', got '%s'", struct2.Field1)
	}

	if struct2.Field2 != 1 {
		t.Errorf("expected 1, got %d", struct2.Field2)
	}

	if struct2.Field3 != true {
		t.Errorf("expected true, got %t", struct2.Field3)
	}
}

type testStructTag struct {
	Field1 string `luai:"field1"`
	Field2 int    `luai:"field2"`
	Field3 bool   `luai:"field3"`
}

func TestTag(t *testing.T) {
	luaVm := lua.NewState()

	test := testStructTag{
		Field1: "test",
		Field2: 1,
		Field3: true,
	}

	_table, err := Marshal(luaVm, &test)
	if err != nil {
		t.Fatal(err)
	}

	table := _table.(*lua.LTable)

	field1 := table.RawGetString("field1")
	if field1.Type() != lua.LTString {
		t.Errorf("expected string, got %s", field1.Type())
	}

	if field1.String() != "test" {
		t.Errorf("expected 'test', got '%s'", field1.String())
	}

	field2 := table.RawGetString("field2")
	if field2.Type() != lua.LTNumber {
		t.Errorf("expected number, got %s", field2.Type())
	}

	if field2.String() != "1" {
		t.Errorf("expected '1', got '%s'", field2.String())
	}

	field3 := table.RawGetString("field3")
	if field3.Type() != lua.LTBool {
		t.Errorf("expected bool, got %s", field3.Type())
	}

	if field3.String() != "true" {
		t.Errorf("expected 'true', got '%s'", field3.String())
	}

	struct2 := testStructTag{}
	err = Unmarshal(table, &struct2)
	if err != nil {
		t.Fatal(err)
	}

	if struct2.Field1 != "test" {
		t.Errorf("expected 'test', got '%s'", struct2.Field1)
	}

	if struct2.Field2 != 1 {
		t.Errorf("expected 1, got %d", struct2.Field2)
	}

	if struct2.Field3 != true {
		t.Errorf("expected true, got %t", struct2.Field3)
	}
}
