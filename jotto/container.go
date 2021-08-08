package motto

import (
	"context"
	"fmt"
	"reflect"
)

// Container - an IoC container
type Container interface {
	Register(kind interface{}, tag interface{}, factory DepFactory, options *DepOptions) error
	Make(ctx context.Context, value interface{}, tag interface{}) error
}

// NewContainer - creates an IoC container
func NewContainer(app Application) Container {
	return &container{
		app:      app,
		registry: make(map[interface{}]*typeRegistry),
	}
}

// DepFactory - a factory function that makes a specific type of dependency
type DepFactory func(ctx context.Context, app Application) (interface{}, error)

// DepOptions - options for a dependency item
type DepOptions struct {
	Singleton bool
}

type registryEntry struct {
	factory DepFactory
	options *DepOptions
}

type typeRegistry struct {
	entries map[interface{}]*registryEntry
	objects map[interface{}]interface{}
}

type container struct {
	app      Application
	registry map[interface{}]*typeRegistry
}

// Register - register a dependency into the IoC container
func (ioc *container) Register(template interface{}, tag interface{}, factory DepFactory, options *DepOptions) (err error) {
	var (
		kind     reflect.Type
		ok       bool
		registry *typeRegistry
	)
	kind = reflect.TypeOf(template)

	if registry, ok = ioc.registry[kind]; !ok {
		registry = &typeRegistry{
			entries: make(map[interface{}]*registryEntry),
			objects: make(map[interface{}]interface{}),
		}
		ioc.registry[kind] = registry
	}

	if _, ok = registry.entries[tag]; ok {
		return fmt.Errorf("Entry `%v`.`%v` already exists", kind, tag)
	}

	registry.entries[tag] = &registryEntry{
		factory: factory,
		options: options,
	}

	return nil
}

// Make - retrieve/make a dependency entry out of the IoC container
func (ioc *container) Make(ctx context.Context, value interface{}, tag interface{}) (err error) {
	var (
		kind     reflect.Type
		registry *typeRegistry
		entry    *registryEntry
		ok       bool
		object   interface{}
	)

	kind = reflect.TypeOf(value)

	fmt.Println("kind", kind)

	if registry, ok = ioc.registry[kind]; !ok {
		return fmt.Errorf("Type `%v` is not registered", kind)
	}

	fmt.Println("registry", registry)

	if entry, ok = registry.entries[tag]; !ok {
		return fmt.Errorf("Entry `%v`.`%v` is not registered", kind, tag)
	}

	fmt.Println("entry", entry)

	if !entry.options.Singleton {
		object, err = entry.factory(ctx, ioc.app)
	} else {
		if object, ok = registry.objects[tag]; !ok {
			object, err = entry.factory(ctx, ioc.app)
		}
		registry.objects[tag] = object
	}

	if err != nil {
		return err
	}

	v := reflect.Indirect(reflect.ValueOf(value))
	v.Set(reflect.ValueOf(object))

	return nil
}