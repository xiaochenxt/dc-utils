package try

import (
	"errors"
	"fmt"
)

func Run(exec func()) error {
	var err error
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case error:
				err = x
			case string:
				err = errors.New(x)
			default:
				err = fmt.Errorf("%v", x)
			}
		}
	}()
	exec()
	return err
}

func RunWithFinally(exec func(), finally func()) error {
	var err error
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case error:
				err = x
			case string:
				err = errors.New(x)
			default:
				err = fmt.Errorf("%v", x)
			}
		}
		if finally != nil {
			finallyErr := Run(finally)
			if err == nil {
				err = finallyErr
			}
		}
	}()
	exec()
	return err
}
