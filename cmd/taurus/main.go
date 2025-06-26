package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

type ProjectData struct {
	ProjectName  string
	PackageName  string
	TemplatePath string
}

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

	// 准备项目数据
	data := ProjectData{
		ProjectName:  projectName,
		PackageName:  projectName,
		TemplatePath: "templates/basic",
	}

	// 复制项目模板
	if err := copyTemplate(data, projectDir, includeExamples == "y"); err != nil {
		fmt.Printf("复制项目模板失败: %v\n", err)
		return
	}

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

	// 编译项目
	cmd = exec.Command("go", "build", "-o", "bin/app", "cmd/main.go")
	cmd.Dir = projectDir
	if err := cmd.Run(); err != nil {
		fmt.Printf("编译项目失败: %v\n", err)
		return
	}

	fmt.Printf("\n项目 %s 创建成功！\n", projectName)
	fmt.Printf("请进入项目目录并运行：\n")
	fmt.Printf("cd %s\n", projectDir)
	fmt.Printf("./bin/app\n")
}

func copyTemplate(data ProjectData, destDir string, includeExamples bool) error {
	// 获取模板根目录
	templateRoot := filepath.Join(data.TemplatePath)

	// 遍历模板目录
	return filepath.Walk(templateRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 计算目标路径
		relPath, err := filepath.Rel(templateRoot, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(destDir, relPath)

		// 如果是目录，创建它
		if info.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		// 如果不包含示例代码，跳过示例相关文件
		if !includeExamples {
			// 跳过整个 example 目录
			if strings.Contains(path, "/example/") {
				return nil
			}
			// 跳过其他目录中的示例文件
			if strings.Contains(path, "example") || strings.Contains(path, "middleware") {
				return nil
			}
		}

		// 创建空的日志目录
		if strings.Contains(path, "/logs/") {
			return os.MkdirAll(destPath, 0755)
		}

		// 处理模板文件
		if strings.HasSuffix(path, ".go") || strings.HasSuffix(path, ".md") || strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
			return processTemplate(path, destPath, data)
		}

		// 复制其他文件
		return copyFile(path, destPath)
	})
}

func processTemplate(srcPath, destPath string, data ProjectData) error {
	// 读取模板内容
	content, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}

	// 解析模板
	tmpl, err := template.New(filepath.Base(srcPath)).Parse(string(content))
	if err != nil {
		return err
	}

	// 创建目标文件
	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// 执行模板
	return tmpl.Execute(destFile, data)
}

func copyFile(src, dest string) error {
	// 打开源文件
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// 创建目标文件
	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// 复制内容
	_, err = io.Copy(destFile, srcFile)
	return err
}
