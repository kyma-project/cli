package di

import (
	"fmt"
	"reflect"
	"sync"
)

// DIContainer is a simple dependency injection container for singleton instances
type DIContainer struct {
	mu        sync.RWMutex
	instances map[reflect.Type]interface{}
	factories map[reflect.Type]Factory
}

type Factory func(container *DIContainer) (interface{}, error)

func NewDIContainer() *DIContainer {
	return &DIContainer{
		instances: make(map[reflect.Type]interface{}),
		factories: make(map[reflect.Type]Factory),
	}
}

// Get retrieves or creates a singleton instance of the specified type
func (c *DIContainer) Get(targetType reflect.Type) (interface{}, error) {
	// Fast path: check if instance already exists
	if instance, exists := c.instances[targetType]; exists {
		return instance, nil
	}

	factory, exists := c.factories[targetType]

	if !exists {
		return nil, fmt.Errorf("no registration found for type %s", targetType.String())
	}

	// Create instance (without holding lock to allow nested dependency resolution)
	instance, err := factory(c)
	if err != nil {
		return nil, fmt.Errorf("failed to create instance of type %s: %w", targetType.String(), err)
	}

	// Check if another goroutine created it while we were creating ours
	if existing, exists := c.instances[targetType]; exists {
		return existing, nil
	}
	c.instances[targetType] = instance

	return instance, nil
}

// GetTyped is a generic helper to get a singleton instance with type safety
func GetTyped[T any](c *DIContainer) (T, error) {
	var zero T
	targetType := reflect.TypeOf((*T)(nil)).Elem()

	instance, err := c.Get(targetType)
	if err != nil {
		return zero, err
	}

	typed, ok := instance.(T)
	if !ok {
		return zero, fmt.Errorf("instance is not of expected type %T", zero)
	}

	return typed, nil
}

// RegisterTyped is a generic helper to register a factory with type safety
// Usage: RegisterTyped[MyInterface](container, func(c *DIContainer) (MyInterface, error) { ... })
func RegisterTyped[T any](c *DIContainer, factory func(*DIContainer) (T, error)) {
	targetType := reflect.TypeOf((*T)(nil)).Elem()
	c.factories[targetType] = func(container *DIContainer) (interface{}, error) {
		return factory(container)
	}
}
