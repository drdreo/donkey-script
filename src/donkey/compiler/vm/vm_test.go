package vm

import (
	"donkey/compiler"
	"donkey/object"
	"donkey/utils"
	"fmt"
	"testing"
)

type vmTestCase struct {
	input    string
	expected interface{}
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []vmTestCase{
		{"1", 1},
		{"2", 2},
		{"1 + 2", 3},
		{"1 - 2", -1},
		{"1 * 2", 2},
		{"4 / 2", 2},
		{"50 / 2 * 2 + 10 - 5", 55},
		{"5 * (2 + 10)", 60},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"5 * (2 + 10)", 60},
		{"-5", -5},
		{"-10", -10},
		{"-50 + 100 + -50", 0},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
	}

	runVmTests(t, tests)
}

func TestBooleanExpressions(t *testing.T) {
	tests := []vmTestCase{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"1 <= 2", true},
		{"1 <= 1", true},
		{"1 >= 1", true},
		{"1 >= 2", false},
		{"true == true", true},
		{"true == 1", false},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
		{"(1 >= 2) == false", true},
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
		{"!(1 >= 1)", false},
		{"!(if(false){5})", true},
	}

	runVmTests(t, tests)
}

func TestConditionals(t *testing.T) {
	tests := []vmTestCase{
		{"if (true) { 10 }", 10},
		{"if (true) { 10 } else { 20 }", 10},
		{"if (false) { 10 } else { 20 } ", 20},
		{"if (1) { 10 }", 10},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 < 2) { 10 } else { 20 }", 10},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (1 > 2) { 10 }", Null},
		{"if (false) { 10 }", Null},
		{"if ((if (false) { 10 })) { 10 } else { 20 }", 20},
	}

	runVmTests(t, tests)
}

func TestGlobalLetStatements(t *testing.T) {
	tests := []vmTestCase{
		{"let one = 1; one", 1},
		{"let one = 1; let two = 2; one + two", 3},
		{"let one = 1; let two = one + one; one + two", 3},
	}

	runVmTests(t, tests)
}

func TestStringExpression(t *testing.T) {
	tests := []vmTestCase{
		{`"donkey"`, "donkey"},
		{`"don" + "key"`, "donkey"},
		{`"don" + "key" + " kong"`, "donkey kong"},
		{`"donkey" - "key"`, "don"},
		{`"donkey kong" - "d" - " "`, "onkeykong"},
		{`"hey" == "hey"`, true},
		{`"hey" != "hey"`, false},
		{`"hey" == "nej"`, false},
		{`"hey" != "nej"`, true},
	}

	runVmTests(t, tests)
}

func TestArrayLiterals(t *testing.T) {
	tests := []vmTestCase{
		{`[]`, []int{}},
		{`[1,2,3]`, []int{1, 2, 3}},
		{`[1+2, (3-4), 5*5]`, []int{3, -1, 25}},
	}

	runVmTests(t, tests)
}

func TestHashLiterals(t *testing.T) {
	tests := []vmTestCase{
		{`{}`, map[object.HashKey]int64{}},
		{`{1:2,3:4,5:6}`, map[object.HashKey]int64{
			(&object.Integer{Value: 1}).HashKey(): 2,
			(&object.Integer{Value: 3}).HashKey(): 4,
			(&object.Integer{Value: 5}).HashKey(): 6,
		}},
		{`{1+1 : 2*2, 3+3 : 4*4}`, map[object.HashKey]int64{
			(&object.Integer{Value: 2}).HashKey(): 4,
			(&object.Integer{Value: 6}).HashKey(): 16,
		}},
	}

	runVmTests(t, tests)
}

func TestIndexExpressions(t *testing.T) {
	tests := []vmTestCase{
		{`[1,2,3][1]`, 2},
		{`[1,2,3][0+2]`, 3},
		{`[[1,1,1]][0][0]`, 1},
		{`[][0]`, Null},
		{`[1,2,3][99]`, Null},
		{`[1][-1]`, 1},
		{`[1,2][-1]`, 2},

		{`{1:1,2:2}[1]`, 1},
		{`{1:1,2:2}[2]`, 2},
		{`{}[0]`, Null},
		{`{1:1}[0]`, Null},
		{`{1:2}[-1]`, Null},
	}

	runVmTests(t, tests)
}

func TestFunctionCallsWithoutArguments(t *testing.T) {
	tests := []vmTestCase{
		{`
		let fivePlusTen = fn(){ 5 + 10; }
		fivePlusTen();
		`,
			15,
		},
		{`
		let one = fn(){ 1 }
		let two = fn(){ 2 }
		one() + two();
		`,
			3,
		},
		{`
		let a = fn(){ 1 }
		let b = fn(){ a() + 1 }
		let c = fn(){ b() + 1 }
		c();
		`,
			3,
		},
		{`
		let earlyOut = fn(){ return 69; 420; }
		earlyOut();
		`,
			69,
		},
		{`
		let earlyOut = fn(){ return 69; return 420; }
		earlyOut();
		`,
			69,
		},
	}

	runVmTests(t, tests)
}

