package generator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/stones-hub/taurus-pro-core/pkg/components"
	"github.com/stones-hub/taurus-pro-core/pkg/components/types"
)

type ProjectGenerator struct {
	projectPath        string
	selectedComponents []string
	templateDir        string
}

func NewProjectGenerator(projectPath string, selectedComponents []string) *ProjectGenerator {
	return &ProjectGenerator{
		projectPath:        projectPath,
		selectedComponents: selectedComponents,
	}
}

// SetTemplateDir 设置模板目录
func (g *ProjectGenerator) SetTemplateDir(dir string) {
	g.templateDir = dir
}

func (g *ProjectGenerator) Generate() error {
	// 创建项目目录
	if err := os.MkdirAll(g.projectPath, 0755); err != nil {
		return fmt.Errorf("创建项目目录失败: %v", err)
	}

	// 复制模板文件
	if err := g.copyTemplateFiles(); err != nil {
		return fmt.Errorf("复制模板文件失败: %v", err)
	}

	// 生成 go.mod
	if err := g.generateGoMod(); err != nil {
		return fmt.Errorf("生成 go.mod 失败: %v", err)
	}

	// 扫描并生成 wire.go
	appPath := filepath.Join(g.projectPath, "app")
	if err := g.generateWire(appPath); err != nil {
		return fmt.Errorf("生成 wire.go 失败: %v", err)
	}

	fmt.Println("成功生成项目文件")
	return nil
}

func (g *ProjectGenerator) copyTemplateFiles() error {
	return filepath.Walk(g.templateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 计算相对路径
		relPath, err := filepath.Rel(g.templateDir, path)
		if err != nil {
			return err
		}

		// 跳过 go.mod 文件，因为我们会单独生成它
		if relPath == "go.mod" {
			return nil
		}

		targetPath := filepath.Join(g.projectPath, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		return g.copyFile(path, targetPath, info.Mode())
	})
}

func (g *ProjectGenerator) copyFile(src, dst string, mode os.FileMode) error {
	// 读取源文件
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	// 替换模板变量
	content := string(data)
	content = strings.ReplaceAll(content, "{{.ProjectName}}", filepath.Base(g.projectPath))

	// 创建目标文件
	return os.WriteFile(dst, []byte(content), mode)
}

// generateWire 生成 wire.go 文件
func (g *ProjectGenerator) generateWire(appPath string) error {
	// 确保 app 目录存在
	if err := os.MkdirAll(appPath, 0755); err != nil {
		return fmt.Errorf("创建 app 目录失败: %v", err)
	}

	selectedComponents := make([]types.Component, 0)

	for _, comp := range g.selectedComponents {
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
	tidyCmd.Dir = g.projectPath
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
	moduleName := filepath.Base(g.projectPath)

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
	for _, selectedComp := range g.selectedComponents {
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

	goModPath := filepath.Join(g.projectPath, "go.mod")
	return os.WriteFile(goModPath, []byte(content), 0644)
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
