package store

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
)

type Entity struct {
	Id                  string       `bson:"_id,omitempty"` //omitempty 忽略掉空字符串
	Code                string       `bson:"code"`
	OverallViewPicurl   string       `bson:"overall_view_picurl"`
	LocationPicurl      string       `bson:"location_picurl"`
	InteriorPicurl      string       `bson:"interior_picurl"`
	PeripheryPicurl     string       `bson:"periphery_picurl"`
	OtherPicurl         []string     `bson:"other_picurl"`
	Title               string       `bson:"title"`
	CreateDate          int32        `bson:"create_date"`
	FloorTree           []*floorTree `bson:"floor_tree,omitempty"`
	BusinessType        int32        `bson:"business_type" json:"business_type"`
	LandBusinessType    int32        `bson:"land_business_type" json:"land_business_type"`
	WareType            []int32      `bson:"ware_type" json:"ware_type"`
	UsableArea          float64      `bson:"usable_area" json:"usable_area"`
	SaleableArea        float64      `bson:"saleable_area" json:"saleable_area"`
	SurfaceRent         float64      `bson:"surface_rent" json:"surface_rent"`
	SurfaceRentCurrency int32        `bson:"surface_rent_currency" json:"surface_rent_currency"`
	SellingPrice        float64      `bson:"selling_price" json:"selling_price"`
	KeyWords            string       `bson:"key_words" json:"key_words"`
	ParkId              string       `bson:"park_id" json:"park_id"`
}

type floorTree struct {
	BuildingId        string  `bson:"building_id"`
	FloorId           string  `bson:"floor_id"`
	SurfaceRent       float64 `bson:"surface_rent"`        //最低租金
	UpsetSurfaceArea  float64 `bson:"upset_surface_area"`  //起租面积
	UsableArea        float64 `bson:"usable_area"`         //出租面积
	SellingPrice      float64 `bson:"selling_price"`       //售价
	SaleableArea      float64 `bson:"saleable_area"`       //可售面积
	UpsetSellingPrice float64 `bson:"upset_selling_price"` //最低售价
	UpsetSellingArea  float64 `bson:"upset_selling_area"`  // 起售面积
}

// 必须实现IEntity接口
func (this *Entity) GetTable() string {
	return "store"
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

// 是否是土地
func (this *Entity) CheckLand() bool {
	for _, wt := range this.WareType {
		if wt == 4 {
			return true
		}
	}
	return false
}
