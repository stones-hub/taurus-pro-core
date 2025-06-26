package service

import (
	"github.com/google/wire"
)

// IndexService 示例服务
type IndexService struct{}

// IndexServiceSet wire provider set
var IndexServiceSet = wire.NewSet(wire.Struct(new(IndexService), "*"))

func (s *IndexService) Home() string {
	return "Hello from my-service!"
}

type IndexService2 struct{}
