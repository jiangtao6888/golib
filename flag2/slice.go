package flag2

import (
	"fmt"
	"strconv"
)

type Strings []string

func (t *Strings) String() string {
	return fmt.Sprint(*t)
}

func (t *Strings) Set(value string) error {
	*t = append(*t, value)
	return nil
}

type Integers []int

func (t *Integers) String() string {
	return fmt.Sprint(*t)
}

func (t *Integers) Set(value string) error {
	n, err := strconv.Atoi(value)

	if err != nil {
		return err
	}

	*t = append(*t, n)
	return nil
}
