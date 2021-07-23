package main
import (
	"fmt"
	"os"

	"dbx/database"
	//ms "dbx/mapstruct"
)

var (
	db *database.Db
	tb = "demo"
)

func main() {
	dbPath := "db.sqlite3"
	if len(os.Args) > 1 {
		dbPath = os.Args[1]
	}
	if err := Connect(dbPath); err != nil {
		fmt.Printf("Connect to %s: %s\n", dbPath, err.Error())
		return
	}
	err := CreateTb(tb, false)
	if err != nil {
		fmt.Printf("Create Table %s: %s\n", tb, err.Error())
		return
	}
	fmt.Printf("Done\n")
}

// 每一個 api 應該自行負責用到的資料表格，用法:
//   if err := api.CreateTb("termgrp", false); err != nil { ... }
func Connect(path string) (err error) {
	db,err = database.Connect(path)
	if err != nil {
		return fmt.Errorf("Fatal: connect to database: %s\n", err.Error())
	}
	return nil
}

///////////////////////////////////////////////////////////////////////////
// 底下是對 database 的包裝，
// 主要是需要有個 db 物件來記住，省得每次呼叫 database api 都要建立 db 物件
///////////////////////////////////////////////////////////////////////////

func CreateTb(tb string, dropFirst bool) (err error) {
	return db.CreateTb(tb, dropFirst)
}

func Insert(tb string, input interface{}) (map[string]interface{}, error) {
 	return db.Insert(tb, input)
}

func Update(tb string, input interface{}) error {
	return db.Update(tb, input)
}

func InsOrEdit(tb string, input interface{}) map[string]interface{} {
	return db.InsOrEdit(tb, input)
}

func InsIfNotExist(tb string, input interface{}) map[string]interface{} {
	return db.InsIfNotExist(tb, input)
}

func Get(tb string, id int) map[string]interface{} {
	return db.Get(tb, id)
}

func Gets(tb string) []map[string]interface{} {
	return db.Gets(tb)
}

func GetsByFilter(tb, filter string) []map[string]interface{} {
	return db.GetsByFilter(tb, filter)
}

func Del(tb string, id int) error {
	return db.Del(tb, id)
}

func DelsBy(tb, field string, min, max int) error {
	return db.DelsBy(tb, field, min, max)
}

func MapInsert(tb string, input map[string]interface{}) (map[string]interface{}, error) {
	return db.MapInsert(tb, input)
}

func MapAryInsert(tb string, data []map[string]string, check bool) error {
	return db.MapAryInsert(tb, data, check)
}

func MapUpdate(tb string, input map[string]interface{}) error {
	return db.MapUpdate(tb, input)
}

func MapInsOrEdit(tb string, input map[string]interface{}) map[string]interface{} {
	return db.MapInsOrEdit(tb, input)
}

func MapInsIfNotExist(tb string, input map[string]interface{}) map[string]interface{} {
	return db.MapInsIfNotExist(tb, input)
}

func MapGetId(input map[string]interface{}) int {
	return db.MapGetId(input)
}

func MaxId(tb string) int {
	return db.MaxId(tb)
}

func NextId(tb string) int {
	return db.NextId(tb)
}