/**
TEST HELPERS
**/

func runVmTests(t *testing.T, tests []vmTestCase) {
	t.Helper()

	for _, tt := range tests {
		program := utils.ParseProgram(tt.input)

		comp := compiler.New()
		err := comp.Compile(program)
		if err != nil {
			t.Fatalf(utils.Blue("'%s'", tt.input) + utils.Red(" - Compiler Error: %s", err))
		}

		vm := New(comp.Bytecode())
		err = vm.Run()
		if err != nil {
			t.Fatalf(utils.Blue("'%s'", tt.input) + utils.Red(" - VM Error: %s", err))
		}

		stackElem := vm.LastPoppedStackElem()

		testExpectedObject(t, tt, stackElem)
	}
}

func testExpectedObject(
	t *testing.T,
	tC vmTestCase,
	actual object.Object,
) {
	t.Helper()

	switch expected := tC.expected.(type) {
	case int:
		err := testIntegerObject(int64(expected), actual)
		if err != nil {
			t.Errorf("%s - testIntegerObject failed: %s", utils.Blue(tC.input), err)
		}

	case bool:
		err := testBooleanObject(expected, actual)
		if err != nil {
			t.Errorf("%s - testBooleanObject failed: %s", utils.Blue(tC.input), err)
		}

	case string:
		err := testStringObject(expected, actual)
		if err != nil {
			t.Errorf("%s - testStringObject failed: %s", utils.Blue(tC.input), err)
		}

	case []int:
		err := testIntArrayObject(expected, actual)
		if err != nil {
			t.Errorf("%s - testIntArrayObject failed: %s", utils.Blue(tC.input), err)
		}
	case map[object.HashKey]int64:
		err := testIntHashObject(expected, actual)
		if err != nil {
			t.Errorf("%s - testIntArrayObject failed: %s", utils.Blue(tC.input), err)
		}
	case *object.Null:
		if actual != Null {
			t.Errorf("%s - object is not Null: %T (%+v)", utils.Blue(tC.input), actual, actual)
		}
	}
}

func testIntegerObject(expected int64, actual object.Object) error {
	result, ok := actual.(*object.Integer)
	if !ok {
		return fmt.Errorf(utils.Red("object is not Integer.")+" got=%T (%+v)",
			actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf(utils.Red("object has wrong value.")+" got=%d, want=%d",
			result.Value, expected)
	}

	return nil
}

func testBooleanObject(expected bool, actual object.Object) error {
	result, ok := actual.(*object.Boolean)
	if !ok {
		return fmt.Errorf(utils.Red("object is not Boolean.")+" got = %T(%+v)",
			actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf(utils.Red("object has wrong value.")+" got=%t, want=%t",
			result.Value, expected)
	}

	return nil
}

func testStringObject(expected string, actual object.Object) error {
	result, ok := actual.(*object.String)
	if !ok {
		return fmt.Errorf(utils.Red("object is not String.")+" got = %T(%+v)",
			actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf(utils.Red("object has wrong value.")+" got=%q, want=%q",
			result.Value, expected)
	}

	return nil
}

func testIntArrayObject(expected []int, actual object.Object) error {
	arr, ok := actual.(*object.Array)
	if !ok {
		return fmt.Errorf(utils.Red("object is not Array")+" got = %T(%+v)",
			actual, actual)
	}

	if len(arr.Elements) != len(expected) {
		return fmt.Errorf(utils.Red("array has wrong length.")+" got=%d, want=%d",
			len(arr.Elements), len(expected))
	}

	for i, expectedEl := range expected {
		err := testIntegerObject(int64(expectedEl), arr.Elements[i])
		if err != nil {
			return fmt.Errorf("testIntegerObject failed: %s", err)
		}
	}

	return nil
}

func testIntHashObject(expected map[object.HashKey]int64, actual object.Object) error {
	hash, ok := actual.(*object.Hash)
	if !ok {
		return fmt.Errorf(utils.Red("object is not Hash")+" got = %T(%+v)",
			actual, actual)
	}

	if len(hash.Pairs) != len(expected) {
		return fmt.Errorf(utils.Red("hash pair amount mismatch.")+" got=%d, want=%d",
			len(hash.Pairs), len(expected))
	}

	for expectedKey, expectedVal := range expected {
		pair, ok := hash.Pairs[expectedKey]
		if !ok {
			return fmt.Errorf("no pair in hash for key=%d", expectedKey.Value)
		}
		err := testIntegerObject(expectedVal, pair.Value)
		if err != nil {
			return fmt.Errorf("testIntegerObject failed: %s", err)
		}
	}

	return nil
}
