# 性能测试包

本目录包含各种性能测试工具和测试用例。

## 目录结构

```
performance/
  ├── generator/       # 负载生成器
  │   └── load_generator.go
  ├── memory/         # 内存相关测试
  │   └── memory_test.go
  └── README.md
```

## 使用方法

### 内存泄漏测试

1. 启动测试：
```bash
go test -v ./test/performance/memory -run TestMemoryLeak
```

2. 查看实时内存使用：
```bash
go tool pprof http://localhost:6060/debug/pprof/heap
```

3. 生成内存使用图表：
```bash
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/heap
```

### 负载生成器

负载生成器用于模拟真实用户行为，可以：
- 模拟多个并发用户
- 随机执行各种操作
- 可配置测试持续时间
- 支持优雅停止

## 注意事项

1. 测试前确保应用配置了 pprof
2. 测试时间建议至少 10 分钟
3. 定期检查内存快照
4. 关注内存增长趋势 