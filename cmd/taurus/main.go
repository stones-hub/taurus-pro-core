package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	// 1. 获取项目名称
	fmt.Print("请输入项目名称: ")
	projectName, _ := reader.ReadString('\n')
	projectName = strings.TrimSpace(projectName)

	// 2. 获取项目路径
	fmt.Print("请输入项目路径 (默认为当前目录): ")
	projectPath, _ := reader.ReadString('\n')
	projectPath = strings.TrimSpace(projectPath)
	if projectPath == "" {
		projectPath = "."
	}

	// 3. 是否包含示例代码
	fmt.Print("是否需要包含示例代码? (y/n): ")
	includeExamples, _ := reader.ReadString('\n')
	includeExamples = strings.TrimSpace(strings.ToLower(includeExamples))

	// 创建项目目录
	projectDir := filepath.Join(projectPath, projectName)
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		fmt.Printf("创建项目目录失败: %v\n", err)
		return
	}

	// 创建项目结构
	createProjectStructure(projectDir, includeExamples == "y")

	// 初始化 go.mod
	cmd := exec.Command("go", "mod", "init", projectName)
	cmd.Dir = projectDir
	if err := cmd.Run(); err != nil {
		fmt.Printf("初始化 go.mod 失败: %v\n", err)
		return
	}

	// 添加依赖
	cmd = exec.Command("go", "get", "github.com/stones-hub/taurus-pro-http@v0.0.1")
	cmd.Dir = projectDir
	if err := cmd.Run(); err != nil {
		fmt.Printf("添加依赖失败: %v\n", err)
		return
	}

	fmt.Printf("\n项目 %s 创建成功！\n", projectName)
	fmt.Printf("请进入项目目录并运行：\n")
	fmt.Printf("cd %s\n", projectDir)
	fmt.Printf("go run cmd/main.go\n")
}

func createProjectStructure(projectDir string, includeExamples bool) {
	// 创建基本目录结构
	dirs := []string{
		"cmd",
		"internal/config",
		"internal/handler",
		"internal/middleware",
		"internal/model",
		"internal/service",
		"pkg",
	}

	for _, dir := range dirs {
		os.MkdirAll(filepath.Join(projectDir, dir), 0755)
	}

	// 创建主程序文件
	createMainFile(projectDir, includeExamples)

	// 创建配置文件
	createConfigFile(projectDir)

	// 如果需要示例代码
	if includeExamples {
		createExampleHandler(projectDir)
		createExampleMiddleware(projectDir)
	}

	// 创建 README.md
	createReadme(projectDir, filepath.Base(projectDir))
}

func createMainFile(projectDir string, includeExamples bool) {
	mainContent := []string{
		"package main",
		"",
		"import (",
		"	\"log\"",
		"",
		"	\"github.com/stones-hub/taurus-pro-http/server\"",
	}

	if includeExamples {
		mainContent = append(mainContent,
			fmt.Sprintf("	\"%s/internal/handler\"", filepath.Base(projectDir)),
			fmt.Sprintf("	\"%s/internal/middleware\"", filepath.Base(projectDir)),
		)
	}

	mainContent = append(mainContent,
		")",
		"",
		"func main() {",
		"	// 创建 HTTP 服务器实例",
		"	srv := server.New()",
		"",
	)

	if includeExamples {
		mainContent = append(mainContent,
			"	// 添加全局中间件",
			"	srv.Use(middleware.Logger())",
			"",
			"	// 注册路由",
			"	handler.NewExampleHandler().Register(srv)",
			"",
		)
	}

	mainContent = append(mainContent,
		"	// 启动服务器",
		"	if err := srv.Start(); err != nil {",
		"		log.Fatalf(\"服务器启动失败: %v\", err)",
		"	}",
		"}",
	)

	content := strings.Join(mainContent, "\n")
	os.WriteFile(filepath.Join(projectDir, "cmd", "main.go"), []byte(content), 0644)
}

