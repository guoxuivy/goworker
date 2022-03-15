package park

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
)

type Entity struct {
	Id                string   `bson:"_id,omitempty"` //omitempty 忽略掉空字符串
	Code              string   `bson:"code"`
	OverallViewPicurl string   `bson:"overall_view_picurl"`
	LocationPicurl    string   `bson:"location_picurl"`
	InteriorPicurl    string   `bson:"interior_picurl"`
	PeripheryPicurl   string   `bson:"periphery_picurl"`
	OtherPicurl       []string `bson:"other_picurl"`
}

// "overall_view_picurl", //仓库全景图'
// "location_picurl", //仓库外景图'
// "interior_picurl", //仓库内景图'
// "periphery_picurl", //周边环境'
// "other_picurl",   //其他图片id[]

// 必须实现IEntity接口
func (this *Entity) GetTable() string {
	return "park"
}

// 必须实现IEntity接口
func (this *Entity) UnsetId() {
	this.Id = ""
}

// 必须实现IEntity接口
func (this *Entity) GetId() string {
	return this.Id
}

// 必须实现IEntity接口
func (this *Entity) SetId(id string) {
	this.Id = id
}

func (this *Entity) toBsonBytes() []byte {
	/* 结构体序列化成bson格式[]byte */
	data, err := bson.Marshal(this)
	if nil != err {
		fmt.Println("序列化Bson失败")
		return nil
	}
	return data
}
