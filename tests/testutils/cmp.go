package testutils

import (
	"regexp"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

type AnyVal struct{}

type UUID struct{}

func (u UUID) match(v interface{}) bool {
	switch val := v.(type) {
	case UUID:
		return true
	case string:
		id, err := uuid.Parse(val)
		if err != nil {
			return false
		}
		if id == uuid.Nil {
			return false
		}
		return true
	default:
		return false
	}
}

type Timestamp struct{}

func (t Timestamp) match(v interface{}) bool {
	switch val := v.(type) {
	case Timestamp:
		return true
	case string:
		tm, err := time.Parse(time.RFC3339, val)
		if err != nil {
			return false
		}
		if tm.IsZero() {
			return false
		}
		return true
	default:
		return false
	}
}

type Regexp struct {
	Pattern string
}

func (r Regexp) match(v interface{}) bool {
	switch val := v.(type) {
	case Regexp:
		return r.Pattern == val.Pattern
	case string:
		return regexp.MustCompile(r.Pattern).MatchString(val)
	default:
		return false
	}
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
			if xv, ok := x.(UUID); ok {
				return xv.match(y)
			}
			if yv, ok := y.(UUID); ok {
				return yv.match(x)
			}
			return cmp.Equal(x, y, opts...)
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
			if xv, ok := x.(Timestamp); ok {
				return xv.match(y)
			}
			if yv, ok := y.(Timestamp); ok {
				return yv.match(x)
			}
			return cmp.Equal(x, y, opts...)
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
			if xv, ok := x.(Regexp); ok {
				return xv.match(y)
			}
			if yv, ok := y.(Regexp); ok {
				return yv.match(x)
			}
			return cmp.Equal(x, y, opts...)
		}),
	))

	// Compare uuid.UUID
	opts = append(opts, cmp.FilterValues(
		func(x, y interface{}) bool {
			_, isXUUID := x.(uuid.UUID)
			_, isYUUID := y.(uuid.UUID)
			return isXUUID || isYUUID
		},
		cmp.Comparer(func(x, y interface{}) bool {
			match := func(xv uuid.UUID, y interface{}) bool {
				switch yv := y.(type) {
				case uuid.UUID:
					return xv == yv
				case string:
					return xv.String() == yv
				default:
					return false
				}
			}
			if xv, ok := x.(uuid.UUID); ok {
				return match(xv, y)
			}
			if yv, ok := y.(uuid.UUID); ok {
				return match(yv, x)
			}
			return cmp.Equal(x, y, opts...)
		}),
	))

	return cmp.Diff(x, y, opts...)
}
