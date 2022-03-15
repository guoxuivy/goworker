package store

import (
	. "cxe/db"
)

// 库房表数据模型  ActiveRecord实现
type Store struct {
	*Mgo
	IEntity
}

// 自定义模型标准构造函数
func NewStore() *Store {
	model := &Store{&Mgo{}, &Entity{}}
	model.Mgo.Model = model
	return model
}

func (s *Store) NewEntity() IEntity {
	return &Entity{}
}

func (s *Store) GetEntity() IEntity {
	return s.IEntity
}

func (s *Store) SetEntity(entity IEntity) {
	s.IEntity = entity
}
