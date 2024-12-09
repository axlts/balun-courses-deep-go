package main

import (
	"os"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

// go test -v homework_test.go

type UserService struct {
	// not need to implement
	NotEmptyStruct bool
}
type MessageService struct {
	// not need to implement
	NotEmptyStruct bool
}

type Container struct {
	types map[string]interface{}
}

func NewContainer() *Container {
	return &Container{
		types: make(map[string]interface{}),
	}
}

func (c *Container) RegisterType(name string, ctor interface{}) {
	c.types[name] = ctor
}

func (c *Container) RegisterSingletonType(name string, ctor interface{}) {
	obj := ctor.(func() interface{})()
	c.types[name] = func() interface{} {
		return obj
	}
}

func (c *Container) Resolve(name string) (interface{}, error) {
	ctor, ok := c.types[name]
	if !ok {
		return nil, os.ErrNotExist
	}
	return ctor.(func() interface{})(), nil
}

func TestDIContainer(t *testing.T) {
	container := NewContainer()
	container.RegisterType("UserService", func() interface{} {
		return &UserService{}
	})
	container.RegisterType("MessageService", func() interface{} {
		return &MessageService{}
	})

	userService1, err := container.Resolve("UserService")
	assert.NoError(t, err)
	userService2, err := container.Resolve("UserService")
	assert.NoError(t, err)

	u1 := userService1.(*UserService)
	u2 := userService2.(*UserService)
	assert.False(t, u1 == u2)
	assert.False(t, unsafe.Pointer(u1) == unsafe.Pointer(u2))

	messageService, err := container.Resolve("MessageService")
	assert.NoError(t, err)
	assert.NotNil(t, messageService)

	paymentService, err := container.Resolve("PaymentService")
	assert.Error(t, err)
	assert.Nil(t, paymentService)
}

func TestDIContainer_RegisterSingletonType(t *testing.T) {
	container := NewContainer()
	container.RegisterSingletonType("UserService", func() interface{} {
		return &UserService{}
	})

	userService1, err := container.Resolve("UserService")
	assert.NoError(t, err)

	u1 := userService1.(*UserService)
	u1.NotEmptyStruct = true

	userService2, err := container.Resolve("UserService")
	assert.NoError(t, err)

	u2 := userService2.(*UserService)
	assert.True(t, u2.NotEmptyStruct)

	assert.True(t, unsafe.Pointer(u1) == unsafe.Pointer(u2))
}
