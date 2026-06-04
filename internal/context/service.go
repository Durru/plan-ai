package context

type Service struct{ Registry }

func NewService(repo Repository) Service { return Service{Registry: NewRegistry(repo)} }
