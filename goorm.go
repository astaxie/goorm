package goorm

import (
	"errors"
	"fmt"
	"github.com/mikespook/mymysql/mysql"
	_ "github.com/mikespook/mymysql/native"
	"reflect"
	"strconv"
	"strings"
)

type ORM struct {
	Db mysql.Conn
}

/**
 * 新注册一个ORM对象
 */
func NewORM(dbhost, dbport, dbname, dbuser, dbpass, dbcharset string) ORM {
	link := fmt.Sprintf("%v:%v", dbhost, dbport)
	db := mysql.New("tcp", "", link, dbuser, dbpass, dbname)
	err := db.Connect()
	if err != nil {
		panic(err)
	}
	charsetsql := fmt.Sprintf("set Names %v", dbcharset)
	db.Query(charsetsql)
	return ORM{db}
}

//更换数据库
func (orm *ORM) SelectDb(dbname string) error {
	orm.Db.Use(dbname)
	return nil
}

//查询数据返回map列表数据
func (orm *ORM) getResultsForQuery(columnNames, tableName, condition string, args []interface{}) (resultsSlice []map[string][]byte, err error) {
	a := fmt.Sprintf("select %v from %v %v", columnNames, tableName, condition)
	fmt.Println(a)
	s, err := orm.Db.Prepare(fmt.Sprintf("select %v from %v %v", columnNames, tableName, condition))
	if err != nil {
		return nil, err
	}
	rows, res, err := s.Exec(args...)
	fields := res.Fields()
	if err != nil {
		panic(err)
	}
	for _, row := range rows {
		result := make(map[string][]byte)
		i := 0
		for _, key := range fields {
			if row[i] == nil {
				result[key.Name] = []byte("")
			} else {
				switch key.Type {
				case 1, 2, 3, 7, 8, 9, 16:
					result[key.Name] = []byte(strconv.Itoa(row.Int(i)))
				case 4, 5, 246:
					result[key.Name] = []byte(strconv.FormatFloat(row.Float(i), 'f', -1, 64))
				case 254, 253, 252:
					result[key.Name] = row[i].([]byte)
				case 10, 11, 12, 13:
					result[key.Name] = row[i].([]byte)
				}
			}
			i++
		}
		resultsSlice = append(resultsSlice, result)
	}
	return
}

//插入或者更新，如果结构中id字段不为0就更新，等于0插入数据
func (orm *ORM) Save(rowStruct interface{}) error {
	results, _ := scanStructIntoMap(rowStruct)
	tableName := getTableName(rowStruct)

	id := results["id"]
	delete(results, "id")

	if id == 0 {
		id, err := orm.insert(tableName, results)
		if err != nil {
			return nil
		}

		structPtr := reflect.ValueOf(rowStruct)
		structVal := structPtr.Elem()
		structField := structVal.FieldByName("Id")
		structField.Set(reflect.ValueOf(id))

		return nil
	} else {
		condition := fmt.Sprintf("id=%v", id)
		_, err := orm.Update(tableName, results, condition)
		if err != nil {
			return err
		}
	}
	return nil
}

//插入数据
func (orm *ORM) insert(tableName string, properties map[string]interface{}) (uint64, error) {
	var keys []string
	var placeholders []string
	var args []interface{}

	for key, val := range properties {
		keys = append(keys, key)
		placeholders = append(placeholders, "?")
		args = append(args, val)
	}

	statement := fmt.Sprintf("insert into %v (%v) values (%v)",
		tableName,
		strings.Join(keys, ", "),
		strings.Join(placeholders, ", "))

	stmt, err := orm.Db.Prepare(statement)
	if err != nil {
		return 0, err
	}
	defer stmt.Delete()
	res, err := stmt.Run(args)
	if err != nil {
		return 0, err
	}
	defer res.End()
	id := res.InsertId()

	return id, nil
}

// 更新数据
func (orm *ORM) Update(tableName string, properties map[string]interface{}, condition string) (uint64, error) {
	var updates []string
	var args []interface{}

	for key, val := range properties {
		updates = append(updates, fmt.Sprintf("`%v` = ?", key))
		args = append(args, val)
	}

	statement := fmt.Sprintf("update `%v` set %v where %v",
		tableName,
		strings.Join(updates, ", "),
		condition)
	stmt, err := orm.Db.Prepare(statement)
	if err != nil {
		return 0, err
	}
	defer stmt.Delete()
	res, err := stmt.Run(args)
	if err != nil {
		return 0, err
	}
	defer res.End()
	rows := res.AffectedRows()
	return rows, nil
}

//获取一条数据
func (orm *ORM) Get(rowStruct interface{}, condition interface{}, args ...interface{}) error {
	var keys []string
	results, _ := scanStructIntoMap(rowStruct)
	tableName := getTableName(rowStruct)
	conditionStr := ""

	switch condition := condition.(type) {
	case string:
		conditionStr = condition
	case int:
		conditionStr = "id = ?"
		args = append(args, condition)
	}

	conditionStr = fmt.Sprintf("where %v limit 1", conditionStr)

	for key, _ := range results {
		keys = append(keys, key)
	}

	resultsSlice, err := orm.getResultsForQuery(strings.Join(keys, ", "), tableName, conditionStr, args)

	if err != nil {
		return err
	}

	switch len(resultsSlice) {
	case 0:
		return errors.New("did not find any results")
	case 1:
		results := resultsSlice[0]
		scanMapIntoStruct(rowStruct, results)
	default:
		return errors.New("more than one row matched")
	}

	return nil
}

//获取多条数据
func (orm *ORM) GetAll(rowsSlicePtr interface{}, condition string, args ...interface{}) error {
	sliceValue := reflect.Indirect(reflect.ValueOf(rowsSlicePtr))
	if sliceValue.Kind() != reflect.Slice {
		return errors.New("needs a pointer to a slice")
	}

	sliceElementType := sliceValue.Type().Elem()
	st := reflect.New(sliceElementType)
	var keys []string
	results, _ := scanStructIntoMap(st.Interface())
	tableName := getTableName(rowsSlicePtr)
	for key, _ := range results {
		keys = append(keys, key)
	}

	condition = strings.TrimSpace(condition)
	if len(condition) > 0 {
		condition = fmt.Sprintf("where %v", condition)
	}

	resultsSlice, err := orm.getResultsForQuery(strings.Join(keys, ", "), tableName, condition, args)
	if err != nil {
		return err
	}

	for _, results := range resultsSlice {
		newValue := reflect.New(sliceElementType)
		scanMapIntoStruct(newValue.Interface(), results)
		fmt.Println(newValue.Interface())
		sliceValue.Set(reflect.Append(sliceValue, reflect.Indirect(reflect.ValueOf(newValue.Interface()))))
	}

	return nil
}
