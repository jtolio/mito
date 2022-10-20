package uel

import (
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	checkResult := func(input string, env map[any]any, expected interface{}) {
		t.Helper()
		val, err := Eval(input, env)
		if err != nil {
			t.Fatal(err)
		}
		if val != expected {
			t.Fatalf("input %#v with env %#v expected %#v, got %#v", input, env, expected, val)
		}
	}

	emptyEnv := map[any]any{}

	checkResult("false || false", emptyEnv, false)
	checkResult("true || false", emptyEnv, true)
	checkResult("false || true", emptyEnv, true)
	checkResult("false || (true && false)", emptyEnv, false)
	checkResult("false || (true && true)", emptyEnv, true)

	checkResult("1 + 2", emptyEnv, int64(3))
	checkResult("1+2", emptyEnv, int64(3))
	checkResult("1 - 2", emptyEnv, int64(-1))
	checkResult("1+2 * 3 / 4 * 5", emptyEnv, 1+((2*3)/float64(4))*5)
	checkResult("(1+2)*3/4*5", emptyEnv, (1+2)*3*5/float64(4))
	checkResult(
		`
    1 # a one
    + 2 # add a two
  `,
		emptyEnv,
		int64(3),
	)
	checkResult("1 < 2", emptyEnv, true)
	checkResult("1 > 2", emptyEnv, false)
	checkResult("1 <= 2", emptyEnv, true)
	checkResult("1 >= 2", emptyEnv, false)
	checkResult("2 <= 2", emptyEnv, true)
	checkResult("2 >= 2", emptyEnv, true)
	checkResult("2 == 2", emptyEnv, true)
	checkResult("not (2 != 2)", emptyEnv, true)
	checkResult("2 != 2", emptyEnv, false)
	checkResult("2 != 1", emptyEnv, true)
	checkResult("2 == 1", emptyEnv, false)
	checkResult("1 + (10 / 2) ", emptyEnv, float64(6))
	checkResult("1 + (10 / 2) > 3", emptyEnv, true)

	var printed int64
	checkResult(`print(8) + 3`, map[any]any{
		"print": func(a int64) (int64, error) {
			printed = a
			return a, nil
		},
	}, int64(11))
	if printed != 8 {
		t.Fatal("printed didn't work")
	}

	checkResult(`print(2) + 3`, map[any]any{
		"print": func(a int64) int64 {
			printed = a
			return a
		},
	}, int64(5))
	if printed != 2 {
		t.Fatal("printed didn't work")
	}

	var printedStr string
	checkResult(`print("hello") + "world"`, map[any]any{
		"print": func(a string) (string, error) {
			printedStr = a
			return a + ", ", nil
		},
	}, "hello, world")
	if printedStr != "hello" {
		t.Fatal("printed didn't work")
	}

	checkResult("2h", emptyEnv, 2*time.Hour)
	checkResult("2s == 2 * (1s)", emptyEnv, true)

}

