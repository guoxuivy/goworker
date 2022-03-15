package park

import (
	. "cxe/db"
)

// 库房表数据模型  ActiveRecord实现
type Park struct {
	*Mgo
	IEntity
}

// 自定义模型标准构造函数
func NewPark() *Park {
	model := &Park{&Mgo{}, &Entity{}}
	model.Mgo.Model = model
	return model
}

func (s *Park) NewEntity() IEntity {
	return &Entity{}
}

func (s *Park) GetEntity() IEntity {
	return s.IEntity
}

func (s *Park) SetEntity(entity IEntity) {
	s.IEntity = entity
}
