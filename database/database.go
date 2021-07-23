package database

// 會用到 reflect, 儲存時將 struct 變成資料表，查詢時轉成 map[string]interface{}
// 關於 map[string]interface{} 再轉成 struct, 請見 utils.mapstuct 套件
//
// 會用到 subquery, 子查詢，就是 SELECT 中還有 SELECT, 
// 子查詢的用途有二，
//   1. 最基本的是將兩個表格的資料關聯起來，這種用法跟 JOIN 相似但有區別，不多說
//   2. 這邊因為是直式表格設計，用來過濾特定欄位範例，請見 GetsByFilter()
// 參考: [SQL 子查詢](https://www.1keydata.com/tw/sql/sql-subquery.html)

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"strings"
    "strconv"
    "time"

    t "dbx/time"
)

// 採用 struct 的方式，可以在 Db struct 放入更多屬性
type Db struct {
	Db *sqlx.DB
}

// 所有資料表都使用制式表格, 為一種直式表格，
// 每一欄屬性都是一筆資料，因此欄位可以無限擴充
type Table struct {
	Id		int		`db:"Id"`
	ObjId	int		`db:"ObjId"`	// 例如 TerminalId, AppId, RdsId 等等，屬於同一筆資料的關鍵字
	Attr	string	`db:"Attr"`
	Val		string	`db:"Val"`
	Typ		string	`db:"Typ"`
}

// Connect 連接到 MariaDB, 需要
// (u)sername, (p)assword, ip, port(=3306), (d)atabase
func Connect(path string) (db *Db, err error) {
	db = &Db{}
	db.Db, err = sqlx.Connect("sqlite3", path)
	return db, err
}

func (db *Db) String() string {
	return "sqlite3"
}

// dropFirst = true 時會先 drop table, 失敗則停止，
//             false 則會嚐試 create table, 已存在仍返回成功
func (db *Db) CreateTb(tb string, dropFirst bool) (err error) {
	if dropFirst {
		var dropSql = fmt.Sprintf("DROP TABLE IF EXISTS %s;", tb)
		_, err = db.Db.Exec(dropSql)
		if err != nil {
			return fmt.Errorf("%s\n\t%s", err.Error(), dropSql)
		}
	}

	var schema = fmt.Sprintf("CREATE TABLE %s ( Id INTEGER PRIMARY KEY AUTOINCREMENT, ObjId INTEGER DEFAULT 1, Attr TEXT DEFAULT 'UNKNOWN', Val TEXT DEFAULT 'UNKNOWN', Typ TEXT DEFAULT 'UNKNOWN', UNIQUE(ObjId,Attr));", tb)
	_, err = db.Db.Exec(schema)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("%s\n\t%s", err.Error(), schema)
		}
	}
	return nil
}

func (db *Db) MaxId(tb string) int {
	// 找出目前筆數，以防止在找 ObjId 時出錯
	count := 0
	sql := "SELECT COUNT(ObjId) from "+tb+";"
	if err := db.Db.Get(&count, sql); err != nil {
		return -1
	}

	// 找出最大的 ObjId
	objId := 0
	if count > 0 {
		sql = "SELECT MAX(ObjId) from "+tb+";"
		if err := db.Db.Get(&objId, sql); err != nil {
			return -1
		}
	}
	return objId
}

func (db *Db) NextId(tb string) int {
	return db.MaxId(tb) + 1
}

// 相當於 GetsByFilter(tb, "ObjId="+id)
// 當然這邊的特定函式效率高
func (db *Db) Get(tb string, id int) map[string]interface{} {
	data := []Table{}
	res := map[string]interface{}{}
	sql := fmt.Sprintf(`SELECT * FROM %s WHERE ObjId=%d;`, tb, id)
	err := db.Db.Select(&data, sql)
	if err != nil {
		fmt.Printf("database.Get(%d) %s\n\terr: %s\n", id, sql, err.Error())
		return res
	}

	if len(data) == 0 {
		return res
	}
	res["Id"] = data[0].ObjId
	for _,d := range data {
		switch d.Typ {
		case "int", "int64":
			v,_ := strconv.Atoi(d.Val)
			res[d.Attr] = v
		case "bool":
			res[d.Attr] = d.Val == "true"
		case "string":
			res[d.Attr] = d.Val
		case "Time", "time.Time":
			v := strings.ReplaceAll(d.Val, " +0000 UTC", "")
			tt,err := time.Parse("2006-01-02 03:04:05", v)
			if err != nil {
				fmt.Printf("Format: %s\n", err.Error())
			} else {
				res[d.Attr] = (t.Time)(tt)
			}
		default: // 需要處理別種型態，例如 nil
			//fmt.Printf("Get(%s,%d) catch wrong Typ '%v' with Val '%v'\n", tb, id, d.Typ, d.Val)
			res[d.Attr] = d.Val
		}
	}
	return res
}

