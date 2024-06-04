package component

import (
	"errors"
	"reflect"
)

/********************************************************************
created:    2020-08-29
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type (
	// Service implements a specific service, some of it's methods will be
	// called when the correspond events is occurred.
	Service struct {
		Name     string              // name of service
		Type     reflect.Type        // type of the receiver
		Receiver reflect.Value       // receiver of methods for the service
		Handlers map[string]*Handler // registered methods
		Options  options             // options
	}
)

// NewService creates a new service
func NewService(comp Component, opts []Option) *Service {
	s := &Service{
		Type:     reflect.TypeOf(comp),
		Receiver: reflect.ValueOf(comp),
	}

	// apply options
	for i := range opts {
		opt := opts[i]
		opt(&s.Options)
	}
	if name := s.Options.name; name != "" {
		s.Name = name
	} else {
		s.Name = reflect.Indirect(s.Receiver).Type().Name()
	}

	return s
}

// ExtractHandler extract the set of methods from the
// receiver value which satisfy the following conditions:
// - exported method of exported type
// - one or two arguments
// - the first argument is context.Context
// - the second argument (if it exists) is []byte or a pointer
// - zero or two outputs
// - the first output is [] or a pointer
// - the second output is an error
func (service *Service) ExtractHandler() error {
	typeName := reflect.Indirect(service.Receiver).Type().Name()
	if typeName == "" {
		return errors.New("no service name for type " + service.Type.String())
	}

	if !isExported(typeName) {
		return errors.New("type " + typeName + " is not exported")
	}

	// Install the methods
	service.Handlers = suitableHandlerMethods(service.Type, service.Options.nameFunc)

	if len(service.Handlers) == 0 {
		var str = ""
		// To help the user, see if a pointer receiver would work.
		var methods = suitableHandlerMethods(reflect.PointerTo(service.Type), service.Options.nameFunc)
		if len(methods) != 0 {
			str = "type " + service.Name + " has no exported methods of handler type (hint: pass a pointer to value of that type)"
		} else {
			str = "type " + service.Name + " has no exported methods of handler type"
		}

		return errors.New(str)
	}

	for methodName := range service.Handlers {
		service.Handlers[methodName].Receiver = service.Receiver
	}

	return nil
}