func createConfigFile(projectDir string) {
	content := strings.Join([]string{
		"package config",
		"",
		"import (",
		"	\"github.com/stones-hub/taurus-pro-http/server\"",
		")",
		"",
		"// Config 应用配置结构",
		"type Config struct {",
		"	Server *server.Config `yaml:\"server\"`",
		"}",
		"",
		"// DefaultConfig 返回默认配置",
		"func DefaultConfig() *Config {",
		"	return &Config{",
		"		Server: &server.Config{",
		"			Port: 8080,",
		"		},",
		"	}",
		"}",
	}, "\n")

	os.WriteFile(filepath.Join(projectDir, "internal", "config", "config.go"), []byte(content), 0644)
}

func createExampleHandler(projectDir string) {
	content := strings.Join([]string{
		"package handler",
		"",
		"import (",
		"	\"github.com/stones-hub/taurus-pro-http/server\"",
		")",
		"",
		"// ExampleHandler 示例处理器",
		"type ExampleHandler struct{}",
		"",
		"// NewExampleHandler 创建示例处理器实例",
		"func NewExampleHandler() *ExampleHandler {",
		"	return &ExampleHandler{}",
		"}",
		"",
		"// Register 注册路由",
		"func (h *ExampleHandler) Register(srv *server.Server) {",
		"	srv.GET(\"/example\", h.HandleExample)",
		"}",
		"",
		"// HandleExample 处理示例请求",
		"func (h *ExampleHandler) HandleExample(ctx *server.Context) {",
		"	ctx.JSON(200, map[string]string{",
		"		\"message\": \"这是一个示例响应\",",
		"	})",
		"}",
	}, "\n")

	os.WriteFile(filepath.Join(projectDir, "internal", "handler", "example.go"), []byte(content), 0644)
}

func createExampleMiddleware(projectDir string) {
	content := strings.Join([]string{
		"package middleware",
		"",
		"import (",
		"	\"log\"",
		"	\"time\"",
		"",
		"	\"github.com/stones-hub/taurus-pro-http/server\"",
		")",
		"",
		"// Logger 创建一个日志中间件",
		"func Logger() server.HandlerFunc {",
		"	return func(ctx *server.Context) {",
		"		start := time.Now()",
		"",
		"		// 处理请求",
		"		ctx.Next()",
		"",
		"		// 记录请求信息",
		"		log.Printf(\"[%s] %s %s %v\",",
		"			ctx.Request.Method,",
		"			ctx.Request.URL.Path,",
		"			ctx.Request.RemoteAddr,",
		"			time.Since(start),",
		"		)",
		"	}",
		"}",
	}, "\n")

	os.WriteFile(filepath.Join(projectDir, "internal", "middleware", "logger.go"), []byte(content), 0644)
}

func createReadme(projectDir, projectName string) {
	content := strings.Join([]string{
		fmt.Sprintf("# %s", projectName),
		"",
		"这是一个基于 taurus-pro-http 的 Web 服务项目。",
		"",
		"## 项目结构",
		"",
		"```",
		".  ",
		"├── cmd                     # 命令行工具",
		"│   └── main.go            # 主程序入口",
		"├── internal               # 内部代码",
		"│   ├── config            # 配置",
		"│   ├── handler           # HTTP 处理器",
		"│   ├── middleware        # 中间件",
		"│   ├── model            # 数据模型",
		"│   └── service          # 业务逻辑层",
		"└── pkg                   # 可重用的包",
		"```",
		"",
		"## 快速开始",
		"",
		"运行服务器：",
		"",
		"```bash",
		"go run cmd/main.go",
		"```",
		"",
		"## 配置说明",
		"",
		"配置文件位于 `internal/config` 目录下，支持以下配置：",
		"",
		"- HTTP 服务器配置",
		"- 数据库配置",
		"- 日志配置",
		"- 中间件配置",
		"",
		"## 许可证",
		"",
		"Apache-2.0 license",
	}, "\n")

	os.WriteFile(filepath.Join(projectDir, "README.md"), []byte(content), 0644)
}
