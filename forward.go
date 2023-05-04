package bird

import (
	errs "errors"
	"fmt"
	"net/http"

	"github.com/dev-mockingbird/logf"
	"github.com/dev-mockingbird/validate"
	"go-micro.dev/v4/errors"
)

type Forwarder interface {
	Forward(c Actor, creq any, forward func() error, rules ...validate.Rules) error
}

type Forward func(c Actor, creq any, forward func() error, rules ...validate.Rules) error

func (f Forward) Forward(c Actor, creq any, forward func() error, rules ...validate.Rules) error {
	return f(c, creq, forward, rules...)
}

func GetForwarder() Forwarder {
	return Forward(func(c Actor, creq any, forward func() error, rules ...validate.Rules) error {
		if err := c.Bind(creq); err != nil {
			msg := fmt.Sprintf("can't parse request: %s", err.Error())
			c.Write(http.StatusBadRequest, InvalidArguments(err, msg))
			c.Logger().Logf(logf.Trace, msg)
			return err
		}
		if err := c.Validate(creq, rules...); err != nil {
			c.Write(http.StatusBadRequest, InvalidArguments(err, "argument can't be verified"))
			c.Logger().Logf(logf.Trace, "invalid request: %s", err.Error())
			return err
		}
		if err := forward(); err != nil {
			c.Logger().Logf(logf.Error, "forward: %s", err.Error())
			if e, ok := err.(*errors.Error); ok {
				err = errs.New(e.Detail)
				c.Write(int(e.Code), ErrorOccurred(err, e.Status, "internal server error, please try again"))
				return err
			}
			c.Write(http.StatusInternalServerError, UnknownError(err, "internal server error, please try again"))
			return err
		}
		return nil
	})
}
