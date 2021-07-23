package database

// 這邊負責將 map[string]interface{} 存到資料表，主要是給 json 使用

import (
	"fmt"
	"reflect"
	"strings"
)

func (db *Db) MapInsert(tb string, input map[string]interface{}) (map[string]interface{}, error) {
	objId := db.NextId(tb)

	// 要知道的是，input 每個Key:Value，對表格來說都是一筆資料
	sql := "INSERT INTO " + tb + " (ObjId,Attr,Val,Typ) VALUES "
	// 透過 reflect.TypeOf().Field(i) 可以 traverse 每個欄位
    for k, v := range input {
		typ := fmt.Sprintf("%v", reflect.TypeOf(v))
		if typ == "float64" || typ == "int64" || typ == "json.Number" {
			sql = fmt.Sprintf(`%s (%d,"%s","%v","%s"),`, sql,
				objId, k, v, "int")
		} else {
			sql = fmt.Sprintf(`%s (%d,"%s","%v","%s"),`, sql,
				objId, k, v, typ)
		}
    }
	sql = strings.TrimRight(sql, ",")
	sql = sql + ";"
	db.Db.Exec(sql)
	data := db.Get(tb, objId)
	return data, nil
}

// 給 termcap 專用，用有效率的方式一次性插入一堆 []map[string]string
func (db *Db) MapAryInsert(tb string, data []map[string]string, check bool) error {
	objId := db.MaxId(tb)

	tx := db.Db.MustBegin()

	sql := ""
	for _,input := range data {
		objId = objId + 1
		// 要知道的是，input 每個Key:Value，對表格來說都是一筆資料
		sql = "INSERT INTO " + tb + " (ObjId,Attr,Val,Typ) VALUES "
		// 透過 reflect.TypeOf().Field(i) 可以 traverse 每個欄位
	    for k, v := range input {
			typ := fmt.Sprintf("%v", reflect.TypeOf(v))
			if typ == "float64" || typ == "int64" || typ == "json.Number" {
				sql = fmt.Sprintf(`%s (%d,"%s","%v","%s"),`, sql,
					objId, k, v, "int")
			} else {
				sql = fmt.Sprintf(`%s (%d,"%s","%v","%s"),`, sql,
					objId, k, v, typ)
			}
	    }
		sql = strings.TrimRight(sql, ",")
		sql = sql + ";"
		db.Db.MustExec(sql)
	}
	tx.Commit()
	return nil
}

func (db *Db) MapUpdate(tb string, input map[string]interface{}) error {
	objId := db.MapGetId(input)
	if objId <= 0 {
		return fmt.Errorf("Cannot Update table without Id field")
	}

	// 透過 reflect.TypeOf().Field(i) 可以 traverse 每個欄位
	sql := ""
	// 一直找不到適合的 IF EXIST UPDATE ELSE INSERT 語句，只好分兩段，先查，再判斷
    for k, v := range input {
		val := ""
		sql = fmt.Sprintf(`SELECT Val FROM %s WHERE objId=%d AND Attr="%s";`, tb, objId, k)
		err := db.Db.Get(&val, sql)
		exists := true
		if err != nil {
			exists = false
		}
		typ := fmt.Sprintf("%v", reflect.TypeOf(v))
		if typ == "json.Number" {	// json 的數字在轉換時很怪, 需要特別處理
			typ = "int"
		}
		if !exists { // 不存在
			sql = fmt.Sprintf(`INSERT INTO %s (ObjId,Attr,Val,Typ) VALUES
				(%d,"%s","%v","%v");`, tb, objId, k, v, typ)
		} else {
			sql = fmt.Sprintf(`UPDATE %s Set Val="%v" WHERE objId=%d AND Attr="%s";`,
				tb, v, objId, k)
		}
		_,err = db.Db.Exec(sql)
    }
	return nil
}

// 如果給的資料 Id == 0 || 不存在，則 Insert
// 如果 Id > 0 && 存在 則 Update
// PS: 存不存在由 Id 決定
func (db *Db) MapInsOrEdit(tb string, input map[string]interface{}) map[string]interface{} {
	if id := db.MapGetId(input); id > 0 { // id > 0 才有機會是 Update, 否則一律 Insert
		if item := db.Get(tb, id); len(item) == 0 { // Not existed
			item, err := db.MapInsert(tb, input)
			if err == nil {
				return item
			} else {
				return map[string]interface{}{}
			}
		} else {
			db.MapUpdate(tb, input)
			return db.Get(tb, id)
		}
	} else {
		item, err := db.MapInsert(tb, input)
		if err == nil {
			return item
		} else {
			return map[string]interface{}{}
		}
	}
}

// 如果 Id > 0 && 不存在才 Insert, 否則 Skip
// PS: 存不存在由 Id 決定
func (db *Db) MapInsIfNotExist(tb string, input map[string]interface{}) map[string]interface{} {
	if id := db.MapGetId(input); id > 0 { // id > 0 才有機會找出資料項
		if item := db.Get(tb, id); len(item) == 0 { // Not existed
			item, err := db.MapInsert(tb, input)
			if err == nil {
				return item
			} else {
				return map[string]interface{}{}
			}
		}
	}
	return map[string]interface{}{}
}

// 一般 Id 都不會是 <= 0
func (db *Db) MapGetId(input map[string]interface{}) int {
	if _,ok := input["Id"]; ok {
		return input["Id"].(int) // int(input["Id"].(float64))
	}
	return -1
}
