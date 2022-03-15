package db

import (
	"cxe/util"
	"cxe/util/logging"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

/*
查询使用示例：
NewModel(ysyc_agent).
	Table("ysyc_agent t").
	Field("real_name,sex").
	Where("age=1").
	Where("agent_role", 1).
	Where("time", ">=", 10).
	Where("time", "<=", 33).
	WhereOr("time", 88).
	Join("task a ON t.id=a.id", "left").
	Join("good b ON t.id=b.id").
	Group("age").
	Limit(2).
	Page(1,10).
	Paginate(3, 3)
	Paginate(3, 3)
	Find()
	FindAll()
*/

// 统一 *sql.Tx 和  *sql.DB对象 事务处理
// SQLCommon is the minimal database connection functionality gorm requires.  Implemented by *sql.DB.
type SQLCommon interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

// SQL表达式
// protected $selectSql    = 'SELECT%DISTINCT% %FIELD% FROM %TABLE%%FORCE%%JOIN%%WHERE%%GROUP%%HAVING%%ORDER%%LIMIT% %UNION%%LOCK%%COMMENT%';
// protected $insertSql    = '%INSERT% INTO %TABLE% (%FIELD%) VALUES (%DATA%) %COMMENT%';
// protected $insertAllSql = '%INSERT% INTO %TABLE% (%FIELD%) %DATA% %COMMENT%';
// protected $updateSql    = 'UPDATE %TABLE% SET %SET% %JOIN% %WHERE% %ORDER%%LIMIT% %LOCK%%COMMENT%';
// protected $deleteSql    = 'DELETE FROM %TABLE% %USING% %JOIN% %WHERE% %ORDER%%LIMIT% %LOCK%%COMMENT%';

// 快速查询封装 无实体
type Mysql struct {
	table       string
	options     map[string]interface{}
	whereValues []interface{}
}

// 无实体实例化mysql模型
func NewModel(table string) *Mysql {
	model := &Mysql{}
	model.WithTable(table)
	return model
}

// 模型初始化时调用
func (model *Mysql) WithTable(table string) *Mysql {
	model.table = table
	return model
}

func (model *Mysql) FindAll() (RowRecords, error) {
	sql := model.getSelectSql()
	list, err := model.Query(sql)
	model.clean()
	return list, err
}
func (model *Mysql) Find() (RowRecord, error) {
	sql := model.Limit(1).getSelectSql()
	list, err := model.Query(sql, true)
	model.clean()
	if len(list) == 1 {
		return list[0], nil
	}
	return nil, err
}

// 清理查询现场
func (model *Mysql) clean() {
	model.whereValues = nil
	model.options = nil
}

// 最多接受一个参数，多的丢弃  Count("1")
func (model *Mysql) Count(args ...string) (int64, error) {
	f := "*"
	if len(args) > 0 {
		f = args[0]
	}
	sql := model.Field("COUNT(" + f + ") as cxe_count").Limit(1).Order("").getSelectSql()
	list, err := model.Query(sql, true)
	model.clean()
	if len(list) == 1 {
		c, _ := strconv.ParseInt(util.Strval(list[0]["cxe_count"]), 10, 64)
		return c, nil
	}
	return 0, err
}

// 分页查询
func (model *Mysql) Paginate(page int64, size int64) (result *PageRecords, err error) {
	result = &PageRecords{Page: page, Size: size, Total: 0, LastPage: 1, List: make(RowRecords, 0)}
	f := model.getOption("FIELD")
	l := model.getOption("LIMIT")
	o := model.getOption("ORDER")
	countSql := model.Field("COUNT(*) as cxe_count").Limit(1).Order("").getSelectSql()
	list, err := model.Query(countSql, true)
	if err != nil {
		return result, err
	}
	var total int64
	if len(list) == 1 {
		total, err = strconv.ParseInt(util.Strval(list[0]["cxe_count"]), 10, 64)
		if err != nil {
			return result, err
		}
	}
	if total == 0 {
		return result, nil
	}
	model.setOption("FIELD", f).setOption("LIMIT", l).setOption("ORDER", o)
	sql := model.Page(page, size).getSelectSql()
	list, err = model.Query(sql)
	if err != nil {
		return result, err
	}
	result.Total = total
	result.List = list
	result.LastPage = int64(math.Ceil(float64(total) / float64(size)))

	model.clean()
	return result, nil
}

func (model *Mysql) getSelectSql() string {
	sql := "SELECT %FIELD% FROM %TABLE% %JOIN%%WHERE%%GROUP%%HAVING%%ORDER% %LIMIT% %UNION%"
	sql = strings.Replace(sql, "%FIELD%", model.getOption("FIELD"), 1)
	sql = strings.Replace(sql, "%TABLE%", model.getOption("TABLE"), 1)
	sql = strings.Replace(sql, "%JOIN%", model.getOption("JOIN"), 1)
	sql = strings.Replace(sql, "%WHERE%", model.getOption("WHERE"), 1)
	sql = strings.Replace(sql, "%GROUP%", model.getOption("GROUP"), 1)
	sql = strings.Replace(sql, "%HAVING%", model.getOption("HAVING"), 1)
	sql = strings.Replace(sql, "%ORDER%", model.getOption("ORDER"), 1)
	sql = strings.Replace(sql, "%LIMIT%", model.getOption("LIMIT"), 1)
	sql = strings.Replace(sql, "%UNION%", model.getOption("UNION"), 1)
	return sql
}
func (model *Mysql) getOption(key string) string {
	if v, ok := model.options[key]; ok {
		return util.Strval(v)
	}
	val := ""
	switch key {
	case "FIELD":
		val = "*"
	case "TABLE":
		val = model.table
	}
	return val
}
func (model *Mysql) setOption(key string, v interface{}) *Mysql {
	if model.options == nil {
		model.options = make(map[string]interface{})
	}
	model.options[key] = v
	return model
}
func (model *Mysql) Field(str string) *Mysql {
	return model.setOption("FIELD", str)
}
func (model *Mysql) Group(str string) *Mysql {
	return model.setOption("GROUP", " GROUP BY "+str)
}
func (model *Mysql) Having(str string) *Mysql {
	return model.setOption("HAVING", " HAVING "+str)
}
func (model *Mysql) Union(str string) *Mysql {
	return model.setOption("UNION", " UNION "+str)
}
func (model *Mysql) Order(str string) *Mysql {
	if str == "" {
		return model.setOption("ORDER", "")
	} else {
		return model.setOption("ORDER", " ORDER BY "+str)
	}
}

func (model *Mysql) Where(where ...interface{}) *Mysql {
	str := model.prepareWhere(where...)
	w := model.getOption("WHERE")
	if w != "" {
		return model.setOption("WHERE", model.getOption("WHERE")+" AND "+str)
	}
	return model.setOption("WHERE", " WHERE "+str)
}

func (model *Mysql) WhereOr(where ...interface{}) *Mysql {
	str := model.prepareWhere(where...)
	w := model.getOption("WHERE")
	if w != "" {
		return model.setOption("WHERE", model.getOption("WHERE")+" OR "+str)
	}
	return model.setOption("WHERE", " WHERE "+str)
}

// 支持 1、2、3个参数
func (model *Mysql) prepareWhere(where ...interface{}) string {
	whereStr := []string{}
	switch len(where) {
	case 1:
		whereStr = append(whereStr, util.Strval(where[0]))
	case 2:
		whereStr = append(whereStr, util.Strval(where[0]))
		whereStr = append(whereStr, "=")
		whereStr = append(whereStr, model.prepareWhereValue(where[1]))
	case 3:
		whereStr = append(whereStr, util.Strval(where[0]))
		whereStr = append(whereStr, util.Strval(where[1]))
		whereStr = append(whereStr, model.prepareWhereValue(where[2]))
	}
	return strings.Join(whereStr, "")
}

func (model *Mysql) prepareWhereValue(value interface{}) string {
	model.whereValues = append(model.whereValues, value)
	return "?"
}

// array?
func (model *Mysql) Join(join ...string) *Mysql {
	joinStr := ""
	switch len(join) {
	case 1:
		joinStr = "LEFT JOIN " + join[0]
	case 2:
		joinStr = strings.ToUpper(join[1]) + " JOIN " + join[0]
	default:
		joinStr = ""
	}

	j := model.getOption("JOIN")
	if j != "" {
		return model.setOption("JOIN", j+" "+joinStr)
	}
	return model.setOption("JOIN", joinStr)
}

func (model *Mysql) Table(tableName string) *Mysql {
	return model.setOption("TABLE", tableName)
}
func (model *Mysql) Page(page int64, size int64) *Mysql {
	offset := (page - 1) * size
	return model.Limit(offset, size)
}

// 接受1~2个参数
func (model *Mysql) Limit(args ...int64) *Mysql {
	var offset, size int64
	offset = 0
	switch len(args) {
	case 0:
		size = 1
	case 1:
		size = args[0]
	case 2:
		offset = args[0]
		size = args[1]
	default:
		size = 1
	}
	return model.setOption("LIMIT", fmt.Sprintf("LIMIT %v,%v", offset, size))
}

// 查询 返回字符串mapList集
func (model *Mysql) Query(sqlstr string, one ...bool) (list RowRecords, err error) {
	// Execute the query
	logging.Sql(sqlstr, model.whereValues)
	rows, err := model.link().Query(sqlstr, model.whereValues...)
	if err != nil {
		logging.Error(err.Error())
		return nil, err
	}
	defer rows.Close()
	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		logging.Error(err.Error())
		return nil, err
	}
	// Make a slice for the values
	values := make([]sql.RawBytes, len(columns))
	// rows.Scan wants '[]interface{}' as an argument, so we must copy the
	// references into such a slice
	// See http://code.google.com/p/go-wiki/wiki/InterfaceSlice for details
	scanArgs := make([]interface{}, len(columns))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	list = make(RowRecords, 0)
	// Fetch rows
	for rows.Next() {
		// get RawBytes from data
		rows.Scan(scanArgs...)
		// Now do something with the data.
		// Here we just print each column as a string.
		// var value string
		tmp := make(RowRecord)
		for i, col := range values {
			// Here we can check if the value is nil (NULL value)
			// if col == nil {
			// 	value = "NULL"
			// } else {
			// 	value = string(col)
			// }
			tmp[columns[i]] = string(col) //value
		}
		list = append(list, tmp)
		if len(one) > 0 && one[0] {
			// 单条查询
			break
		}
	}
	if err = rows.Err(); err != nil {
		logging.Error(err.Error())
	}
	return list, nil
}

