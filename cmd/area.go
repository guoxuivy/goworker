package cmd

import (
	"context"
	"cxe/db"

	"errors"
	"strconv"
	"time"

	"cxe/helper"
	"cxe/util/logging"

	//"database/sql"
	"fmt"

	"encoding/json"

	"github.com/urfave/cli/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	TableName   = "area_trend"
	DailyTable  = "area_daily"
	MonthTable  = "area_monthly"
	DayFormat   = "20060102"
	MonthFormat = "200601"
)

//modify_log是否已分表
//const SPLIT_LOG = false
var IsSplit bool

//是否凌晨运行
//const RUN_MORNING = true
//1统计到今天，0到昨天
var RunMode int

var ctx context.Context

type JsonEntity struct {
	Id         string `bson:"_id,omitempty"` //omitempty 忽略掉空字符串
	Code       string `bson:"code"`
	Title      string `bson:"title"`
	CreateDate int32  `bson:"create_date"`
	//FloorTree           []*floorTree `bson:"floor_tree,omitempty" json:"-"`
	BusinessType        int32       `bson:"business_type" json:"business_type"`
	LandBusinessType    int32       `bson:"land_business_type" json:"land_business_type"`
	WareType            interface{} `bson:"ware_type" json:"ware_type"`
	UsableArea          interface{} `bson:"usable_area" json:"usable_area"`
	SaleableArea        interface{} `bson:"saleable_area" json:"saleable_area"`
	SurfaceRent         float64     `bson:"surface_rent" json:"surface_rent"`
	SurfaceRentCurrency int32       `bson:"surface_rent_currency" json:"surface_rent_currency"`
	SellingPrice        float64     `bson:"selling_price" json:"selling_price"`
	KeyWords            string      `bson:"key_words" json:"key_words"`
	ParkId              string      `bson:"park_id" json:"park_id"`
}

/*
type floorTree struct {
	BuildingId        string      `bson:"building_id"`
	FloorId           string      `bson:"floor_id"`
	SurfaceRent       interface{} `bson:"surface_rent"`        //最低租金
	UpsetSurfaceArea  interface{} `bson:"upset_surface_area"`  //起租面积
	UsableArea        interface{} `bson:"usable_area"`         //出租面积
	SellingPrice      interface{} `bson:"selling_price"`       //售价
	SaleableArea      interface{} `bson:"saleable_area"`       //可售面积
	UpsetSellingPrice interface{} `bson:"upset_selling_price"` //最低售价
	UpsetSellingArea  interface{} `bson:"upset_selling_area"`  // 起售面积
}*/

func (self *JsonEntity) GetWareType() []int32 {
	var res []int32
	//fmt.Printf("GetWareType:%# \n", self.WareType)
	if arr, isArr := self.WareType.([]interface{}); isArr {
		res = toInt32Array(arr)
	} else {
		res = append(res, toInt32(self.WareType))
	}
	//fmt.Println(res)
	return res
}

// 是否是土地
func (this *JsonEntity) CheckLand() bool {
	for _, wt := range this.GetWareType() {
		if wt == 4 {
			return true
		}
	}
	return false
}
func toFloat64(v interface{}) float64 {
	if v1, ok1 := v.(float64); ok1 {
		return v1
	}
	if v2, ok2 := v.(float32); ok2 {
		return float64(v2)
	}
	if v3, ok3 := v.(string); ok3 {
		vv3, err4 := strconv.ParseFloat(v3, 64)
		if err4 == nil {
			return vv3
		}
	}
	if v4, ok4 := v.(int); ok4 {
		return float64(v4)
	}
	return 0
}
func toInt32(v interface{}) int32 {
	if f1, isF := v.(float64); isF {
		return int32(f1)
	}
	if f2, isF2 := v.(float32); isF2 {
		return int32(f2)
	}
	if i1, isI := v.(int64); isI {
		return int32(i1)
	}
	if s1, isS := v.(string); isS {
		ss1, _ := strconv.Atoi(s1)
		return int32(ss1)
	}
	return 0
}
func toInt32Array(arr []interface{}) []int32 {
	var res []int32
	for _, v := range arr {
		res = append(res, toInt32(v))
	}
	return res
}
func toString(v interface{}) string {
	if str, ok := v.(string); ok {
		return str
	}
	return ""
}

