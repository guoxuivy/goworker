package cmd

import (
	"cxe/util/logging"
	"os"
	"runtime/debug"

	"github.com/urfave/cli/v2"
)

// 可执行命令接口
type ICmd interface {
	Run(c *cli.Context) error
}

// 存取全部命令
var Commands []*cli.Command

func Run() {

	defer func() {
		// panic 记录到日志
		if p := recover(); p != nil {
			logging.Error("panic recover! p:", p, string(debug.Stack()))
			debug.PrintStack()
		}
	}()

	app := &cli.App{
		Name:    "cxe",
		Version: "v0.0.1",
		Flags: []cli.Flag{
			// &cli.BoolFlag{
			// 	Name:    "version",
			// 	Aliases: []string{"v"},
			// 	Usage:   "print only the version",
			// },
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Config file",
			},
		},
		// Commands: []*cli.Command{
		// 	{
		// 		Name:  "price",
		// 		Usage: "处理库房价格变动",
		// 		// ArgsUsage: "[month]",
		// 		Flags: []cli.Flag{
		// 			&cli.StringFlag{Name: "month", Aliases: []string{"m"}, Usage: "指定月份 202101"},
		// 		},
		// 		Action: (&cmd.Price{}).Run,
		// 	},
		// },
	}
	app.Commands = Commands
	// app.Commands = append(app.Commands, &cli.Command{
	// 	Name:  "area",
	// 	Usage: "处理库房面积变动",
	// 	// ArgsUsage: "[month]",
	// 	Flags: []cli.Flag{
	// 		&cli.StringFlag{Name: "month", Aliases: []string{"m"}, Usage: "指定月份 202101"},
	// 	},
	// 	Action: (&cmd.Price{}).Run,
	// })

	app.Run(os.Args)
}
