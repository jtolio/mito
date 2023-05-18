package uel

import (
	"encoding/hex"
	"fmt"
	"math"
	"strings"
	"time"
)

var defaultEnv = map[any]any{}

func init() {
	defaultEnv = map[any]any{
		OpOr: func(env map[any]any, a, b any) (any, error) {
			x, aok := a.(bool)
			y, bok := b.(bool)
			if !aok || !bok {
				return nil, fmt.Errorf("%w: bool expected: %#v, %#v", ErrTypeMismatch, a, b)
			}
			return x || y, nil
		},

		OpAnd: func(env map[any]any, a, b any) (any, error) {
			x, aok := a.(bool)
			y, bok := b.(bool)
			if !aok || !bok {
				return nil, fmt.Errorf("%w: bool expected: %#v, %#v", ErrTypeMismatch, a, b)
			}
			return x && y, nil
		},

		OpAdd: func(env map[any]any, a, b any) (any, error) {
			switch x := a.(type) {
			case string:
				switch y := b.(type) {
				case string:
					return x + y, nil
				case int64:
					return x + fmt.Sprint(y), nil
				case float64:
					return x + fmt.Sprint(y), nil
				case []byte:
					return x + "0x" + hex.EncodeToString(y), nil
				case bool:
					return x + fmt.Sprint(y), nil
				case time.Duration:
					return x + y.String(), nil
				case time.Time:
					return x + y.String(), nil
				default:
					return nil, fmt.Errorf("%w: unsupported type for addition %T + %T", ErrTypeMismatch, a, b)
				}
			case int64:
				switch y := b.(type) {
				case string:
					return fmt.Sprint(x) + y, nil
				case int64:
					return x + y, nil
				case float64:
					return float64(x) + y, nil
				case []byte:
					return fmt.Sprint(x) + "0x" + hex.EncodeToString(y), nil
				case bool:
					if y {
						return x + 1, nil
					}
					return x, nil
				case time.Duration:
					return (time.Duration(x) * time.Second) + y, nil
				case time.Time:
					return y.Add(time.Duration(x) * time.Second), nil
				default:
					return nil, fmt.Errorf("%w: unsupported type for addition %T + %T", ErrTypeMismatch, a, b)
				}
			case float64:
				switch y := b.(type) {
				case string:
					return fmt.Sprint(x) + y, nil
				case int64:
					return x + float64(y), nil
				case float64:
					return x + y, nil
				case []byte:
					return fmt.Sprint(x) + "0x" + hex.EncodeToString(y), nil
				case bool:
					if y {
						return x + 1, nil
					}
					return x, nil
				case time.Duration:
					return time.Duration(x*float64(time.Second)) + y, nil
				case time.Time:
					return y.Add(time.Duration(x * float64(time.Second))), nil
				default:
					return nil, fmt.Errorf("%w: unsupported type for addition %T + %T", ErrTypeMismatch, a, b)
				}
			case []byte:
				switch y := b.(type) {
				case string:
					return "0x" + hex.EncodeToString(x) + y, nil
				case int64:
					return "0x" + hex.EncodeToString(x) + fmt.Sprint(y), nil
				case float64:
					return "0x" + hex.EncodeToString(x) + fmt.Sprint(y), nil
				case []byte:
					return append(append([]byte(nil), x...), y...), nil
				case bool:
					return "0x" + hex.EncodeToString(x) + fmt.Sprint(y), nil
				case time.Duration:
					return "0x" + hex.EncodeToString(x) + y.String(), nil
				case time.Time:
					return "0x" + hex.EncodeToString(x) + y.String(), nil
				default:
					return nil, fmt.Errorf("%w: unsupported type for addition %T + %T", ErrTypeMismatch, a, b)
				}
			case bool:
				switch y := b.(type) {
				case string:
					return fmt.Sprint(x) + y, nil
				case int64:
					if x {
						return y + 1, nil
					}
					return y, nil
				case float64:
					if x {
						return y + 1, nil
					}
					return y, nil
				case []byte:
					return fmt.Sprint(x) + "0x" + hex.EncodeToString(y), nil
				case bool:
					if x && y {
						return int64(2), nil
					}
					if x || y {
						return int64(1), nil
					}
					return int64(0), nil
				case time.Duration:
					if x {
						return y + time.Second, nil
					}
					return y, nil
				case time.Time:
					if x {
						return y.Add(time.Second), nil
					}
					return y, nil
				default:
					return nil, fmt.Errorf("%w: unsupported type for addition %T + %T", ErrTypeMismatch, a, b)
				}
			case time.Duration:
				switch y := b.(type) {
				case string:
					return x.String() + y, nil
				case int64:
					return x + (time.Duration(y) * time.Second), nil
				case float64:
					return x + (time.Duration(y * float64(time.Second))), nil
				case []byte:
					return x.String() + "0x" + hex.EncodeToString(y), nil
				case bool:
					if y {
						return x + time.Second, nil
					}
					return x, nil
				case time.Duration:
					return x + y, nil
				case time.Time:
					return y.Add(x), nil
				default:
					return nil, fmt.Errorf("%w: unsupported type for addition %T + %T", ErrTypeMismatch, a, b)
				}
			case time.Time:
				switch y := b.(type) {
				case string:
					return x.String() + y, nil
				case int64:
					return x.Add(time.Duration(y) * time.Second), nil
				case float64:
					return x.Add(time.Duration(y * float64(time.Second))), nil
				case []byte:
					return x.String() + "0x" + hex.EncodeToString(y), nil
				case bool:
					if y {
						return x.Add(time.Second), nil
					}
					return x, nil
				case time.Duration:
					return x.Add(y), nil
				case time.Time:
					return nil, fmt.Errorf("%w: unsupported type for addition %T + %T", ErrTypeMismatch, a, b)
				default:
					return nil, fmt.Errorf("%w: unsupported type for addition %T + %T", ErrTypeMismatch, a, b)
				}
			default:
				return nil, fmt.Errorf("%w: unsupported type for addition %T", ErrTypeMismatch, a)
			}
		},

		OpSub: func(env map[any]any, a, b any) (any, error) {
			switch x := a.(type) {
			case string, []byte:
				return nil, fmt.Errorf("%w: unsupported type for subtraction %T - %T", ErrTypeMismatch, a, b)
			case int64:
				switch y := b.(type) {
				case int64:
					return x - y, nil
				case float64:
					return float64(x) - y, nil
				case bool:
					if y {
						return x - 1, nil
					}
					return x, nil
				case time.Duration:
					return (time.Duration(x) * time.Second) - y, nil
				default:
					return nil, fmt.Errorf("%w: unsupported type for subtraction %T - %T", ErrTypeMismatch, a, b)
				}
			case float64:
				switch y := b.(type) {
				case int64:
					return x - float64(y), nil
				case float64:
					return x - y, nil
				case bool:
					if y {
						return x - 1, nil
					}
					return x, nil
				case time.Duration:
					return time.Duration(x*float64(time.Second)) - y, nil
				default:
					return nil, fmt.Errorf("%w: unsupported type for subtraction %T - %T", ErrTypeMismatch, a, b)
				}
			case bool:
				switch y := b.(type) {
				case int64:
					if x {
						return 1 - y, nil
					}
					return -y, nil
				case float64:
					if x {
						return 1 - y, nil
					}
					return -y, nil
				case bool:
					if x && y {
						return int64(0), nil
					}
					if x {
						return int64(1), nil
					}
					if y {
						return int64(-1), nil
					}
					return int64(0), nil
				case time.Duration:
					if x {
						return time.Second - y, nil
					}
					return -y, nil
				default:
					return nil, fmt.Errorf("%w: unsupported type for subtraction %T - %T", ErrTypeMismatch, a, b)
				}
			case time.Duration:
				switch y := b.(type) {
				case int64:
					return x - (time.Duration(y) * time.Second), nil
				case float64:
					return x - (time.Duration(y * float64(time.Second))), nil
				case bool:
					if y {
						return x - time.Second, nil
					}
					return x, nil
				case time.Duration:
					return x - y, nil
				default:
					return nil, fmt.Errorf("%w: unsupported type for subtraction %T - %T", ErrTypeMismatch, a, b)
				}
			case time.Time:
				switch y := b.(type) {
				case int64:
					return x.Add(-time.Duration(y) * time.Second), nil
				case float64:
					return x.Add(-time.Duration(y * float64(time.Second))), nil
				case bool:
					if y {
						return x.Add(-time.Second), nil
					}
					return x, nil
				case time.Duration:
					return x.Add(-y), nil
				case time.Time:
					return x.Sub(y), nil
				default:
					return nil, fmt.Errorf("%w: unsupported type for subtraction %T - %T", ErrTypeMismatch, a, b)
				}
			default:
				return nil, fmt.Errorf("%w: unsupported type for subtraction %T", ErrTypeMismatch, a)
			}
		},

		OpMul: func(env map[any]any, a, b any) (any, error) {
			switch x := a.(type) {
			case string:
				switch y := b.(type) {
				case int64:
					if y < 0 {
						return "", fmt.Errorf("%w: negative string repeat", ErrValueMismatch)
					}
					return strings.Repeat(x, int(y)), nil
				case float64:
					if y < 0 {
						return "", fmt.Errorf("%w: negative string repeat", ErrValueMismatch)
					}
					if math.IsInf(y, 0) || math.IsNaN(y) {
						return "", fmt.Errorf("%w: invalid string repeat", ErrValueMismatch)
					}
					return strings.Repeat(x, int(y)), nil
				case bool:
					if y {
						return x, nil
					}
					return "", nil
				default:
					return nil, fmt.Errorf("%w: unsupported type for multiplication %T * %T", ErrTypeMismatch, a, b)
				}
			case int64:
				switch y := b.(type) {
				case int64:
					return x * y, nil
				case float64:
					return float64(x) * y, nil
				case bool:
					if y {
						return x, nil
					}
					return int64(0), nil
				case time.Duration:
					return time.Duration(x) * y, nil
				default:
					return nil, fmt.Errorf("%w: unsupported type for multiplication %T * %T", ErrTypeMismatch, a, b)
				}
			case float64:
				switch y := b.(type) {
				case int64:
					return x * float64(y), nil
				case float64:
					return x * y, nil
				case bool:
					if y {
						return x, nil
					}
					return float64(0), nil
				case time.Duration:
					return time.Duration(int64(x)) * y, nil
				default:
					return nil, fmt.Errorf("%w: unsupported type for multiplication %T * %T", ErrTypeMismatch, a, b)
				}
			case []byte:
				switch y := b.(type) {
				case int64:
					rv := []byte(nil)
					for i := int64(0); i < y; i++ {
						rv = append(rv, x...)
					}
					return rv, nil
				case float64:
					rv := []byte(nil)
					for i := int64(0); i < int64(y); i++ {
						rv = append(rv, x...)
					}
					return rv, nil
				case bool:
					if y {
						return x, nil
					}
					return []byte(nil), nil
				default:
					return nil, fmt.Errorf("%w: unsupported type for multiplication %T * %T", ErrTypeMismatch, a, b)
				}
			case bool:
				switch y := b.(type) {
				case string:
					if x {
						return y, nil
					}
					return "", nil
				case int64:
					if x {
						return y, nil
					}
					return int64(0), nil
				case float64:
					if x {
						return y, nil
					}
					return float64(0), nil
				case []byte:
					if x {
						return y, nil
					}
					return []byte(nil), nil
				case bool:
					return x && y, nil
				case time.Duration:
					if x {
						return y, nil
					}
					return time.Duration(0), nil
				default:
					return nil, fmt.Errorf("%w: unsupported type for multiplication %T * %T", ErrTypeMismatch, a, b)
				}
			case time.Duration:
				switch y := b.(type) {
				case int64:
					return x * (time.Duration(y)), nil
				case float64:
					return x * (time.Duration(int64(y))), nil
				case bool:
					if y {
						return x, nil
					}
					return time.Duration(0), nil
				case time.Duration:
					return x * y, nil
				default:
					return nil, fmt.Errorf("%w: unsupported type for multiplication %T * %T", ErrTypeMismatch, a, b)
				}
			default:
				return nil, fmt.Errorf("%w: unsupported type for multiplication %T", ErrTypeMismatch, a)
			}
		},

		OpDiv: func(env map[any]any, a, b any) (any, error) {
			switch x := a.(type) {
			case int64:
				switch y := b.(type) {
				case int64:
					if y == 0 {
						return 0, fmt.Errorf("%w: division by zero", ErrValueMismatch)
					}
					return float64(x) / float64(y), nil
				case float64:
					if y == 0 {
						return 0, fmt.Errorf("%w: division by zero", ErrValueMismatch)
					}
					return float64(x) / y, nil
				default:
					return nil, fmt.Errorf("%w: unsupported type for division %T / %T", ErrTypeMismatch, a, b)
				}
			case float64:
				switch y := b.(type) {
				case int64:
					if y == 0 {
						return 0, fmt.Errorf("%w: division by zero", ErrValueMismatch)
					}
					return x / float64(y), nil
				case float64:
					if y == 0 {
						return 0, fmt.Errorf("%w: division by zero", ErrValueMismatch)
					}
					return x / y, nil
				default:
					return nil, fmt.Errorf("%w: unsupported type for division %T / %T", ErrTypeMismatch, a, b)
				}
			case time.Duration:
				switch y := b.(type) {
				case int64:
					if y == 0 {
						return 0, fmt.Errorf("%w: division by zero", ErrValueMismatch)
					}
					return x / time.Duration(y), nil
				case float64:
					if y == 0 {
						return 0, fmt.Errorf("%w: division by zero", ErrValueMismatch)
					}
					return time.Duration(float64(x) / y), nil
				case time.Duration:
					if y == 0 {
						return 0, fmt.Errorf("%w: division by zero", ErrValueMismatch)
					}
					return float64(x) / float64(y), nil
				default:
					return nil, fmt.Errorf("%w: unsupported type for division %T / %T", ErrTypeMismatch, a, b)
				}
			default:
				return nil, fmt.Errorf("%w: unsupported type for division %T", ErrTypeMismatch, a)
			}
		},

		// type order:
		//  bool/int64/float64/time.Duration, time.Time, string/[]byte

		OpLess: func(env map[any]any, a, b any) (any, error) {
			switch x := a.(type) {
			case string:
				switch y := b.(type) {
				case string:
					return x < y, nil
				case int64, bool, float64, time.Duration, time.Time:
					return false, nil
				case []byte:
					return x < string(y), nil
				default:
					return nil, fmt.Errorf("%w: unsupported type for comparison %T < %T", ErrTypeMismatch, a, b)
				}
			case int64:
				switch y := b.(type) {
				case string, time.Time, []byte:
					return true, nil
				case int64:
					return x < y, nil
				case float64:
					return float64(x) < y, nil
				case bool:
					if y {
						return x < 1, nil
					}
					return x < 0, nil
				case time.Duration:
					return (time.Duration(x) * time.Second) < y, nil
				default:
					return nil, fmt.Errorf("%w: unsupported type for comparison %T < %T", ErrTypeMismatch, a, b)
				}
			case float64:
				switch y := b.(type) {
				case string, time.Time, []byte:
					return true, nil
				case int64:
					return x < float64(y), nil
				case float64:
					return x < y, nil
				case bool:
					if y {
						return x < 1, nil
					}
					return x < 0, nil
				case time.Duration:
					return time.Duration(x*float64(time.Second)) < y, nil
				default:
					return nil, fmt.Errorf("%w: unsupported type for comparison %T < %T", ErrTypeMismatch, a, b)
				}
			case []byte:
				switch y := b.(type) {
				case string:
					return string(x) < y, nil
				case int64, float64, time.Duration, bool, time.Time:
					return false, nil
				case []byte:
					return string(x) < string(y), nil
				default:
					return nil, fmt.Errorf("%w: unsupported type for comparison %T < %T", ErrTypeMismatch, a, b)
				}
			case bool:
				switch y := b.(type) {
				case string, time.Time, []byte:
					return true, nil
				case int64:
					if x {
						return 1 < y, nil
					}
					return 0 < y, nil
				case float64:
					if x {
						return 1 < y, nil
					}
					return 0 < y, nil
				case bool:
					return !x && y, nil
				case time.Duration:
					if x {
						return time.Second < y, nil
					}
					return 0 < y, nil
				default:
					return nil, fmt.Errorf("%w: unsupported type for comparison %T < %T", ErrTypeMismatch, a, b)
				}
			case time.Duration:
				switch y := b.(type) {
				case string, []byte, time.Time:
					return true, nil
				case int64:
					return x < (time.Duration(y) * time.Second), nil
				case float64:
					return x < (time.Duration(y * float64(time.Second))), nil
				case bool:
					if y {
						return x < time.Second, nil
					}
					return x < 0, nil
				case time.Duration:
					return x < y, nil
				default:
					return nil, fmt.Errorf("%w: unsupported type for comparison %T < %T", ErrTypeMismatch, a, b)
				}
			case time.Time:
				switch y := b.(type) {
				case string, []byte:
					return false, nil
				case int64, bool, float64, time.Duration:
					return true, nil
				case time.Time:
					return x.Before(y), nil
				default:
					return nil, fmt.Errorf("%w: unsupported type for comparison %T < %T", ErrTypeMismatch, a, b)
				}
			default:
				return nil, fmt.Errorf("%w: unsupported type for comparison %T", ErrTypeMismatch, a)
			}
		},

		OpLessEqual: func(env map[any]any, a, b any) (any, error) {
			less, err := lessHelper(env, a, b)
			if err != nil {
				return nil, err
			}
			if less {
				return true, nil
			}
			greater, err := lessHelper(env, b, a)
			return !greater, err
		},

		OpEqual: func(env map[any]any, a, b any) (any, error) {
			less, err := lessHelper(env, a, b)
			if err != nil {
				return nil, err
			}
			if less {
				return false, nil
			}
			greater, err := lessHelper(env, b, a)
			return !greater, err
		},

		OpNotEqual: func(env map[any]any, a, b any) (any, error) {
			less, err := lessHelper(env, a, b)
			if err != nil {
				return nil, err
			}
			if less {
				return true, nil
			}
			greater, err := lessHelper(env, b, a)
			return greater, err
		},

		OpGreater: func(env map[any]any, a, b any) (any, error) {
			greater, err := lessHelper(env, b, a)
			return greater, err
		},

		OpGreaterEqual: func(env map[any]any, a, b any) (any, error) {
			greater, err := lessHelper(env, b, a)
			if err != nil {
				return nil, err
			}
			if greater {
				return true, nil
			}
			less, err := lessHelper(env, a, b)
			return !less, err
		},

		OpExp: func(env map[any]any, a, b any) (any, error) {
			switch x := a.(type) {
			case int64:
				switch y := b.(type) {
				case int64:
					return math.Pow(float64(x), float64(y)), nil
				case float64:
					return math.Pow(float64(x), y), nil
				case bool:
					if y {
						return float64(x), nil
					}
					return float64(1), nil
				default:
					return nil, fmt.Errorf("%w: unsupported type for exponentiation %T ^ %T", ErrTypeMismatch, a, b)
				}
			case float64:
				switch y := b.(type) {
				case int64:
					return math.Pow(x, float64(y)), nil
				case float64:
					return math.Pow(x, y), nil
				case bool:
					if y {
						return x, nil
					}
					return float64(1), nil
				default:
					return nil, fmt.Errorf("%w: unsupported type for exponentiation %T ^ %T", ErrTypeMismatch, a, b)
				}
			case bool:
				xb := float64(0)
				if x {
					xb = 1
				}
				switch y := b.(type) {
				case int64:
					return math.Pow(xb, float64(y)), nil
				case float64:
					return math.Pow(xb, y), nil
				case bool:
					yb := float64(0)
					if y {
						yb = 1
					}
					return math.Pow(xb, yb), nil
				default:
					return nil, fmt.Errorf("%w: unsupported type for exponentiation %T ^ %T", ErrTypeMismatch, a, b)
				}
			default:
				return nil, fmt.Errorf("%w: unsupported type for exponentiation %T", ErrTypeMismatch, a)
			}

		},

		ModNot: func(env map[any]any, a any) (any, error) {
			x, aok := a.(bool)
			if !aok {
				return nil, fmt.Errorf("%w: bool expected: %#v", ErrTypeMismatch, a)
			}
			return !x, nil
		},

		ModNeg: func(env map[any]any, a any) (any, error) {
			switch x := a.(type) {
			case int64:
				return -x, nil
			case float64:
				return -x, nil
			case bool:
				return !x, nil
			case time.Duration:
				return -x, nil
			default:
				return nil, fmt.Errorf("%w: unsupported type for negation %T", ErrTypeMismatch, a)
			}

		},

		"true":  true,
		"false": false,
	}
}

func lessHelper(env map[any]any, a, b any) (bool, error) {
	lessUncasted, ok := env[OpLess]
	if !ok {
		lessUncasted, ok = defaultEnv[OpLess]
	}
	if !ok {
		return false, fmt.Errorf("environment doesn't define less")
	}
	less, ok := lessUncasted.(func(env map[any]any, a, b any) (any, error))
	if !ok {
		return false, fmt.Errorf("environment defines less wrong")
	}
	res, err := less(env, a, b)
	if err != nil {
		return false, err
	}
	if res, ok := res.(bool); ok {
		return res, nil
	}
	return false, fmt.Errorf("less doesn't return bool")
}
