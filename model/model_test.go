package model

import (
	"cxe/db"
	"cxe/model/agent"
	"testing"
)

// func Test_Save(t *testing.T) {
// 	var storeModel = store.NewStore()
// 	storeModel.FindByID("61791568a3d24b0a06189466")
// 	rand.Seed(time.Now().Unix()) // 防止rand不dfd变
// 	storeModel.GetEntity().(*store.Entity).Title = "测试修改title再次" + strconv.Itoa(rand.Intn(100)) + "次"
// 	storeModel.Save()
// 	storeModel.FindByID("61791568a3d24b0a06189466")
// 	t.Log(storeModel.GetEntity())
// }

// func Test_GetList(t *testing.T) {
// 	var storeModel = store.NewStore()
// 	filter := bson.M{"usable_area": bson.M{"$lt": 10}}
// 	// filter := bson.D{{"usable_area", 10}}
// 	// filter := bson.M{"usable_area": 10}
// 	ops := options.Find().SetLimit(10).SetSort(bson.D{{Key: "create_date", Value: -1}, {Key: "_id", Value: -1}})
// 	list, _ := storeModel.FindAll(filter, ops)
// 	for _, v := range list {
// 		t.Log(v.(*store.Entity).UsableArea)
// 	}
// }

// func Test_GetOne(t *testing.T) {
// 	var storeModel = store.NewStore()
// 	one := storeModel.Find(bson.M{"code": "KF211027005"}, nil)
// 	t.Log(one)
// }

// func Test_Mysql(t *testing.T) {
// 	model := agent.NewAgent()
// 	row, _ := model.
// 		Table("ysyc_agent t").
// 		Field("real_name,sex").
// 		// Where("name=guox").
// 		// Where("age=1").
// 		Where("agent_role", 1).
// 		// Where("time", ">=", 10).
// 		// Where("time", "<=", 33).
// 		// WhereOr("time", 88).
// 		// // Join("task a ON t.id=a.id", "left").
// 		// // Join("good b ON t.id=b.id").
// 		// Group("age").
// 		Paginate(3, 3)
// 	t.Log(row)
// 	c, _ := db.NewModel("ysyc_agent").
// 		Field("real_name,sex").
// 		// Where("name=guox").
// 		// Where("age=1").
// 		Where("agent_role", "2 or 1 = 1 -- and password='").
// 		// Where("time", ">=", 10).
// 		// Where("time", "<=", 33).
// 		// WhereOr("time", 88).
// 		// // Join("task a ON t.id=a.id", "left").
// 		// // Join("good b ON t.id=b.id").
// 		// Group("age").
// 		Count()
// 	t.Log(c)
// }

// mysql 事务测试

// mysql 嵌套事务测试
func Test_Mysql_Transaction(t *testing.T) {
	model := agent.NewAgent()

	model.Transaction(func(txModel *db.Mysql) error {

		defer func() {
			if err := recover(); err != nil {
				t.Log(err)
			}
		}()
		txModel.Exec("UPDATE `cxe`.`ysyc_agent` SET `real_name` = '笪敏1' WHERE `id` = '28563'")
		txModel.Transaction(func(txModel1 *db.Mysql) error {
			// t.Log(txModel)
			// panic("my panic")
			txModel1.Exec("UPDATE `cxe`.`ysyc_agent` SET `real_name` = '夏枫1' WHERE `id` = '29882'")

			panic("my panic")
			// return errors.New("内层回滚")
			return nil
		})
		return nil
		// return errors.New("外层回滚")
	})
	rows, _ := model.Where("id IN(28563,29882)").Field("real_name,sex").FindAll()
	t.Log(rows)
	// if row["real_name"] != "笪敏" {
	// 	t.Error("事务异常")
	// }
	model.Exec("UPDATE `cxe`.`ysyc_agent` SET `real_name` = '笪敏' WHERE `id` = '28563'")
	model.Exec("UPDATE `cxe`.`ysyc_agent` SET `real_name` = '夏枫' WHERE `id` = '29882'")
}
