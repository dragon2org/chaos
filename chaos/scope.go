package main

import (
	"github.com/urfave/cli/v2"

	"github.com/dragon2org/chaos/generator/clause"
)

func scopeCMD() *cli.Command {
	var (
		output    string
		typeNames cli.StringSlice
		pkg       string
	)

	cmd := cli.Command{
		Name:  "scope",
		Usage: "gorm model scope 生成器",
		UsageText: `model scope 生成器
eg:
	//person.go
	//go:generate chaos scope -t=Person -t=Car -o=person_scope

	import "fmt"
	
	type Person struct {
		Name string
	}
	
	func (p Person) String() {return p.name}
	
	type Car struct {
		Model string
	}
`,
		Action: func(context *cli.Context) error {
			clause.Gen(typeNames.Value(), output)
			return nil
		},
	}

	outputFlag := &cli.StringFlag{
		Name:        "output",
		Aliases:     []string{"o"},
		Usage:       "输出文件路径",
		Required:    true,
		Destination: &output,
	}

	typeNamesFlag := &cli.StringSliceFlag{
		Name:        "type",
		Aliases:     []string{"t"},
		Usage:       "需要生成的模型",
		Destination: &typeNames,
	}

	pkgFlag := &cli.StringFlag{
		Name:        "package",
		Aliases:     []string{"p"},
		Usage:       "生成文件的 package, 默认跟 type 使用同一 package",
		Destination: &pkg,
	}

	cmd.Flags = append(cmd.Flags, outputFlag, typeNamesFlag, pkgFlag)
	return &cmd
}

func init() {
	app.Commands = append(app.Commands, scopeCMD())
}