// 每日数据统计
type AreaDaily struct {
	StoreId         string  `bson:"store_id"`
	ParkId          string  `bson:"park_id"`
	UsableArea      float64 `bson:"usable_area"`
	NewUsableArea   float64 `bson:"new_usable_area"`
	NewSaleableArea float64 `bson:"new_saleable_area"`
	WareType        int32   `bson:"ware_type"`
	BusinessType    int32   `bson:"business_type"`
	Day             string  `bson:"day"`
}

// 每月数据统计
type AreaMonthly struct {
	StoreId         string  `bson:"store_id"`
	ParkId          string  `bson:"park_id"`
	UsableArea      float64 `bson:"usable_area"`
	NewUsableArea   float64 `bson:"new_usable_area"`
	NewSaleableArea float64 `bson:"new_saleable_area"`
	WareType        int32   `bson:"ware_type"`
	BusinessType    int32   `bson:"business_type"`
	MonthDay        string  `bson:"month_day"`
}

//modify_log以及数据清洗
type ModifyLog struct {
	Id           int32      `bson:"-"`
	StoreId      string     `json:"store_id" bson:"store_id"`
	ParkId       string     `json:"park_id" bson:"park_id"`
	UsableArea   float64    `json:"usable_area" bson:"usable_area"`
	SaleableArea float64    `json:"saleable_area" bson:"saleable_area"`
	WareType     int32      `json:"ware_type" bson:"ware_type"`
	BusinessType int32      `json:"business_type" bson:"business_type"`
	Created      int32      `json:"create_date" bson:"create_date"`
	NewData      JsonEntity `bson:"-"`
	OldData      JsonEntity `bson:"-"`
	keyword      string
	Day          string `bson:"day"`
}

func (m *ModifyLog) parseData() {
	//fmt.Println(m.NewData.WareType)
	//fmt.Println(m.NewData.BusinessType)
	m.Day = time.Unix(int64(m.Created), 0).Format("20060102")
	store_waretype := m.NewData.GetWareType()
	if len(store_waretype) > 0 {
		m.WareType = store_waretype[0]
	} else {
		m.WareType = 0
	}
	if m.NewData.CheckLand() {
		m.BusinessType = m.NewData.LandBusinessType
		if m.BusinessType == 2 {
			m.SaleableArea = toFloat64(m.NewData.SaleableArea)
		} else {
			m.UsableArea = toFloat64(m.NewData.UsableArea)
		}
	} else {
		m.BusinessType = m.NewData.BusinessType
		if m.BusinessType == 2 {
			m.SaleableArea = toFloat64(m.NewData.SaleableArea)
		} else {
			m.UsableArea = toFloat64(m.NewData.UsableArea)
		}
	}
}
func (m *ModifyLog) getArea() float64 {
	if m.BusinessType == 2 {
		return m.SaleableArea
	} else {
		return m.UsableArea
	}
}

// fetch latest datetime from db
func fetchDate(collectionName string, pctx context.Context) (time.Time, error) {
	filter := bson.M{}
	opt := options.FindOne()
	opt.Sort = bson.D{{"create_date", -1}}

	ctx, cancel := context.WithTimeout(pctx, 5*time.Second)
	defer cancel()
	collection := db.DB.GetCollection(collectionName)

	var res ModifyLog
	err := collection.FindOne(ctx, filter, opt).Decode(&res)
	fmt.Println(res.Created)

	if err != nil {
		return time.Now(), err
	}
	if res.Created == 0 {
		return time.Now(), errors.New("Not Found")
	}
	return time.Unix(int64(res.Created), 0), nil
	/*if err == nil && res.Created > 0 {
		return time.Unix(int64(res.Created), 0), nil
	} else {
		//todo: get the earliest date from db
		row := db.DB.Db.QueryRow("SELECT min(create_date) as min_create_date FROM ysyc_modify_log")
		var t int
		err2 := row.Scan(&t)
		if err2 != nil {
			return time.Now(), err2
		}
		return time.Unix(int64(t), 0), nil
	}*/
}

