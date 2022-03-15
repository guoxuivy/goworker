package cmd

import (
	"cxe/model/store"
	"fmt"

	"github.com/urfave/cli/v2"
	"go.mongodb.org/mongo-driver/bson"
)

type Price struct {
}

func (me *Price) Run(c *cli.Context) error {
	fmt.Println("price month: ", c.String("month"))
	fmt.Println("price user: ", c.Int("user"))
	var storeModel = store.NewStore()
	fmt.Println(storeModel.GetTable())
	storeModel.Find(bson.M{"code": "KF200427012"}, nil)
	switch t := storeModel.GetEntity().(type) {
	case *store.Entity:
		fmt.Println(t.Title)
		fmt.Println("*store.Entity")
	default:
		fmt.Println("unknown")
	}

	// fmt.Println(reflect.TypeOf(storeModel.GetEntity()))
	// fmt.Println(storeModel.GetEntity().(*store.Entity))
	// storeModel.SetEntity(&store.Entity{Code: "test123456"})
	// storeModel.Save()
	// logging.Info(storeModel.GetEntity())
	return nil
}

func init() {
	Commands = append(Commands, &cli.Command{
		Name:  "price",
		Usage: "处理库房价格变动",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "month", Aliases: []string{"m"}, Usage: "指定月份 202101"},
			&cli.IntFlag{Name: "user", Aliases: []string{"u"}, Usage: "指定用户id 125"},
		},
		Action: (&Price{}).Run,
	})
}
