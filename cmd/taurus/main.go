package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/stones-hub/taurus-pro-core/pkg/generator"
	"gopkg.in/yaml.v3"
)

var (
	projectName string
	projectPath string
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "taurus",
		Short: "Taurus Pro CLI tool",
		Long:  `Taurus Pro is a CLI tool for creating and managing Go microservice projects`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	var createCmd = &cobra.Command{
		Use:   "create [project-name]",
		Short: "Create a new Taurus Pro project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName = args[0]
			return runCreate()
		},
	}

	rootCmd.AddCommand(createCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// generateComponentConfig 生成组件配置信息
func generateComponentConfig(selectedComponents []string) ([]byte, error) {
	// 创建组件配置映射
	components := make(map[string]map[string]interface{})

	// 遍历所有选中的组件
	for _, name := range selectedComponents {
		comp, exists := generator.GetComponentByName(name)
		if !exists {
			continue
		}

		// 构建组件配置
		componentConfig := map[string]interface{}{
			"package":     comp.Package,
			"version":     comp.Version,
			"description": comp.Description,
			"enabled":     true,
			"is_custom":   comp.IsCustom,
		}

		// 如果有依赖，添加依赖信息
		if len(comp.Dependencies) > 0 {
			componentConfig["dependencies"] = comp.Dependencies
		}

		// 将组件配置添加到映射中
		components[name] = componentConfig
	}

	// 创建最终的配置结构
	config := map[string]interface{}{
		"components": components,
	}

	// 转换为YAML格式
	return yaml.Marshal(config)
}

func runCreate() error {
	// 获取可选组件
	optionalComponents := generator.GetOptionalComponents()
	componentOptions := make([]string, 0, len(optionalComponents))
	for _, comp := range optionalComponents {
		componentOptions = append(componentOptions, fmt.Sprintf("%s (%s)", comp.Description, comp.Package))
	}

	// 获取必需组件
	requiredComponents := generator.GetRequiredComponents()
	var requiredComponentNames []string
	for _, comp := range requiredComponents {
		requiredComponentNames = append(requiredComponentNames, comp.Name)
	}

	// 定义问题
	questions := []*survey.Question{
		{
			Name: "projectPath",
			Prompt: &survey.Input{
				Message: "请输入项目路径:",
				Default: filepath.Join(".", projectName),
				Help:    "项目将被创建在这个目录下",
			},
			Validate: func(val interface{}) error {
				str, ok := val.(string)
				if !ok {
					return fmt.Errorf("输入的路径无效")
				}
				if str == "" {
					return fmt.Errorf("路径不能为空")
				}
				return nil
			},
		},
	}

	// 只有在有可选组件的情况下才添加组件选择问题
	if len(componentOptions) > 0 {
		questions = append(questions, &survey.Question{
			Name: "components",
			Prompt: &survey.MultiSelect{
				Message: "选择要包含的组件:",
				Options: componentOptions,
				Help:    "使用空格键选择/取消选择组件，按回车确认",
			},
		})
	}

	answers := struct {
		ProjectPath string   `survey:"projectPath"`
		Components  []string `survey:"components"`
	}{}

	// 执行问题
	if err := survey.Ask(questions, &answers); err != nil {
		return fmt.Errorf("问卷调查失败: %v", err)
	}

	// 确保项目路径包含项目名称
	projectPath = answers.ProjectPath
	// 如果输入的路径不是以项目名结尾，则将项目名添加到路径中
	if !strings.HasSuffix(projectPath, projectName) {
		projectPath = filepath.Join(projectPath, projectName)
	}

	// 将选中的组件转换为组件名称
	var selectedComponents []string
	for _, comp := range answers.Components {
		// 从选项字符串中提取组件包名
		packageStart := strings.Index(comp, "(") + 1
		packageEnd := strings.Index(comp, ")")
		if packageStart > 0 && packageEnd > packageStart {
			packageName := comp[packageStart:packageEnd]
			// 查找对应的组件名称
			for _, c := range optionalComponents {
				if c.Package == packageName {
					selectedComponents = append(selectedComponents, c.Name)
					break
				}
			}
		}
	}

	// 添加必需组件
	selectedComponents = append(selectedComponents, requiredComponentNames...)

	// 生成组件配置
	componentConfig, err := generateComponentConfig(selectedComponents)
	if err != nil {
		return fmt.Errorf("生成组件配置失败: %v", err)
	}

	// 创建项目生成器
	gen := generator.NewProjectGenerator(projectPath, selectedComponents)

	// 设置组件配置
	gen.SetComponentConfig(componentConfig)

	// 生成项目
	if err := gen.Generate(); err != nil {
		return fmt.Errorf("生成项目失败: %v", err)
	}

	fmt.Printf("\n项目已成功创建在: %s\n", projectPath)

	// 显示已包含的组件
	fmt.Println("\n已包含的组件:")
	fmt.Println("必需组件:")
	for _, name := range requiredComponentNames {
		if comp, exists := generator.GetComponentByName(name); exists {
			fmt.Printf("- %s (%s)\n", comp.Description, comp.Package)
		}
	}

	if len(selectedComponents) > len(requiredComponentNames) {
		fmt.Println("\n可选组件:")
		for _, name := range selectedComponents {
			// 跳过必需组件
			isRequired := false
			for _, req := range requiredComponentNames {
				if req == name {
					isRequired = true
					break
				}
			}
			if !isRequired {
				if comp, exists := generator.GetComponentByName(name); exists {
					fmt.Printf("- %s (%s)\n", comp.Description, comp.Package)
				}
			}
		}
	}

	return nil
}
