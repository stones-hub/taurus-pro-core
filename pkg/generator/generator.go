package generator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/stones-hub/taurus-pro-core/pkg/components"
	"github.com/stones-hub/taurus-pro-core/pkg/components/types"
)

// Generator 定义项目生成器接口
type Generator interface {
	Generate() error
}

// ProjectGenerator 项目生成器
type ProjectGenerator struct {
	ProjectPath string
	Components  []string
	TemplateDir string
}

// getSystemGoVersion 获取系统的 Go 版本
func getSystemGoVersion() (string, error) {
	cmd := exec.Command("go", "version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("获取 Go 版本失败: %v", err)
	}

	version := strings.TrimSpace(string(output))
	parts := strings.Split(version, " ")
	if len(parts) < 3 {
		return "", fmt.Errorf("无法解析 Go 版本信息: %s", version)
	}

	versionNum := strings.TrimPrefix(parts[2], "go")
	versionParts := strings.Split(versionNum, ".")
	if len(versionParts) < 2 {
		return "", fmt.Errorf("无法解析版本号: %s", versionNum)
	}

	return fmt.Sprintf("%s.%s", versionParts[0], versionParts[1]), nil
}

// NewProjectGenerator 创建新的项目生成器
func NewProjectGenerator(projectPath string, selectedComponents []string) *ProjectGenerator {
	// 获取当前包的路径
	_, currentFile, _, _ := runtime.Caller(0)
	templateDir := filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(currentFile))), "templates")

	// 验证组件依赖关系
	if err := components.ValidateComponents(selectedComponents); err != nil {
		fmt.Printf("警告: 组件依赖验证失败: %v\n", err)
	}

	return &ProjectGenerator{
		ProjectPath: projectPath,
		Components:  selectedComponents,
		TemplateDir: templateDir,
	}
}

// Generate 生成项目结构
func (g *ProjectGenerator) Generate() error {
	required := components.GetRequiredComponents()
	fmt.Printf("开始生成项目，包含基础组件: ")
	for i, comp := range required {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(comp.Name)
	}
	fmt.Println()

	optional := components.GetOptionalComponents()
	if len(optional) > 0 {
		fmt.Printf("可选组件: ")
		for i, comp := range optional {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(comp.Name)
		}
		fmt.Println()
	}

	if _, err := os.Stat(g.TemplateDir); os.IsNotExist(err) {
		return fmt.Errorf("模板目录不存在: %s", g.TemplateDir)
	}

	if err := os.MkdirAll(g.ProjectPath, 0755); err != nil {
		return fmt.Errorf("创建项目目录失败: %v", err)
	}

	// 首先复制所有模板文件
	err := filepath.Walk(g.TemplateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(g.TemplateDir, path)
		if err != nil {
			return err
		}

		if relPath == "go.mod" {
			return nil
		}

		targetPath := filepath.Join(g.ProjectPath, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		return g.copyFile(path, targetPath, info.Mode())
	})

	if err != nil {
		return fmt.Errorf("复制模板文件失败: %v", err)
	}

	// 生成 go.mod 文件
	if err := g.generateGoMod(); err != nil {
		return fmt.Errorf("生成 go.mod 失败: %v", err)
	}

	// 扫描并生成 wire.go
	appPath := filepath.Join(g.ProjectPath, "app")
	if err := g.generateWire(appPath); err != nil {
		return fmt.Errorf("生成 wire.go 失败: %v", err)
	}

	fmt.Println("成功生成项目文件")
	return nil
}

// generateWire 生成 wire.go 文件
func (g *ProjectGenerator) generateWire(appPath string) error {
	// 确保 app 目录存在
	if err := os.MkdirAll(appPath, 0755); err != nil {
		return fmt.Errorf("创建 app 目录失败: %v", err)
	}

	selectedComponents := make([]types.Component, 0)

	for _, comp := range g.Components {
		component, ok := components.GetComponentByName(comp)
		if !ok {
			return fmt.Errorf("组件 %s 不存在", comp)
		}
		selectedComponents = append(selectedComponents, component)
	}

	// 扫描并生成 wire.go
	if err := GenerateWire(appPath, selectedComponents); err != nil {
		return fmt.Errorf("生成 wire.go 失败: %v", err)
	}

	// 执行 go mod tidy
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = g.ProjectPath
	if output, err := tidyCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("执行 go mod tidy 失败: %v\n输出: %s", err, output)
	}

	// 对wire.go 文件执行 go fmt
	fmtCmd := exec.Command("go", "fmt", filepath.Join(appPath, "wire.go"))
	if output, err := fmtCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("执行 go fmt 失败: %v\n输出: %s", err, output)
	}

	// 执行 wire 命令生成实现
	cmd := exec.Command("wire")
	cmd.Dir = appPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("执行 wire 命令失败: %v\n输出: %s", err, output)
	}

	return nil
}

// generateGoMod 生成新的 go.mod 文件
func (g *ProjectGenerator) generateGoMod() error {
	moduleName := filepath.Base(g.ProjectPath)

	goVersion, err := getSystemGoVersion()
	if err != nil {
		fmt.Printf("警告: 获取系统 Go 版本失败，将使用默认版本 1.21: %v\n", err)
		goVersion = "1.21"
	}

	// 生成require部分
	requires := []string{
		"require (",
	}

	// 添加基础组件
	addedPackages := make(map[string]bool)
	// 首先添加必需组件
	for _, comp := range components.AllComponents {
		if comp.Required {
			if !addedPackages[comp.Package] {
				requires = append(requires, "\t"+comp.Package+" "+comp.Version)
				addedPackages[comp.Package] = true
			}
		}
	}

	// 添加选择的可选组件
	for _, selectedComp := range g.Components {
		for _, comp := range components.AllComponents {
			if comp.Name == selectedComp && !comp.Required {
				if !addedPackages[comp.Package] {
					requires = append(requires, "\t"+comp.Package+" "+comp.Version)
					addedPackages[comp.Package] = true
				}
			}
		}
	}

	requires = append(requires, ")")

	// 准备go.mod内容
	content := fmt.Sprintf(`module %s

go %s

%s`, moduleName, goVersion, strings.Join(requires, "\n"))

	goModPath := filepath.Join(g.ProjectPath, "go.mod")
	return os.WriteFile(goModPath, []byte(content), 0644)
}

// copyFile 复制单个文件
func (g *ProjectGenerator) copyFile(src, dst string, mode os.FileMode) error {
	// 读取源文件内容
	content, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	// 替换模板变量
	fileContent := string(content)
	fileContent = strings.ReplaceAll(fileContent, "{{.ProjectName}}", filepath.Base(g.ProjectPath))

	// 写入目标文件
	return os.WriteFile(dst, []byte(fileContent), mode)
}
