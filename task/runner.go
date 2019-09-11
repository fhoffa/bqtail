package task

import (
	"context"
	"fmt"
	"github.com/viant/toolbox"
)

//RunWithService handlers service request or error
func Run(ctx context.Context, registry Registry, action *Action) error {
	serviceAction, err := registry.Action(action.Action)
	if err != nil {
		return err
	}
	request, err := serviceAction.NewRequest(action)
	if err != nil {
		return err
	}
	return RunWithService(ctx, registry, serviceAction.Service, request)
}

//RunWithService handlers service request or error
func RunWithService(ctx context.Context, registry Registry, serviceName string, request Request) error {
	service, err := registry.Service(serviceName)
	if err != nil {
		return err
	}
	fmt.Printf("running %T\n", service)
	toolbox.Dump(request)
	return service.Run(ctx, request)
}