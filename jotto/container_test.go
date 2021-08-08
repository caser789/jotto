package motto_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"git.garena.com/duanzy/motto/motto"
)

type MyAgent interface {
	AgentMethod()
}

type dummyAgent struct{}

func (a *dummyAgent) AgentMethod() {}

func TestContainerCanRegisterAndMakeSingletons(t *testing.T) {
	var (
		app      motto.Application
		template MyAgent
		err      error
	)
	container := motto.NewContainer(app)

	agent := &dummyAgent{}

	container.Register(&template, nil, func(ctx context.Context, app motto.Application) (interface{}, error) {
		return agent, nil
	}, &motto.DepOptions{
		Singleton: true,
	})

	err = container.Make(context.TODO(), &template, nil)

	assert.Nil(t, err)
	assert.Equal(t, reflect.TypeOf(agent), reflect.TypeOf(template))
	assert.Equal(t, agent, template.(*dummyAgent))
}

func TestContainerCanRegisterAndMakeNonSingletons(t *testing.T) {
	var (
		app      motto.Application
		template MyAgent
		err      error
	)
	container := motto.NewContainer(app)
	container.Register(&template, nil, func(ctx context.Context, app motto.Application) (interface{}, error) {
		return &dummyAgent{}, nil
	}, &motto.DepOptions{
		Singleton: false,
	})

	err = container.Make(context.TODO(), &template, nil)

	assert.Nil(t, err)
	assert.Equal(t, reflect.TypeOf(&dummyAgent{}), reflect.TypeOf(template))
}
