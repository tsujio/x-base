package testutils

import (
	"regexp"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

type AnyVal struct{}

type UUID struct{}

func (u UUID) match(s string) bool {
	id, err := uuid.Parse(s)
	if err != nil {
		return false
	}
	if id == uuid.Nil {
		return false
	}
	return true
}

type Timestamp struct{}

func (t Timestamp) match(s string) bool {
	tm, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return false
	}
	if tm.IsZero() {
		return false
	}
	return true
}

type Regexp struct {
	Pattern string
}

func (r Regexp) match(s string) bool {
	return regexp.MustCompile(r.Pattern).MatchString(s)
}

func CompareJson(x, y interface{}) string {
	var opts []cmp.Option

	// Compare AnyVal
	opts = append(opts, cmp.FilterValues(
		func(x, y interface{}) bool {
			_, isXAnyVal := x.(AnyVal)
			_, isYAnyVal := y.(AnyVal)
			return isXAnyVal || isYAnyVal
		},
		cmp.Comparer(func(x, y interface{}) bool {
			return true
		}),
	))

	// Compare UUID
	opts = append(opts, cmp.FilterValues(
		func(x, y interface{}) bool {
			_, isXUUID := x.(UUID)
			_, isYUUID := y.(UUID)
			return isXUUID || isYUUID
		},
		cmp.Comparer(func(x, y interface{}) bool {
			switch xv := x.(type) {
			case UUID:
				switch yv := y.(type) {
				case UUID:
					return true
				case string:
					return xv.match(yv)
				default:
					return false
				}
			case string:
				switch yv := y.(type) {
				case UUID:
					return yv.match(xv)
				default:
					return cmp.Equal(x, y, opts...)
				}
			default:
				return cmp.Equal(x, y, opts...)
			}
		}),
	))

	// Compare Timestamp
	opts = append(opts, cmp.FilterValues(
		func(x, y interface{}) bool {
			_, isXTimestamp := x.(Timestamp)
			_, isYTimestamp := y.(Timestamp)
			return isXTimestamp || isYTimestamp
		},
		cmp.Comparer(func(x, y interface{}) bool {
			switch xv := x.(type) {
			case Timestamp:
				switch yv := y.(type) {
				case Timestamp:
					return true
				case string:
					return xv.match(yv)
				default:
					return false
				}
			case string:
				switch yv := y.(type) {
				case Timestamp:
					return yv.match(xv)
				default:
					return cmp.Equal(x, y, opts...)
				}
			default:
				return cmp.Equal(x, y, opts...)
			}
		}),
	))

	// Compare Regexp
	opts = append(opts, cmp.FilterValues(
		func(x, y interface{}) bool {
			_, isXRegexp := x.(Regexp)
			_, isYRegexp := y.(Regexp)
			return isXRegexp || isYRegexp
		},
		cmp.Comparer(func(x, y interface{}) bool {
			switch xv := x.(type) {
			case Regexp:
				switch yv := y.(type) {
				case Regexp:
					return xv.Pattern == yv.Pattern
				case string:
					return xv.match(yv)
				default:
					return false
				}
			case string:
				switch yv := y.(type) {
				case Regexp:
					return yv.match(xv)
				default:
					return cmp.Equal(x, y, opts...)
				}
			default:
				return cmp.Equal(x, y, opts...)
			}
		}),
	))

	return cmp.Diff(x, y, opts...)
}