// 只有 Gets() 有個 checkVer 參數，而 Get()/GetsBy() 兩個都沒有 checkVer，原因是
//  Gets() 有機會用在取得 system config 資料表，而
//  Get()/GetsBy() 只用來取得特定資料，通常非 system config 應用
func (db *Db) Gets(tb string) []map[string]interface{} {
	data := []Table{}
	sql := ""
	sql = fmt.Sprintf(`SELECT * FROM %s ORDER BY ObjId;`, tb)
	err := db.Db.Select(&data, sql)
	if err != nil {
		fmt.Printf("database.Gets() %s\n\terr: %s\n", sql, err.Error())
		return nil
	}

	res := []map[string]interface{}{}
	if len(data) == 0 {
		return res
	}
	r := map[string]interface{}{}
	oid := 0
	for _,d := range data {
		if d.ObjId != oid { // 新的資料
			if len(r) > 0 {
				res = append(res, r)
				r = map[string]interface{}{}
				oid = d.ObjId
			}
		}
		if oid == 0 {
			oid = d.ObjId
		}
		r["Id"] = d.ObjId
		switch d.Typ {
		case "int", "int64":
			v,_ := strconv.Atoi(d.Val)
			r[d.Attr] = v
		case "bool":
			r[d.Attr] = d.Val == "true"
		case "string":
			r[d.Attr] = d.Val
		case "Time":
			v := strings.ReplaceAll(d.Val, " +0000 UTC", "")
			tt,err := time.Parse("2006-01-02 03:04:05", v)
			if err != nil {
				fmt.Printf("Format: %s\n", err.Error())
			} else {
				r[d.Attr] = (t.Time)(tt)
			}
		default: // 需要處理別種型態，例如 nil
			// fmt.Printf("Get(%s,%d) catch wrong Typ '%v' with Val '%v'\n", tb, oid, d.Typ, d.Val)
			r[d.Attr] = d.Val
		}
	}
	res = append(res, r)
	return res
}

// 用來過濾欄位, 例如 IsGroup = true, 可以指定欄位的值, 如:
// GetsByFilter("term", `Attr="IsGroup" AND Val="false"`)
// 原先有設計 GetsByField(), 後來併入 GetsByFilter(), 
//   主要是因為後者比較有彈性，可以像 1<=ObjId AND ObjId<=10 AND Attr="..." 這樣的複式條件
// 注意: 多欄位要使用 subquery，單一的呼叫 GetsByFilter() 目前做不到，橫式比較好做
func (db *Db)GetsByFilter(tb, filter string) []map[string]interface{} {
	data := []Table{}
	sql := fmt.Sprintf(`SELECT * FROM %s WHERE ObjId IN (SELECT ObjId FROM %s WHERE %s) ORDER BY ObjId;`,
		tb, tb, filter)
	err := db.Db.Select(&data, sql)
	if err != nil {
		fmt.Printf("database.GetsByFilter() %s\n\terr: %s\n", sql, err.Error())
		return nil
	}
	res := []map[string]interface{}{}
	r := map[string]interface{}{}
	oid := 0
	for _,d := range data {
		if d.ObjId != oid { // 新的資料
			if len(r) > 0 { // 已經有 r 值
				res = append(res, r)
				r = map[string]interface{}{}
				oid = d.ObjId
			}
		}
		if oid == 0 {
			oid = d.ObjId
		}
		r["Id"] = d.ObjId
		switch d.Typ {
		case "int", "int64":
			v,_ := strconv.Atoi(d.Val)
			r[d.Attr] = v
		case "bool":
			r[d.Attr] = d.Val == "true"
		case "string":
			r[d.Attr] = d.Val
		case "Time":
			v := strings.ReplaceAll(d.Val, " +0000 UTC", "")
			tt,err := time.Parse("2006-01-02 03:04:05", v)
			if err != nil {
				fmt.Printf("Format: %s\n", err.Error())
			} else {
				r[d.Attr] = (t.Time)(tt)
				fmt.Printf("%v: %+v\n", d.Attr, r[d.Attr])
			}
		default: // 需要處理別種型態，例如 nil
			// fmt.Printf("Get(%s,%d) catch wrong Typ '%v' with Val '%v'\n", tb, oid, d.Typ, d.Val)
			r[d.Attr] = d.Val
		}
	}
	res = append(res, r)
	return res
}

// 相當於 DelsBy(tb, "Id", id, id)
// 當然這邊的特定用途的效率較高
func (db *Db) Del(tb string, id int) error {
	sql := fmt.Sprintf(`DELETE FROM %s WHERE ObjId=%d;`, tb, id)
	_,err := db.Db.Exec(sql)
	if err != nil {
		fmt.Printf("database.Del() %s\n\terr: %s\n", sql, err.Error())
		return err
	}

	return nil
}

func (db *Db) DelsBy(tb, field string, min, max int) error {
	sql := ""
	if field == "Id" || field == "ObjId" {
		sql = fmt.Sprintf(`DELETE FROM %s WHERE %d<=ObjId AND ObjId<=%d;`,
			tb, min, max)
	} else {
		sql = fmt.Sprintf(`DELETE FROM %s WHERE ObjId IN (SELECT ObjId FROM %s WHERE Attr="%s" AND %d<=Val AND Val<=%d);`,
		tb, tb, field, min, max)
	}
	_,err := db.Db.Exec(sql)
	if err != nil {
		fmt.Printf("database.DelsBy() %s\n\terr: %s\n", sql, err.Error())
		return err
	}

	return nil
}
