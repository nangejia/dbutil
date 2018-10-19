package dbutil

import (
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
	"os"
	"fmt"
	"bufio"
	"io"
	"strings"
	"log"
)

var db *sql.DB
//读取配置文件
func readIni() (username, password, host, database string, err error) {
	//读取配置文件
	file, error := os.Open("db.ini")
	if error != nil {
		err = error
		return
	}
	defer file.Close()
	//按行读取配置文件
	br := bufio.NewReader(file)

	for {
		//读取一行
		buf, err := br.ReadBytes('\n')
		//将读取的字节转成字符串，并去掉头部和尾部的空格
		line := strings.Trim(string(buf),"")
		//判断行首是否以#开头
		if strings.HasPrefix(line, "#") {
			continue
		}
		word := strings.Split(line, "=")
		key, value := word[0], strings.Trim(word[1], "\r\n")

		switch key {
		case "username":
			username = value
		case "password":
			password = value
		case "host":
			host = value
		case "database":
			database = value
		}
		if err != nil && err == io.EOF {
			break
		}
	}
	return
}

//获取数据库
func GetDB() (*sql.DB) {
	if db == nil {
		db = connDB()
	}
	return db
}

func connDB() (db *sql.DB) {
	username, password, host, database, err := readIni();
	if err != nil {
		fmt.Println("读取数据库配置文件失败:", err)
		return nil
	}
	dsn := username + ":" + password + "@tcp(" + host + ")/" + database
	//fmt.Println(dsn)
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("打开数据库%s失败\n", database)
		return nil
	}

	err = db.Ping()
	if err != nil {
		fmt.Printf("尝试连接数据库%s失败\n", database)
		return
	}
	fmt.Println("数据库连接......成功!")
	return
}

//关闭数据库
func CloseDB() {
	if db != nil {
		db.Close()
	}
}

//封装增、删、改、查
//查询返回一条数据
func QueryRow(sqlstr string, args ...interface{}) (rest []interface{}) {
	db := GetDB()
	if db == nil {
		fmt.Println("Query 失败")
		return
	}

	//生成SQL预编译
	stmt, _ := db.Prepare(sqlstr)
	rows, _ := stmt.Query(args...)
	if rows.Next() {
		//得到返回列数
		cols, _ := rows.Columns()
		length := len(cols)

		if length == 0 {
			rest = make([]interface{}, 1)
		} else {
			rest = make([]interface{}, length)
		}
		var dest []interface{} = make([]interface{}, len(rest))
		for i := 0; i < len(dest); i++ {
			dest[i] = &rest[i]
		}
		rows.Scan(dest...)
	}
	return
}

//查询多行多列数据
func Query(sqlstr string, args ...interface{}) (rest [][]interface{}) {
	db := GetDB()
	if db == nil {
		fmt.Println("Query 失败")
		return
	}

	//生成SQL预编译
	stmt, _ := db.Prepare(sqlstr)
	rows, _ := stmt.Query(args...)

	//得到返回列数
	cols, _ := rows.Columns()
	length := len(cols)

	//循环读取数据库数据
	for rows.Next() {
		//生成一行数据
		var dest []interface{}
		if length == 0 {
			dest = make([]interface{}, 1)
		} else {
			dest = make([]interface{}, length)
		}

		var ps []interface{} = make([]interface{}, len(dest))
		for i := 0; i < len(dest); i++ {
			ps[i] = &dest[i]
		}
		rows.Scan(ps...)
		//向结果集中追加一行数据
		rest = append(rest, dest)
	}
	return
}

//返回首行首列
func QueryUnique(sql string, args ...interface{}) (dest interface{}) {
	dest = QueryRow(sql, args...)[0]
	return
}

//返回统计
func QueryCount(sql string, args ...interface{}) (count int64) {
	result := QueryUnique(sql, args...)
	if v, ok := result.(int64); ok {
		count = v
	}
	return
}

//执行数据操作
func execute(sql string, args ...interface{}) (result sql.Result) {
	db := GetDB()
	if db == nil {
		fmt.Println("Query 失败")
		return
	}
	result, err := db.Exec(sql, args...)
	if err != nil {
		log.Fatal("执行数据库语句时出错:", err)
		return
	}
	return
}

//执行插入数据
func Insert(sql string, args ...interface{}) (result int64) {
	res := execute(sql, args...)
	if res == nil {
		result = 0
	} else {
		result, _ = res.LastInsertId()
	}
	return
}

//执行删除操作
func Delete(sql string, args ...interface{}) (result int64) {
	res := execute(sql, args...)
	if res == nil {
		result = 0
	} else {
		result, _ = res.RowsAffected()
	}
	return
}

//执行更新操作
func Update(sql string, args ...interface{}) (result int64) {
	res := execute(sql, args...)
	if res == nil {
		result = 0
	} else {
		result, _ = res.RowsAffected()
	}
	return
}