func getInitModifyDate() (time.Time, error) {
	row := db.DB.Db.QueryRow("SELECT min(create_date) as min_create_date FROM ysyc_modify_log WHERE create_date IS NOT NULL")
	var t int64
	err2 := row.Scan(&t)
	return time.Unix(t, 0), err2
}

func getInitDailyDate() (time.Time, error) {
	filter := bson.M{}
	opt := options.FindOne()
	opt.Sort = bson.D{{"day", -1}}

	ctx := context.TODO()
	//defer cancel()
	collection := db.DB.GetCollection(DailyTable)

	var res AreaDaily
	err := collection.FindOne(ctx, filter, opt).Decode(&res)

	if err == nil {
		tm, err1 := time.Parse(DayFormat, res.Day)
		if err1 == nil {
			return tm.AddDate(0, 0, 1), nil
		}
	}
	return getInitModifyDate()
}

func getInitMonthlyDate() (time.Time, error) {
	filter := bson.M{}
	opt := options.FindOne()
	opt.Sort = bson.D{{"month_day", -1}}

	ctx := context.TODO()
	//defer cancel()
	collection := db.DB.GetCollection(MonthTable)

	var res AreaMonthly
	err := collection.FindOne(ctx, filter, opt).Decode(&res)

	if err == nil {
		tm, err1 := time.Parse(DayFormat, res.MonthDay)
		if err1 == nil {
			return tm, nil
		}
	}
	return getInitModifyDate()
}

//todo: push to queue
func pushData(d []ModifyLog) error {
	if len(d) == 0 {
		return errors.New("No data")
	}
	//ch <- json.Marshal(d)
	opt := options.InsertMany().SetOrdered(false)
	collection := db.DB.GetCollection(TableName)
	ctx := context.Background()
	var data []interface{}
	for _, el := range d {
		a, er := bson.Marshal(el)
		if er == nil {
			data = append(data, a)
		}
	}
	_, err := collection.InsertMany(ctx, data, opt)
	/*if err == nil {
		fmt.Println(r.InsertedIDs)
	}*/

	return err
}

