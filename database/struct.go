package database

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func (db *Db) Insert(tb string, input interface{}) (map[string]interface{}, error) {
    getType := reflect.TypeOf(input)
    getValue := reflect.ValueOf(input)
	objId := db.NextId(tb)

	// 要知道的是，input 每個欄位，對表格來說都是一筆資料
	sql := "INSERT INTO " + tb + " (ObjId,Attr,Val,Typ) VALUES "
	// 透過 reflect.TypeOf().Field(i) 可以 traverse 每個欄位
    for i := 0; i < getType.NumField(); i++ {
        field := getType.Field(i)
        value := getValue.Field(i).Interface()
		sql = fmt.Sprintf(`%s (%d,"%s","%v","%s"),`, sql,
			objId, field.Name, value, field.Type.Name())
    }
	sql = strings.TrimRight(sql, ",")
	sql = sql + ";"
	db.Db.Exec(sql)
	data := db.Get(tb, objId)
	return data, nil
}

func (db *Db) Update(tb string, input interface{}) error {
    getType := reflect.TypeOf(input)
    getValue := reflect.ValueOf(input)

	objId := getId(input)
	if objId == 0 {
		return fmt.Errorf("Cannot Update table without Id field")
	}

	// 透過 reflect.TypeOf().Field(i) 可以 traverse 每個欄位
    for i := 0; i < getType.NumField(); i++ {
        field := getType.Field(i)
        value := getValue.Field(i).Interface()
		sql := ""

        val := ""
        sql = fmt.Sprintf(`SELECT Val FROM %s WHERE objId=%d AND Attr="%s";`, tb, objId, field.Name)
        err := db.Db.Get(&val, sql)
		typ := field.Type.Name()
        if err != nil { // !exists
            sql = fmt.Sprintf(`INSERT INTO %s (objId,Attr,Val,Typ) VALUES
                (%d,"%s","%v","%v");`, tb, objId, field.Name, value, typ)
        } else {
            sql = fmt.Sprintf(`UPDATE %s Set Val="%v" WHERE objId=%d AND Attr="%s";`,
                tb, value, objId, field.Name)
		}
		db.Db.Exec(sql)
    }
	return nil
}

// 如果給的資料 Id == 0 || 不存在，則 Insert
// 如果 Id > 0 && 存在 則 Update
// PS: 存不存在由 Id 決定
func (db *Db) InsOrEdit(tb string, input interface{}) map[string]interface{} {
	if id := getId(input); id > 0 { // id > 0 才有機會是 Update, 否則一律 Insert
		if item := db.Get(tb, id); len(item) == 0 { // Not existed
			item, err := db.Insert(tb, input)
			if err == nil {
				return item
			} else {
				return map[string]interface{}{}
			}
		} else {
			db.Update(tb, input)
			return db.Get(tb, id)
		}
	} else {
		item, err := db.Insert(tb, input)
		if err == nil {
			return item
		} else {
			return map[string]interface{}{}
		}
	}
}

// 如果 Id > 0 && 不存在才 Insert, 否則 Skip
// PS: 存不存在由 Id 決定
func (db *Db) InsIfNotExist(tb string, input interface{}) map[string]interface{} {
	if id := getId(input); id > 0 { // id > 0 才有機會找出資料項
		if item := db.Get(tb, id); len(item) == 0 { // Not existed
			item, err := db.Insert(tb, input)
			if err == nil {
				return item
			} else {
				return map[string]interface{}{}
			}
		}
	}
	return map[string]interface{}{}
}

func getId(input interface{}) int {
    getType := reflect.TypeOf(input)
    getValue := reflect.ValueOf(input)
    for i := 0; i < getType.NumField(); i++ {
        field := getType.Field(i)
        value := getValue.Field(i).Interface()
		if field.Name == "Id" {
			id,err := strconv.Atoi(fmt.Sprintf("%v", value))
			if err != nil {
				return 0
			}
			return id
		}
    }
	return 0
}
