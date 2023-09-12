package main

import (
	"github.com/urfave/cli/v2"

	"github.com/dragon2org/chaos/generator/clause"
)

func clauseCMD() *cli.Command {
	var (
		output      string
		typeNames   cli.StringSlice
		destination string
	)

	cmd := cli.Command{
		Name:  "clause",
		Usage: "gorm model clause 生成器",
		UsageText: `model clause 生成器
eg:
	//person.go
	//go:generate chaos clause -t=Person -t=Car -o=person_clause.go

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
			clause.GenClauseV2(typeNames.Value(), output, destination)
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
		Required:    true,
		Destination: &typeNames,
	}

	destFlag := &cli.StringFlag{
		Name:        "destination",
		Aliases:     []string{"d", "dst"},
		Usage:       "生成文件的 package name, 默认跟 type 使用同样的 package name",
		Destination: &destination,
	}

	cmd.Flags = append(cmd.Flags, outputFlag, typeNamesFlag, destFlag)
	return &cmd
}

func init() {
	app.Commands = append(app.Commands, clauseCMD())
}