func getModifyLogsByRel(store JsonEntity) ([]ModifyLog, error) {
	var sql string
	if IsSplit == true {
		sql = "SELECT id,rel_id,park_id,create_date,modify_date,old_data,new_data FROM ysyc_modify_log m JOIN ysyc_modify_content mc ON mc.modify_id=m.id WHERE rel_id=? AND audit_status in (-1,2) ORDER BY id ASC"
	} else {
		sql = "SELECT id,rel_id,park_id,create_date,modify_date,old_data,new_data FROM ysyc_modify_log WHERE rel_id=? AND audit_status in (-1,2) ORDER BY id ASC"
	}
	rows, err := db.DB.Db.Query(sql, store.Id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var lastArea float64
	lastArea = -1
	result := make([]ModifyLog, 0)
	for rows.Next() {
		var (
			mdata ModifyLog
			od    []byte
			nd    []byte
			cdate int32
			mdate int32
		)
		if er := rows.Scan(&mdata.Id, &mdata.StoreId, &mdata.ParkId, &cdate, &mdate, &od, &nd); er != nil {
			panic((er))
		}
		//nnd := make(map[string]interface{})
		if err1 := json.Unmarshal(nd, &mdata.NewData); err1 != nil {
			//fmt.Println(err1.Error())
			//fmt.Println(string(nd))
			//return result, err1
			continue
		}
		if err2 := json.Unmarshal(od, &mdata.OldData); err2 != nil {
			//fmt.Println(err2.Error())
			//fmt.Println(string(od))
			//panic(err2)
			//return result, err2
			continue
		}
		if mdate > 0 {
			mdata.Created = mdate
		} else {
			mdata.Created = cdate
		}
		mdata.keyword = store.KeyWords
		mdata.parseData()
		if !helper.Float64Equals(mdata.getArea(), lastArea) {
			lastArea = mdata.getArea()
			result = append(result, mdata)
		} else {
			//fmt.Println("skip ", mdata.getArea())
		}
	}
	return result, nil
}

func getStorePreviousArea(storeId string, time int32, btype int32) (ModifyLog, error) {
	opt := &options.FindOneOptions{
		Sort: bson.D{{"create_date", -1}}, //s.Order,
	}
	filter := bson.M{
		"business_type": btype,
		"store_id":      storeId,
		"create_date":   bson.D{{"$lt", time}},
	}
	collection := db.DB.GetCollection(TableName)
	//fmt.Println("before find")
	ctx := context.TODO()
	var result ModifyLog
	err := collection.FindOne(ctx, filter, opt).Decode(&result)
	if err != nil {
		fmt.Println("getStorePreviousArea:" + err.Error())
		//panic(err)
	}
	return result, err
}

func getModifyLogsByTimeRange(timeArr [2]time.Time) ([]ModifyLog, error) {
	var sql string
	if IsSplit == true {
		sql = `SELECT id,rel_id,park_id,create_date,modify_date,old_data,new_data
		FROM ysyc_modify_log m
		JOIN ysyc_modify_content mc ON mc.modify_id=m.id
		WHERE (create_date BETWEEN ? AND ?) AND type IN (3,7) AND audit_status in (-1,2) ORDER BY id ASC`
	} else {
		sql = `SELECT id,rel_id,park_id,create_date,modify_date,old_data,new_data
		FROM ysyc_modify_log
		WHERE (create_date BETWEEN ? AND ?) AND type IN (3,7) AND audit_status in (-1,2) ORDER BY id ASC`
	}

	rows, err := db.DB.Db.Query(sql, timeArr[0].Unix()+1, timeArr[1].Unix())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tmp := make(map[string]float64)
	//var lastArea float64
	//lastArea = 0
	result := make([]ModifyLog, 0)
	for rows.Next() {
		var (
			mdata ModifyLog
			od    []byte
			nd    []byte
			cdate int32
			mdate int32
		)
		if er := rows.Scan(&mdata.Id, &mdata.StoreId, &mdata.ParkId, &cdate, &mdate, &od, &nd); er != nil {
			panic((er))
		}
		//nnd := make(map[string]interface{})
		if err1 := json.Unmarshal(nd, &mdata.NewData); err1 != nil {
			//fmt.Println(err1.Error())
			//fmt.Println(string(nd))
			continue
		}
		if err2 := json.Unmarshal(od, &mdata.OldData); err2 != nil {
			//fmt.Println(err2.Error())
			//fmt.Println(string(od))
			continue
		}
		if mdate > 0 {
			mdata.Created = mdate
		} else {
			mdata.Created = cdate
		}
		//mdata.keyword = store.KeyWords
		mdata.parseData()
		if _, ok := tmp[mdata.StoreId]; ok == false {
			tmp[mdata.StoreId] = 0
		}
		if !helper.Float64Equals(mdata.getArea(), tmp[mdata.StoreId]) {
			tmp[mdata.StoreId] = mdata.getArea()
			//查找之前的记录
			prev, perr := getStorePreviousArea(mdata.StoreId, mdata.Created, mdata.BusinessType)
			if perr == nil && helper.Float64Equals(mdata.getArea(), prev.getArea()) {
				continue
			}
			result = append(result, mdata)
		} else {
			//fmt.Println("skip ", mdata.getArea())
		}
	}
	return result, nil
}

type StoreList struct {
	Order     map[string]int
	Page      int64
	PageSize  int64
	Done      bool
	Condition map[string]interface{}
}

func (s *StoreList) LoadData(pctx context.Context, ch chan<- JsonEntity) {
	filter := bson.M{
		"delete_status":    0,
		"is_undercarriage": 0,
		"is_finished":      1,
		"is_audited":       bson.D{{"$gt", 1}},
		"ware_type":        bson.D{{"$ne", 4}},
	}
	/*if len(s.Condition) > 0 {
		for k, v := range s.Condition {
			filter[k] = v
		}
	}*/

	s.PageSize = 100
	skip := (s.Page - 1) * s.PageSize
	//fmt.Printf("skip %d limit %d\n", skip, s.PageSize)
	opt := &options.FindOptions{
		Limit: &s.PageSize,
		Skip:  &skip,
		Sort:  bson.D{{"create_date", 1}}, //s.Order,
	}

	//ctx, cancel := context.WithTimeout(pctx, 5*time.Second)
	ctx, cancel := context.WithCancel(pctx)
	defer cancel()
	collection := db.DB.GetCollection("store")
	//fmt.Println("before find")
	cur, err := collection.Find(ctx, filter, opt)
	if err != nil {
		panic(err)
		//return result
	}
	rowsNum := cur.RemainingBatchLength()
	//fmt.Printf("after find %d \n", rowsNum)

	//subctx, subconcel := context.WithCancel(ctx)
	//defer subconcel()
	for cur.Next(ctx) {
		//res := store.Entity{}
		var res JsonEntity
		err := cur.Decode(&res)
		if err != nil {
			panic(err)
			//return result
		}
		//fmt.Println("id=", res.Id)
		ch <- res

	}
	var end JsonEntity
	if rowsNum == 0 {
		s.Done = true
		ch <- end
	}
}

func (s *StoreList) GetAllStore(ctx context.Context) <-chan JsonEntity {
	ch := make(chan JsonEntity)
	//fill data
	go func() {
		for s.Page = 1; s.Done != true; s.Page = s.Page + 1 {
			//fmt.Printf("load data Loop: %v\n", s.Done)
			s.LoadData(ctx, ch)
		}
		//fmt.Printf("isDone=%v\n", s.Done)
	}()
	return ch
}

func saveDailyRecord(d bson.M) error {
	ctx := context.TODO()
	opts := options.FindOne()

	id2, isM := d["_id"].(bson.M)
	if !isM {
		return errors.New("_id is not bson.M")
	}

	filter := bson.D{
		{"store_id", id2["store_id"]},
		{"business_type", id2["btype"]},
		{"day", d["day"]},
	}
	//fmt.Println(filter)
	collection := db.DB.GetCollection(TableName)
	var record bson.M
	err := collection.FindOne(ctx, filter, opts).Decode(&record)
	if err != nil {
		return err
	}

	store_id, isStr := record["store_id"].(string)
	if !isStr {
		return errors.New("store_id is not String")
	}
	oid, oidErr := primitive.ObjectIDFromHex(store_id)
	if oidErr != nil {
		return oidErr
	}

	coll := db.DB.GetCollection(DailyTable)
	usable_area := toFloat64(record["usable_area"])
	saleable_area := toFloat64(record["saleable_area"])
	// 获取每日增量
	t, terr := time.Parse(DayFormat, toString(d["day"]))
	if terr != nil {
		return terr
	}
	opt1 := options.FindOne()
	filter1 := bson.D{
		{"store_id", store_id},
		{"business_type", record["business_type"]},
		{"day", t.AddDate(0, 0, -1).Format(DayFormat)},
	}
	var old bson.M
	err1 := coll.FindOne(ctx, filter1, opt1).Decode(&old)
	delta1 := 0.0
	delta2 := 0.0
	if err1 != nil {
		delta1 = usable_area - toFloat64(old["usable_area"])
		delta2 = saleable_area - toFloat64(old["saleable_area"])
	}

	data := bson.M{
		"store_id":            store_id,
		"ref_store_id":        oid,
		"business_type":       record["business_type"],
		"ware_type":           record["ware_type"],
		"usable_area":         usable_area,
		"saleable_area":       saleable_area,
		"day":                 d["new_day"],
		"usable_area_delta":   delta1,
		"saleable_area_delta": delta2,
	}

	opt2 := options.InsertOne()
	_, err2 := coll.InsertOne(ctx, data, opt2)
	return err2
}

//日志 =》 每日数据
func getDailyLogData(startTime time.Time, endTime time.Time) <-chan bson.M {
	ch := make(chan bson.M)
	ctx := context.TODO()
	group := bson.D{
		{"$group", bson.D{
			{"_id", bson.D{
				{"store_id", "$store_id"},
				{"btype", "$business_type"},
			}},
			{"day", bson.D{
				{"$max", "$day"},
			}},
		}},
	}
	opts := options.Aggregate()
	collection := db.DB.GetCollection(TableName)
	initTime := TidyStartDate(startTime)
	finishTime := TidyEndDate(endTime)
	go func() {
		if initTime.After(finishTime) {
			ch <- bson.M{"_id": nil}
		} else {
			for tm := initTime; tm.Before(finishTime); tm = tm.AddDate(0, 0, 1) {
				day := tm.Format(DayFormat)
				fmt.Println(day)
				is_last := tm.AddDate(0, 0, 1).Add(time.Duration(time.Second * 60)).After(endTime)
				match := bson.D{
					{"$match", bson.D{
						{"day", bson.D{{"$lte", day}}},
					}},
				}

				cur, err := collection.Aggregate(ctx, mongo.Pipeline{match, group}, opts)
				if err != nil {
					fmt.Println(err.Error())
					logging.Error("collection.Aggregate: " + err.Error())
					//panic(err)
				} else {
					for cur.Next(ctx) {
						var res bson.M
						err1 := cur.Decode(&res)
						if err1 == nil {
							res["new_day"] = tm.Format(DayFormat)
							//fmt.Println(res)
							ch <- res
						} else {
							fmt.Println(err1.Error())
							logging.Info("" + err1.Error())
						}
					}
				}

				if is_last {
					ch <- bson.M{"_id": nil}
					break
				}
			}
		}

	}()
	return ch
}

func saveMonthlyRecord(d bson.M) error {
	ctx := context.TODO()
	opts := options.FindOne()

	id2, isM := d["_id"].(bson.M)
	if !isM {
		return errors.New("_id is not bson.M")
	}

	filter := bson.D{
		{"store_id", id2["store_id"]},
		{"business_type", id2["btype"]},
		{"day", d["day"]},
	}
	//fmt.Println(filter)
	collection := db.DB.GetCollection(TableName)
	var record bson.M
	err := collection.FindOne(ctx, filter, opts).Decode(&record)
	if err != nil {
		return err
	}

	store_id, isStr := record["store_id"].(string)
	if !isStr {
		return errors.New("store_id is not String")
	}
	oid, oidErr := primitive.ObjectIDFromHex(store_id)
	if oidErr != nil {
		return oidErr
	}

	coll := db.DB.GetCollection(MonthTable)
	usable_area := toFloat64(record["usable_area"])
	saleable_area := toFloat64(record["saleable_area"])

	data := bson.M{
		"store_id":      store_id,
		"ref_store_id":  oid,
		"business_type": record["business_type"],
		"ware_type":     record["ware_type"],
		"usable_area":   usable_area,
		"saleable_area": saleable_area,
		"month_day":     d["month_day"],
	}

	opt2 := options.InsertOne()
	_, err2 := coll.InsertOne(ctx, data, opt2)
	return err2
}

//日志 =》每月数据
func getMonthLogData(startTime time.Time, endTime time.Time) <-chan bson.M {
	ch := make(chan bson.M)
	ctx := context.TODO()
	group := bson.D{
		{"$group", bson.D{
			{"_id", bson.D{
				{"store_id", "$store_id"},
				{"btype", "$business_type"},
			}},
			{"day", bson.D{
				{"$max", "$day"},
			}},
		}},
	}
	opts := options.Aggregate()
	collection := db.DB.GetCollection(TableName)
	initTime := TidyStartDate(startTime)
	finishTime := TidyEndDate(endTime)
	//fmt.Printf("dayrange: [%s, %s]\n", initTime.Format(time.RFC3339Nano), finishTime.Format(time.RFC3339Nano))
	go func() {
		if initTime.After(finishTime) {
			fmt.Println("initTime > finishTime, Exit.")
			ch <- bson.M{"_id": nil}
		} else {
			for tm := initTime; tm.Before(finishTime); tm = tm.AddDate(0, 1, 0) {
				day := tm.Format(DayFormat)
				fmt.Println(day)
				is_last := tm.AddDate(0, 1, 0).Add(time.Duration(time.Hour * 24)).After(finishTime)
				match := bson.D{
					{"$match", bson.D{
						{"day", bson.D{{"$lt", day}}},
					}},
				}
				//fmt.Printf("day <= %s\n", day)
				cur, err := collection.Aggregate(ctx, mongo.Pipeline{match, group}, opts)
				if err != nil {
					fmt.Println(err.Error())
					logging.Error("collection.Aggregate: " + err.Error())
					//panic(err)
				} else {
					for cur.Next(ctx) {
						var res bson.M
						err1 := cur.Decode(&res)
						if err1 == nil {
							res["month_day"] = day
							//fmt.Println(res)
							ch <- res
						} else {
							fmt.Println(err1.Error())
							logging.Info("" + err1.Error())
						}
					}
				}

				if is_last {
					ch <- bson.M{"_id": nil}
					break
				}
			}
		}

	}()
	return ch
}

//把日期中的时间转换成 00:00:00
func TidyStartDate(t time.Time) time.Time {
	diff := time.Second * time.Duration((t.Hour()*3600+t.Minute()*60+t.Second())*-1)
	return t.Add(diff)
}

//把日期中的时间转换成 23:59:59
func TidyEndDate(t time.Time) time.Time {
	diff := time.Second * time.Duration(86400-(t.Hour()*3600+t.Minute()*60+t.Second())-1)
	return t.Add(diff)
}

//清洗日志数据
func runLog(ctx context.Context) error {
	latestDate, err1 := fetchDate(TableName, ctx)

	storeResult := &StoreList{
		Order:     map[string]int{"create_date": -1},
		Page:      0,
		PageSize:  50,
		Done:      false,
		Condition: make(map[string]interface{}),
	}
	if err1 == nil {
		//storeResult.Condition["create_date"] = bson.D{{"$gt", latestDate.Stamp}}
		timeEnd := time.Now()
		for tm := latestDate; tm.Before(timeEnd); tm = tm.AddDate(0, 0, 1) {
			//fmt.Println(tm.String())
			//hour, minute, second := tm.Clock()
			//t1 := tm.Add(time.Second * time.Duration(-1*(second+(minute*60)+(hour*3600))))
			t2 := tm.AddDate(0, 0, 1)
			//fmt.Println(tm)
			//fmt.Println("start datetime:", tm.String())
			//fmt.Println("end datetime:", t2.String())
			//return nil
			t := [2]time.Time{tm, t2}
			logs, err2 := getModifyLogsByTimeRange(t)
			if err2 != nil {
				fmt.Println(err2.Error())
				panic(err2)
			}
			if len(logs) > 0 {
				rs1 := pushData(logs)
				if rs1 != nil {
					fmt.Println(rs1.Error())
					panic(rs1)
				}
			}

		}
	} else {
		for store := range storeResult.GetAllStore(ctx) {
			if store.Id == "" {
				fmt.Println("Finished")
				break
			}
			//fmt.Println("ID=%#v", store.Id)
			logs, err := getModifyLogsByRel(store)
			if err != nil {
				//panic(err)
				//return err
				fmt.Println(store.Id)
				fmt.Println(err.Error())
			}
			if len(logs) > 0 {
				//fmt.Printf("get logs %d \n", len(logs))
				rs := pushData(logs)
				if rs != nil {
					panic(rs)
				}
			}

		}
	}

	return nil
}

//每日数据
func runDaily(ctx context.Context) error {
	startTime, err1 := getInitDailyDate()
	if err1 != nil {
		fmt.Println(err1.Error())
		panic(err1)
	}
	fmt.Println("get init date: ", startTime.Format(DayFormat))
	//now := time.Now()
	var endTime time.Time
	if RunMode == 1 {
		//today
		endTime = time.Now()
	} else {
		//tomorrow
		endTime = time.Now().AddDate(0, 0, -1)
	}
	if startTime.After(endTime) {
		return nil
	}
	fmt.Printf("from %s to %s \n", startTime.Format(DayFormat), endTime.Format(DayFormat))
	for d := range getDailyLogData(startTime, endTime) {
		if d["_id"] == nil {
			return nil
		} else {
			//todo: save to mongodb
			err := saveDailyRecord(d)
			if err != nil {
				fmt.Println(err.Error())
				return err
			}
		}
	}
	return nil
}

//转换成当前月份1日
func TindyMonth(t time.Time) time.Time {
	d := t.Day()
	return t.AddDate(0, 0, -1*(d-1))
}

//每月数据
func runMonthly(ctx context.Context) error {
	startTime, err1 := getInitMonthlyDate()
	if err1 != nil {
		fmt.Println(err1.Error())
		panic(err1)
	}
	//从后一个月开始
	initMonthDate := TindyMonth(startTime.AddDate(0, 1, 0))

	fmt.Println("get init month: ", initMonthDate.Format(DayFormat))
	//截止月份
	endTime := TindyMonth(time.Now())

	if startTime.After(endTime) {
		return nil
	}
	fmt.Printf("month from %s to %s \n", initMonthDate.Format(DayFormat), endTime.Format(DayFormat))
	for d := range getMonthLogData(initMonthDate, endTime) {
		//fmt.Printf("%v \n", d)
		logging.Info(fmt.Sprintf("getMonthLogData:%v\n", d))
		if d["_id"] == nil {
			return nil
		} else {
			//todo: save to mongodb
			err := saveMonthlyRecord(d)
			if err != nil {
				fmt.Println(err.Error())
				return err
			}
		}
	}
	return nil
}

func DropCollection(ctx context.Context, collectionName string) {
	db.DB.GetCollection(collectionName).Drop(ctx)
}

func run(c *cli.Context) error {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	if c.Bool("split") {
		IsSplit = true
	} else {
		IsSplit = false
	}
	if c.Int("mode") == 1 {
		RunMode = 1
	} else {
		RunMode = 0
	}
	is_remove := c.Bool("remove")
	if c.Bool("logs") {
		if is_remove {
			DropCollection(ctx, TableName)
		}
		return runLog(ctx)
	}
	if c.Bool("daily") {
		if is_remove {
			//fmt.Println("remove old data")
			DropCollection(ctx, DailyTable)
		}
		return runDaily(ctx)
	}
	if c.Bool(("monthly")) {
		if is_remove {
			//fmt.Println("remove old data")
			DropCollection(ctx, MonthTable)
		}
		return runMonthly(ctx)
	}
	return nil

}
func init() {
	Commands = append(Commands, &cli.Command{
		Name:  "area",
		Usage: "处理库房面积变动",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "daily",
				Usage: "生成每日数据",
			},
			&cli.BoolFlag{
				Name:  "logs",
				Usage: "清洗modify_log数据",
			},
			&cli.BoolFlag{
				Name:  "monthly",
				Usage: "生成每月一日数据",
			},
			&cli.BoolFlag{
				Name:  "remove",
				Usage: "删除旧数据",
			},
			&cli.BoolFlag{
				Name:  "split",
				Usage: "使用已分割的modify_log",
			},
			&cli.IntFlag{
				Name:  "mode",
				Usage: "数据统计模式",
			},
		},
		Action: run,
	})
}