// 返回影响行数or插入id
func (model *Mysql) Exec(sqlstr string) (int64, error) {
	logging.Sql(sqlstr)
	stm, err := model.link().Prepare(sqlstr)
	if err != nil {
		logging.Error(err.Error())
		return 0, err
	}
	defer stm.Close()
	result, err := stm.Exec()
	if err != nil {
		logging.Error(err.Error())
		return 0, err
	}

	if find := strings.Contains(strings.ToLower(sqlstr), "insert"); find {
		return result.LastInsertId()
	}
	return result.RowsAffected()
}

// 获取连接【适配事务】
func (model *Mysql) link() SQLCommon {
	if DB.Tx != nil {
		return DB.Tx
	} else {
		return DB.Db
	}
}

// 已经支持事务嵌套 共用同一个全局Tx对象 【非线程安全,后期可考虑用sync.map管理】
// 如果内层事务回滚了，外层事务依次回滚
func (model *Mysql) Transaction(fc func(txModel *Mysql) error) (err error) {
	if DB.Tx == nil {
		tx, _ := DB.Db.Begin()
		DB.Tx = &TX{tx, 1, nil}
		logging.Debug("Transaction Begin")
	} else {
		DB.Tx.TransactionLevel++
		logging.Debug("Begin :TransactionLevel", DB.Tx.TransactionLevel)
	}

	// 创建一个持有tx的新模型
	txModel := NewModel(model.table)
	panicked := true
	defer func() {
		// Make sure to rollback when panic, Block error or Commit error
		if panicked {
			DB.Tx.Err = errors.New("unknow panic")
			// 不捕获异常，留给调用方自主处理
			// if err := recover(); err != nil {
			// 	logging.Debug("Transaction Panic", err)
			// 	DB.Tx.Err = errors.New(fmt.Sprint(err))
			// }
		}
		// 当前闭包内无异常和错误
		if panicked || err != nil {
			DB.Tx.TransactionLevel--
			logging.Debug("Rollback :TransactionLevel", DB.Tx.TransactionLevel)
			if DB.Tx.TransactionLevel == 0 {
				DB.Tx.Rollback()
				logging.Debug("Transaction Rollback")
				DB.Tx = nil
				logging.Debug("Transaction clean")
			}
		}
	}()
	// 执行当前闭包
	err = fc(txModel)
	if err == nil {
		DB.Tx.TransactionLevel--
		logging.Debug("Commit :TransactionLevel", DB.Tx.TransactionLevel)
		if DB.Tx.TransactionLevel == 0 {
			if DB.Tx.Err == nil {
				err = DB.Tx.Commit()
				logging.Debug("Transaction Commit")
			} else {
				err = DB.Tx.Rollback()
				logging.Debug("Transaction Rollback")
			}
			DB.Tx = nil
			logging.Debug("Transaction clean")
		}
	}
	if err != nil {
		DB.Tx.Err = err
	}
	panicked = false
	return
}

func MysqlConnect(dsn string) *sql.DB {
	link, err := sql.Open("mysql", dsn)
	if err != nil {
		logging.Error("数据库连接出错:%s\n", err.Error())
		panic(err)
	}
	// 设置连接池中空闲连接的最大数量。
	link.SetMaxIdleConns(10)
	// 设置打开数据库连接的最大数量。
	link.SetMaxOpenConns(5)
	// 设置连接可复用的最大时间。
	link.SetConnMaxLifetime(time.Second * 30)
	//设置连接最大空闲时间
	link.SetConnMaxIdleTime(time.Second * 30)

	//检查连通 性
	err = link.Ping()
	if err != nil {
		logging.Error("数据库连接出错:%s\n", err.Error())
		panic(err)
	}
	return link
}
