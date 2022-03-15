package agent

import (
	. "cxe/db"
)

// 库房表数据模型  ActiveRecord实现
type Agent struct {
	*Mysql
}

// 自定义模型标准构造函数
func NewAgent() *Agent {
	return &Agent{NewModel("ysyc_agent")}
}
