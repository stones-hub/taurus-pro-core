package service

// ExampleService 示例服务
type ExampleService struct{}

// NewExampleService 创建示例服务实例
func NewExampleService() *ExampleService {
	return &ExampleService{}
}

// DoSomething 执行某些业务逻辑
func (s *ExampleService) DoSomething() string {
	return "这是服务层的响应"
}
