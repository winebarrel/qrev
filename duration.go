package qrev

import (
	"fmt"
	"time"
)

// UnsignedDuration is a time.Duration that rejects a negative value at parse
// time.
type UnsignedDuration time.Duration

func (d *UnsignedDuration) UnmarshalText(text []byte) error {
	v, err := time.ParseDuration(string(text))

	if err != nil {
		return err
	}

	if v < 0 {
		return fmt.Errorf("must not be negative: %s", v)
	}

	*d = UnsignedDuration(v)

	return nil
}
