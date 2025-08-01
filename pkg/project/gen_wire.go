package project

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/stones-hub/taurus-pro-core/pkg/scanner"
)

// wire.go 模板
const wireTemplate = `//go:build wireinject
// +build wireinject

package app

import (
	"github.com/google/wire"
{{- range .Imports}}
	"{{$.ModuleName}}/{{.}}"
{{- end}}
)

// Injector 应用程序结构
type Injector struct {
{{- range .Fields}}
	{{.}}
{{- end}}
}

// buildInjector 构建应用程序
func buildInjector() (*Injector, func(), error) {
	wire.Build(
		// 应用结构
		wire.Struct(new(Injector), "*"),
		// 扫描到的 provider sets
{{- range .ProviderSets}}
		{{.}},
{{- end}}
	)

	return new(Injector), nil, nil
}`

func GenerateProjectWire(scannerPath string) error {
	// 获取项目根目录（app 目录的父目录）
	projectRoot := filepath.Dir(scannerPath)

	// 获取模块名称
	moduleName, err := getModuleName(projectRoot)
	if err != nil {
		return fmt.Errorf("获取模块名称失败: %v", err)
	}

	// 1. 创建扫描器
	scanner := scanner.NewScanner(projectRoot, moduleName)

	// 2. 扫描 app 目录下的所有 provider sets
	if err := scanner.ScanDir(scannerPath); err != nil {
		return fmt.Errorf("扫描目录失败: %v", err)
	}

	// 3. 获取扫描结果
	providerSets := scanner.GetProviderSets()
	log.Printf("Found %d provider sets:", len(providerSets))
	for _, set := range providerSets {
		log.Printf("  - %s.%s (%s)", filepath.Base(set.PkgPath), set.Name, set.StructType)
	}

	data := struct {
		ModuleName   string
		Imports      []string
		ProviderSets []string
		Fields       []string
	}{
		ModuleName:   moduleName,
		Imports:      scanner.GenerateWireImports(),
		ProviderSets: scanner.GenerateWireProviderSets(),
		Fields:       scanner.GenerateApplicationFields(),
	}

	// 创建模板
	tmpl := template.New("wire")

	// 解析模板
	tmpl, err = tmpl.Parse(wireTemplate)
	if err != nil {
		return fmt.Errorf("解析模板失败: %v", err)
	}

	// 创建 wire.go 文件
	f, err := os.Create(filepath.Join(scannerPath, "wire.go"))
	if err != nil {
		return fmt.Errorf("创建 wire.go 失败: %v", err)
	}
	defer f.Close()

	// 执行模板
	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("执行模板失败: %v", err)
	}

	log.Println("Successfully generated wire.go")
	return nil
}

// getModuleName 从 go.mod 文件中获取模块名称
func getModuleName(projectRoot string) (string, error) {
	goModPath := filepath.Join(projectRoot, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return "", fmt.Errorf("读取 go.mod 失败: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}

	return "", fmt.Errorf("在 go.mod 中未找到模块名称")
}
