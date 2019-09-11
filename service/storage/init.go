package storage

import "bqtail/task"

func InitRegistry(registry task.Registry, service Service) {
	registry.RegisterService("storage", service)
	registry.RegisterAction("move", task.NewServiceAction("storage", MoveRequest{}))
	registry.RegisterAction("delete", task.NewServiceAction("storage", DeleteRequest{}))
}