package db

import (
	"context"
	"cxe/util/logging"
	"errors"
	"fmt"
	"net/url"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IMgo interface {
	Find(interface{}, *options.FindOneOptions) error
	FindByID(string) error
	FindAll(interface{}, *options.FindOptions) ([]IEntity, error)
	Update(interface{}, IEntity) (int64, error)
	Save() error
	Chunk(filter interface{}, sort interface{}, limit int64, fun func(page int64, list []IEntity) error)
}

// 提供基础封装
type Mgo struct {
	// 当前持有模型 依赖注入
	Model IModel
}

var databaseName string

func (mgo *Mgo) Table() *mongo.Collection {
	e := mgo.Model.GetEntity()
	if e == nil {
		e = mgo.Model.NewEntity()
	}
	return DB.Mongo.Database(databaseName).Collection(e.GetTable())
}

// @title    查询单条记录
// @param    filter             "查询条件"
// @return
func (mgo *Mgo) Find(filter interface{}, ops *options.FindOneOptions) error {
	// filter := bson.M{"code": "KF211027005"}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res := mgo.Model.NewEntity()
	if ops == nil {
		ops = options.FindOne()
	}
	err := mgo.Table().FindOne(ctx, filter, ops).Decode(res)
	if err != nil {
		logging.Warn(fmt.Sprintf("查询失败 err=%v \n", err))
		return err
	}
	mgo.Model.SetEntity(res)
	return nil
}

// @title    id查询数据
func (mgo *Mgo) FindByID(id string) error {
	_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		logging.Warn(fmt.Sprintf("查询失败 err=%v \n", err))
		return err
	}
	filter := bson.M{"_id": _id}
	return mgo.Find(filter, nil)
}

func (mgo *Mgo) FindAll(filter interface{}, ops *options.FindOptions) ([]IEntity, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cur, err := mgo.Table().Find(ctx, filter, ops)
	if err != nil {
		logging.Warn(fmt.Sprintf("查询失败 err=%v \n", err))
		return nil, err
	}
	var list []IEntity
	for cur.Next(context.Background()) {
		result := mgo.Model.NewEntity()
		err := cur.Decode(result)
		if err != nil {
			logging.Warn(fmt.Sprintf("解析失败 err=%v \n", err))
		}
		list = append(list, result)
	}
	return list, nil
}

// @title    多记录更新操作
// @param    filter             "查询条件"
// @param    entity             "需要更新的字段"
// @return
func (mgo *Mgo) Update(filter interface{}, entity IEntity) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// 排除id
	entity.UnsetId()
	data := bson.M{"$set": entity}
	res, err := mgo.Table().UpdateMany(ctx, filter, data)
	if err != nil {
		logging.Warn(fmt.Sprintf("保存失败 err=%v \n", err))
		return 0, err
	}
	return res.MatchedCount, nil
}

// 保存or插入
func (mgo *Mgo) Save() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// 排除id
	entity := mgo.Model.GetEntity()
	if entity == nil {
		logging.Warn("保存失败 空对象")
		return errors.New("保存失败 空对象")
	}
	id := entity.GetId()
	if id == "" {
		// 插入
		res, err := mgo.Table().InsertOne(ctx, entity)
		if err != nil {
			logging.Warn("保存失败 err=", err)
			return fmt.Errorf("保存失败 err=%v", err)
		}
		entity.SetId(res.InsertedID.(primitive.ObjectID).Hex())
		return nil
	} else {
		// 更新
		entity.UnsetId()
		data := bson.M{"$set": entity}
		_id, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			logging.Warn("保存失败id err=", err)
			return fmt.Errorf("保存失败id err=%v", err)
		}
		_, err = mgo.Table().UpdateByID(ctx, _id, data)
		if err != nil {
			logging.Warn("保存失败 err=", err)
			return fmt.Errorf("保存失败 err=%v", err)
		}
		return nil
	}
}

// 分片处理返回值
func (mgo *Mgo) Chunk(filter interface{}, sort interface{}, limit int64, fun func(int64, []IEntity) error) {
	var index int64 = 0
	for {
		ops := options.Find().SetLimit(limit).SetSkip(limit * index).SetSort(sort)
		list, err := mgo.FindAll(filter, ops)
		if err != nil {
			logging.Warn("Chunk失败 err=", err)
			break
		}
		if len(list) == 0 {
			break
		}
		index++
		err = fun(index, list)
		if err != nil {
			break
		}
	}
}

func MgoConnect(dsn string) *mongo.Client {
	clientOptions := options.Client().ApplyURI(dsn)
	// 获取连接数据库名称
	u, err := url.Parse(dsn)
	if err != nil {
		panic(err)
	}
	if u.Path != "" {
		rs := []rune(u.Path)
		databaseName = string(rs[1:])
	}

	// clientOptions := options.Client().ApplyURI("mongodb://192.168.1.205:27017/cxe")
	clientOptions.SetMaxPoolSize(30)                  // 连接池大小
	clientOptions.SetMaxConnIdleTime(5 * time.Minute) // 指定空闲链接最大生存时间
	clientOptions.SetConnectTimeout(10 * time.Minute)

	// Connect to MongoDB

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		logging.Error(err)
	}
	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		logging.Error(err)
	}
	logging.Info("Connected to MongoDB!")
	return client
}
