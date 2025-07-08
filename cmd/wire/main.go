package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/stones-hub/taurus-pro-core/pkg/project"
)

// 生成 wire.go 文件
// 命令行输入项目根目录地址
// 用于生成一个工具，方便打包的时候，自动扫描项目中的 provider set 文件，并重新生成 wire.go 文件
// go build -o templates/scripts/wire/gen_wire cmd/wire/main.go
// 使用方法： （需要时绝对路径）
// ./gen_wire -project-root /Users/stones/work/taurus-pro-core
func main() {
	projectRoot := flag.String("project-root", "", "项目根目录")
	flag.Parse()

	if *projectRoot == "" {
		log.Fatal("项目根目录不能为空")
	}

	appPath := filepath.Join(*projectRoot, "app")

	// 先删除 app 目录下的 wire.go 文件
	os.RemoveAll(filepath.Join(appPath, "wire.go"))

	// 确保 app 目录存在
	if err := os.MkdirAll(appPath, 0755); err != nil {
		log.Fatalf("创建 app 目录失败: %v", err)
	}

	// 扫描并生成 wire.go
	if err := project.GenerateProjectWire(appPath); err != nil {
		log.Fatalf("生成 wire.go 失败: %v", err)
	}

	// 执行 go mod tidy
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = *projectRoot
	if output, err := tidyCmd.CombinedOutput(); err != nil {
		log.Fatalf("执行 go mod tidy 失败: %v\n输出: %s", err, output)
	}

	// 对wire.go 文件执行 go fmt
	fmtCmd := exec.Command("go", "fmt", filepath.Join(appPath, "wire.go"))
	if output, err := fmtCmd.CombinedOutput(); err != nil {
		log.Fatalf("执行 go fmt 失败: %v\n输出: %s", err, output)
	}

	// 执行 wire 命令生成实现
	cmd := exec.Command("wire")
	cmd.Dir = appPath
	if output, err := cmd.CombinedOutput(); err != nil {
		log.Fatalf("执行 wire 命令失败: %v\n输出: %s", err, output)
	}
}
