package db

type IModel interface {
	NewEntity() IEntity
	GetEntity() IEntity
	SetEntity(IEntity)
}
type IEntity interface {
	// 获取表明
	GetTable() string
	// 更新时排除id
	UnsetId()
	GetId() string
	SetId(string)
}

//RowRecord 基于Map定义RowRecord类例 //转换为model?
type RowRecord map[string]interface{}

//RowRecords 批量记录
type RowRecords []RowRecord

type PageRecords struct {
	List RowRecords
	// 总条数
	Total int64
	// 当前页码
	Page int64
	// 总页码
	LastPage int64
	// 每页条数
	Size int64
}
