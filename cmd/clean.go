package cmd

import (
	"context"
	"cxe/db"
	"cxe/model/park"
	"cxe/model/store"
	"cxe/util/logging"
	"errors"
	"fmt"
	"log"

	"github.com/urfave/cli/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Clean struct {
	table *mongo.Collection
	dns   string
}

func (cl *Clean) Run(c *cli.Context) error {
	// fmt.Println("price month: ", c.String("month"))
	// fmt.Println("price user: ", c.Int("user"))
	// var storeModel = store.NewStore()
	// fmt.Println(storeModel.GetTable())
	// storeModel.Find(bson.M{"code": "KF200427012"}, nil)
	// switch t := storeModel.GetEntity().(type) {
	// case *store.Entity:
	// 	fmt.Println(t.Title)
	// 	fmt.Println("*store.Entity")
	// default:
	// 	fmt.Println("unknown")
	// }

	cl.dns = c.String("dns")
	if cl.dns == "" {
		fmt.Println("dns unknown")
		return nil
	}
	if c.Bool("mark") {
		cl.mark(c)
	}
	if c.Bool("clean") {
		cl.clean(c)
	}
	return nil
}

// 标记用户表身份证agent、文章封面article、品牌地产brand_estate、 合同附件contract_attachment[attach 逗号分割]、企业enterprise、委托合同entrust
// func (cl *Clean) markOther(c *cli.Context) error {
// 	return nil
// }

// 标记库房、园区图片
func (cl *Clean) mark(c *cli.Context) {
	filter := bson.M{"delete_status": 1, "overall_view_picurl": bson.M{"$exists": true}}
	var cate string = "store" // park or store
	var myModel db.IMgo
	if cate == "park" {
		myModel = park.NewPark()
	} else {
		myModel = store.NewStore()
	}
	myModel.Chunk(filter, bson.M{"_id": -1}, 10, func(i int64, list []db.IEntity) error {
		var idss []string
		for _, v := range list {
			// 这里有点扯
			if cate == "park" {
				tmp := v.(*park.Entity)
				idss = append(idss, tmp.OverallViewPicurl, tmp.InteriorPicurl, tmp.LocationPicurl, tmp.PeripheryPicurl)
				idss = append(idss, tmp.OtherPicurl...)

			} else {
				tmp := v.(*store.Entity)
				idss = append(idss, tmp.OverallViewPicurl, tmp.InteriorPicurl, tmp.LocationPicurl, tmp.PeripheryPicurl)
				idss = append(idss, tmp.OtherPicurl...)
			}
		}
		// 空值过滤
		for i := 0; i < len(idss); {
			if idss[i] == "" {
				idss = append(idss[:i], idss[i+1:]...)
			} else {
				i++
			}
		}
		fmt.Printf("第%v页\n:", i)
		// fmt.Println(idss)
		// cl.mk(idss, cate)
		if i == 3000 {
			return errors.New("发生错误")
		}
		return nil
	})
}

func (cl *Clean) clean(c *cli.Context) error {
	fmt.Println("clean")
	_id, err := primitive.ObjectIDFromHex("61d55de7d157ed602e0021c2")
	if err != nil {
		logging.Warn(fmt.Sprintf("查询失败 err=%v \n", err))
		return err
	}
	cl.del(bson.M{"_id": _id})
	return nil
}

func (cl *Clean) getTable() *mongo.Collection {
	if cl.table != nil {
		return cl.table
	}
	client := db.MgoConnect(cl.dns)
	var err error
	if err != nil {
		panic("链接数据库有误!")
	}
	return client.Database("img").Collection("img")
}

// 删除操作
func (cl *Clean) del(m bson.M) {
	deleteResult, err := cl.getTable().DeleteMany(context.Background(), m)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("collection.DeleteOne:", deleteResult)
}

// 标记操作
func (cl *Clean) mk(ids []string, name string) {
	// fmt.Println(ids)
	if len(ids) == 0 {
		return
	}
	var oids []primitive.ObjectID
	for _, v := range ids {
		tmp_id, _ := primitive.ObjectIDFromHex(v)
		oids = append(oids, tmp_id)
	}
	filter := bson.M{"_id": bson.M{"$in": oids}}
	data := bson.M{"$set": bson.M{"use": -1, "cate": name}}
	result, err := cl.getTable().UpdateMany(context.Background(), filter, data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("处理条数%v ", result.ModifiedCount)

	filter = bson.M{"pid": bson.M{"$in": ids}}
	result, err = cl.getTable().UpdateMany(context.Background(), filter, data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf(" 处理thumb条数%v\n", result.ModifiedCount)
}

func init() {
	Commands = append(Commands, &cli.Command{
		Name:  "clean",
		Usage: "img数据库清理",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "mark", Usage: "标记有使用的图片"},
			&cli.BoolFlag{Name: "clean", Usage: "清理未标记的数据"},
			&cli.StringFlag{Name: "dns", Usage: "需要清理的img库", Required: true},
		},
		Action: (&Clean{}).Run,
	})
}
