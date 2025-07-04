package generator

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/stones-hub/taurus-pro-core/pkg/components/types"
	"github.com/stones-hub/taurus-pro-core/pkg/scanner"
)

// wire.go 模板
const wireTemplate = `//go:build wireinject
// +build wireinject

package app

import (
	"fmt"
	"github.com/google/wire"
	"github.com/stones-hub/taurus-pro-config/pkg/config"
{{- range .ComponentImports}}
	{{- range .Path}}
	{{if hasAlias .}}{{getAlias .}} "{{getPath .}}"{{else}}"{{.}}"{{end}}
	{{- end}}
{{- end}}
{{- range .Imports}}
	"{{$.ModuleName}}/{{.}}"
{{- end}}
)

// ConfigOptions 配置选项
type ConfigOptions struct {
	ConfigPath   string
	Env          string
	PrintEnable  bool
}

// Taurus 应用程序结构
type Taurus struct {
	Config *config.Config
{{- range .ComponentFields}}
	{{.Name}} {{.Type}}
{{- end}}
{{- range .Fields}}
	{{.}}
{{- end}}
}

// ProvideConfigComponent 注入配置模块
func ProvideConfigComponent(opts *ConfigOptions) (*config.Config, error) {
	configComponent := config.New(config.WithPrintEnable(opts.PrintEnable))
	if err := configComponent.Initialize(opts.ConfigPath, opts.Env); err != nil {
		return nil, fmt.Errorf("failed to initialize config: %v", err)
	}
	return configComponent, nil
}

{{- range .ComponentProviders}}
{{.Provider}}

{{- end}}

// buildTaurus 构建应用程序
func buildTaurus(opts *ConfigOptions) (*Taurus, func(), error) {
	wire.Build(
		// 配置组件
		ProvideConfigComponent,

		// 组件提供者
{{- range .ComponentProviders}}
		{{.ProviderName}},
{{- end}}

		// 应用结构
		wire.Struct(new(Taurus), "*"),
		// 扫描到的 provider sets
{{- range .ProviderSets}}
		{{.}},
{{- end}}
	)

	return new(Taurus), nil, nil
}`

// hasAlias 检查路径是否包含别名
func hasAlias(path string) bool {
	return strings.Contains(path, "@")
}

// getAlias 从路径中获取别名
func getAlias(path string) string {
	parts := strings.Split(path, "@")
	if len(parts) > 1 {
		return parts[0]
	}
	return ""
}

// getPath 从路径中获取实际路径
func getPath(path string) string {
	parts := strings.Split(path, "@")
	if len(parts) > 1 {
		return parts[1]
	}
	return path
}

func GenerateWire(scannerPath string, components []types.Component) error {
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

	// 4. 处理组件数据
	var componentData struct {
		ComponentImports []struct {
			Path []string
		}
		ComponentFields []struct {
			Name string
			Type string
		}
		ComponentProviders []struct {
			Provider     string
			ProviderName string
		}
	}

	// 处理每个组件
	for _, comp := range components {
		if comp.IsCustom && len(comp.Wire) > 0 {
			for _, wire := range comp.Wire {
				// 添加导入
				componentData.ComponentImports = append(componentData.ComponentImports, struct {
					Path []string
				}{
					Path: wire.RequirePath,
				})

				// 添加字段
				componentData.ComponentFields = append(componentData.ComponentFields, struct {
					Name string
					Type string
				}{
					Name: wire.Name,
					Type: wire.Type,
				})

				// 创建模板以处理 Provider 字符串
				tmpl, err := template.New("provider").Parse(wire.Provider)
				if err != nil {
					return fmt.Errorf("解析 Provider 模板失败: %v", err)
				}

				var providerStr strings.Builder
				err = tmpl.Execute(&providerStr, wire)
				if err != nil {
					return fmt.Errorf("执行 Provider 模板失败: %v", err)
				}

				// 添加Provider
				componentData.ComponentProviders = append(componentData.ComponentProviders, struct {
					Provider     string
					ProviderName string
				}{
					Provider:     providerStr.String(),
					ProviderName: wire.ProviderName,
				})

			}
		}
	}

	// 5. 生成 wire.go 文件
	data := struct {
		ModuleName         string
		Imports            []string
		ProviderSets       []string
		Fields             []string
		ComponentImports   []struct{ Path []string }
		ComponentFields    []struct{ Name, Type string }
		ComponentProviders []struct{ Provider, ProviderName string }
	}{
		ModuleName:         moduleName,
		Imports:            scanner.GenerateWireImports(),
		ProviderSets:       scanner.GenerateWireProviderSets(),
		Fields:             scanner.GenerateApplicationFields(),
		ComponentImports:   componentData.ComponentImports,
		ComponentFields:    componentData.ComponentFields,
		ComponentProviders: componentData.ComponentProviders,
	}

	// 创建模板
	tmpl := template.New("wire")

	// 添加模板函数
	tmpl.Funcs(template.FuncMap{
		"hasAlias": hasAlias,
		"getAlias": getAlias,
		"getPath":  getPath,
	})

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
