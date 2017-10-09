/*
Copyright 2017 WALLIX

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package driver

import (
	"errors"
	"fmt"

	"github.com/wallix/awless/logger"
)

var ErrDriverFnNotFound = errors.New("driver function not found")

func DryRun(cmd interface{}, ctx Context, params map[string]interface{}) (interface{}, error) {
	type I interface {
		Inject(map[string]interface{}) error
	}
	if i, ok := cmd.(I); ok {
		if err := i.Inject(params); err != nil {
			return nil, err
		}
	}

	type V interface {
		Validate() error
	}
	if i, ok := cmd.(V); ok {
		if err := i.Validate(); err != nil {
			return nil, err
		}
	}

	var result interface{}
	type DR interface {
		DryRun() (interface{}, error)
	}
	if i, ok := cmd.(DR); ok {
		var err error
		if result, err = i.DryRun(); err != nil {
			return result, err
		}
	} else {
		return result, errors.New("Command is not a dry runner")
	}

	return result, nil
}

func Run(cmd interface{}, ctx Context, params map[string]interface{}) (interface{}, error) {
	type I interface {
		Inject(map[string]interface{}) error
	}
	if i, ok := cmd.(I); ok {
		if err := i.Inject(params); err != nil {
			return nil, err
		}
	}

	type V interface {
		Validate() error
	}
	if i, ok := cmd.(V); ok {
		if err := i.Validate(); err != nil {
			return nil, err
		}
	}

	var result interface{}
	type R interface {
		Run() (interface{}, error)
	}
	if i, ok := cmd.(R); ok {
		var err error
		if result, err = i.Run(); err != nil {
			return result, err
		}
	} else {
		return result, errors.New("Command is not a runner")
	}

	type AR interface {
		AfterRun() error
	}
	if i, ok := cmd.(AR); ok {
		if err := i.AfterRun(); err != nil {
			return result, err
		}
	}

	return result, nil
}

type LookupFunc func(...string) interface{}

type Driver interface {
	Lookup(...string) (DriverFn, error)
	LookupIface(...string) (interface{}, error)
	SetDryRun(bool)
	SetLogger(*logger.Logger)
}

type Context interface {
	Variables() map[string]interface{}
	References() map[string]interface{} // retro-compatibility with v0.1.2
}

func NewContext(vars map[string]interface{}) Context {
	return &context{vars: vars}
}

var EmptyContext = &context{}

type context struct {
	vars map[string]interface{}
}

func (c *context) Variables() map[string]interface{} {
	return copyMap(c.vars)
}

func (c *context) References() map[string]interface{} { // retro-compatibility with v0.1.2
	return copyMap(c.vars)
}

type DriverFn func(Context, map[string]interface{}) (interface{}, error)

type MultiDriver struct {
	drivers []Driver
}

func NewMultiDriver(drivers ...Driver) Driver {
	return &MultiDriver{drivers: drivers}
}

func (d *MultiDriver) SetDryRun(dry bool) {
	for _, dr := range d.drivers {
		dr.SetDryRun(dry)
	}
}

func (d *MultiDriver) SetLogger(l *logger.Logger) {
	for _, dr := range d.drivers {
		dr.SetLogger(l)
	}
}

func (d *MultiDriver) LookupIface(lookups ...string) (interface{}, error) {
	for _, dr := range d.drivers {
		iface, err := dr.LookupIface(lookups...)
		if err != nil {
			return nil, err
		}
		if iface != nil {
			return iface, nil
		}
	}
	return nil, nil
}

func (d *MultiDriver) Lookup(lookups ...string) (driverFn DriverFn, err error) {
	var funcs []DriverFn
	for _, dr := range d.drivers {
		fn, err := dr.Lookup(lookups...)
		if err == ErrDriverFnNotFound {
			continue
		}
		if err != nil {
			return nil, err
		}
		funcs = append(funcs, fn)
	}
	switch len(funcs) {
	case 0:
		return nil, fmt.Errorf("function corresponding to '%v' not found in drivers", lookups)
	case 1:
		return funcs[0], nil
	default:
		return nil, fmt.Errorf("%d functions corresponding to '%v' found in drivers", len(funcs), lookups)
	}
}

func copyMap(m map[string]interface{}) (copy map[string]interface{}) {
	copy = make(map[string]interface{})
	if m == nil {
		return
	}
	for k, v := range m {
		copy[k] = v
	}
	return
}
