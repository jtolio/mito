package uel

import (
	"fmt"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	testRun(t, nil)
}

func testRun(t testing.TB, inputcb func(string)) {
	checkResult := func(input string, env map[any]any, expected interface{}) {
		if inputcb != nil {
			inputcb(input)
		}
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

	checkResult(
		`(
	# Elevation (ft)
	elevation >= 100
	and
	# Elevation (ft)
	elevation <= 8000
	and
	# Average annual min temperature (deg F), 2050 value
	tmin_avg_min_2050 >= -6
	and
	# Average precipitation (in/year), 2050 value
	prec_avg_2050 >= 20
	and
	# Average annual days above 95 F, 2050 value
	tmax_days_above_95_2050 <= 20
	and
	# Average temperature (deg F), 2010-2050 change
	tmean_avg_2050d <= 3
	and
	# Average annual days at or below freezing, 2050 value
	tmin_days_at_or_below_32_2050 <= 130
	and
	# Average annual max daily average wet-bulb temperature (Stull method, deg F), 2050 value
	wetbulb_avg_max_2050 < 79
	and
	# Average annual days with no precipitation, 2050 value
	prec_days_at_or_below_0_2050 <= 220
	)`,
		map[any]any{
			"elevation":                     int64(101),
			"tmin_avg_min_2050":             int64(-5),
			"prec_avg_2050":                 int64(21),
			"tmax_days_above_95_2050":       int64(19),
			"tmean_avg_2050d":               int64(2),
			"tmin_days_at_or_below_32_2050": int64(129),
			"wetbulb_avg_max_2050":          int64(78),
			"prec_days_at_or_below_0_2050":  int64(220),
		},
		true)

}

func FuzzRun(f *testing.F) {
	testRun(f, func(input string) { f.Add(input) })
	f.Add("")
	f.Add("||0")
	f.Add("0(,000)")
	f.Add("0(0,)")
	f.Add("0s/0")
	f.Add(`""*-1`)
	f.Add("\"0\xb5\xcf00\x97\xa10\x1800000\xab\"*100000000")
	emptyEnv := map[any]any{}

	// just make sure there aren't any panics
	f.Fuzz(func(t *testing.T, input string) {
		val, err := Parse(input)
		if err == nil {
			if val == nil {
				panic(fmt.Sprintf("%q", input))
			}
			_, _ = val.Run(emptyEnv)
		}
	})
}
