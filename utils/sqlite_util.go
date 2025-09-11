package utils

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"reflect"
	"time"

	_ "modernc.org/sqlite"
)

// SQLite3Util 工具类
type SQLite3Util struct {
	db          *sql.DB
	dbPath      string
	MaxIdle     int           // 最大空闲连接数
	MaxOpen     int           // 最大打开连接数
	MaxLifetime time.Duration // 连接最大生命周期
}

// Options 初始化选项
type Options struct {
	DBPath      string
	MaxIdle     int
	MaxOpen     int
	MaxLifetime time.Duration
}

type TencentGroup struct {
	ID          int32  `db:"id"`
	GroupId     string `db:"group_id"`
	GroupOpenid string `db:"group_openid"`
}

type TencentAuthor struct {
	ID           int64  `db:"id"`
	AuthorId     string `db:"author_id"`
	MemberOpenid string `db:"member_openid"`
}
type TencentGroupMessage struct {
	ID           int32  `db:"id"`
	SelfGroupId  int32  `db:"self_group_id"`
	SelfSenderId int64  `db:"self_sender_id"`
	MessageId    string `db:"message_id"`
	TimeStamp    int    `db:"time_stamp"`
}

var DBUtil *SQLite3Util

// DefaultOptions 默认选项
var DefaultOptions = Options{
	DBPath:      ":memory:",
	MaxIdle:     10,
	MaxOpen:     100,
	MaxLifetime: time.Hour,
}

// NewSQLite3Util 工具实例
func NewSQLite3Util(options Options) (*SQLite3Util, error) {
	db, err := sql.Open("sqlite3", options.DBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// 设置连接池参数
	db.SetMaxIdleConns(options.MaxIdle)
	db.SetMaxOpenConns(options.MaxOpen)
	db.SetConnMaxLifetime(options.MaxLifetime)

	// 测试连接
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	return &SQLite3Util{
		db:          db,
		dbPath:      options.DBPath,
		MaxIdle:     options.MaxIdle,
		MaxOpen:     options.MaxOpen,
		MaxLifetime: options.MaxLifetime,
	}, nil
}

// Close 关闭数据库连接
func (s *SQLite3Util) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// GetDB 获取数据库连接实例
func (s *SQLite3Util) GetDB() *sql.DB {
	return s.db
}

// Exec 执行非查询SQL (创建表、插入、更新、删除等)
func (s *SQLite3Util) Exec(sqlStr string, args ...interface{}) (sql.Result, error) {
	if s.db == nil {
		return nil, errors.New("database is not initialized")
	}

	stmt, err := s.db.Prepare(sqlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute statement: %v", err)
	}

	return result, nil
}

// Query 执行查询SQL
func (s *SQLite3Util) Query(sqlStr string, args ...interface{}) (*sql.Rows, error) {
	if s.db == nil {
		return nil, errors.New("database is not initialized")
	}

	stmt, err := s.db.Prepare(sqlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %v", err)
	}

	return rows, nil
}

// QueryRow 查询单行
func (s *SQLite3Util) QueryRow(sqlStr string, args ...interface{}) *sql.Row {
	if s.db == nil {
		return nil
	}

	return s.db.QueryRow(sqlStr, args...)
}

// Insert 插入数据并返回最后插入的ID
func (s *SQLite3Util) Insert(sqlStr string, args ...interface{}) (int64, error) {
	result, err := s.Exec(sqlStr, args...)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %v", err)
	}

	return id, nil
}

// Update 更新数据并返回受影响的行数
func (s *SQLite3Util) Update(sqlStr string, args ...interface{}) (int64, error) {
	result, err := s.Exec(sqlStr, args...)
	if err != nil {
		return 0, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get affected rows: %v", err)
	}

	return rows, nil
}

// Delete 删除数据并返回受影响的行数
func (s *SQLite3Util) Delete(sqlStr string, args ...interface{}) (int64, error) {
	return s.Update(sqlStr, args...)
}

// Begin 开始事务
func (s *SQLite3Util) Begin() (*sql.Tx, error) {
	if s.db == nil {
		return nil, errors.New("database is not initialized")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %v", err)
	}

	return tx, nil
}

// Transaction 执行事务
func (s *SQLite3Util) Transaction(f func(tx *sql.Tx) error) error {
	tx, err := s.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // 重新抛出panic
		}
	}()

	err = f(tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// TableExists 检查表是否存在
func (s *SQLite3Util) TableExists(tableName string) (bool, error) {
	sqlStr := "SELECT name FROM sqlite_master WHERE type='table' AND name=?"
	row := s.QueryRow(sqlStr, tableName)

	var name string
	err := row.Scan(&name)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// GetTableSchema 获取表结构信息
func (s *SQLite3Util) GetTableSchema(tableName string) ([]map[string]interface{}, error) {
	sqlStr := fmt.Sprintf("PRAGMA table_info(%s)", tableName)
	rows, err := s.Query(sqlStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		rowMap := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				rowMap[col] = string(b)
			} else {
				rowMap[col] = val
			}
		}
		result = append(result, rowMap)
	}

	return result, nil
}

// QueryToStructs 将查询结果映射到结构体切片 (需要结构体字段使用tag `db:"column_name"`)
func (s *SQLite3Util) QueryToStructs(dest interface{}, sqlStr string, args ...interface{}) error {
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr || destValue.Elem().Kind() != reflect.Slice {
		return errors.New("dest must be a pointer to a slice")
	}

	sliceValue := destValue.Elem()
	elementType := sliceValue.Type().Elem()

	rows, err := s.Query(sqlStr, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	for rows.Next() {
		// 创建新元素
		elem := reflect.New(elementType).Elem()

		// 准备扫描值的切片
		values := make([]interface{}, len(columns))
		for i := range values {
			fieldIndex := -1
			// 查找匹配的字段
			for j := 0; j < elementType.NumField(); j++ {
				field := elementType.Field(j)
				tag := field.Tag.Get("db")
				if tag == columns[i] {
					fieldIndex = j
					break
				}
			}

			if fieldIndex >= 0 {
				values[i] = elem.Field(fieldIndex).Addr().Interface()
			} else {
				// 如果没有匹配的字段，使用空接口
				values[i] = new(interface{})
			}
		}

		if err := rows.Scan(values...); err != nil {
			return err
		}

		sliceValue.Set(reflect.Append(sliceValue, elem))
	}

	return rows.Err()
}

func SqLiteInit() {
	// 初始化数据库连接
	dbUtil, err := NewSQLite3Util(Options{
		DBPath:      "./sqlite.db",
		MaxIdle:     10,
		MaxOpen:     100,
		MaxLifetime: time.Hour,
	})
	DBUtil = dbUtil
	if err != nil {
		log.Fatal(err)
	}
	//defer dbUtil.Close()

	// 创建表
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS TencentGroup (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		group_id TEXT NOT NULL,
		group_openid TEXT NOT NULL
	)
	`
	_, err = dbUtil.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}

	createTableSQL = `
	CREATE TABLE IF NOT EXISTS TencentAuthor (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		member_openid TEXT NOT NULL,
		author_id TEXT NOT NULL
	)
	`
	_, err = dbUtil.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}

	createTableSQL = `
	CREATE TABLE IF NOT EXISTS TencentGroupMessage (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		message_id TEXT NOT NULL,
		self_group_id  INTEGER NOT NULL,
	    self_sender_id INTEGER NOT NULL,
		time_stamp INTEGER NOT NULL 
	)
	`
	_, err = dbUtil.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}
}

func (s *SQLite3Util) GroupInsert(groupId string, groupOpenId string) (int32, error) {
	selectSQL := "SELECT * FROM TencentGroup WHERE group_id = ? AND group_openid = ?"
	var tgs []TencentGroup
	err := s.QueryToStructs(&tgs, selectSQL, groupId, groupOpenId)
	if tgs != nil {
		return tgs[0].ID, err
	}

	insertSQL := "INSERT INTO TencentGroup (group_id, group_openid) VALUES (?, ?)"
	id, err := s.Insert(insertSQL, groupId, groupOpenId)

	if err != nil {
		return 0, err
	}

	return int32(id), nil
}
func (s *SQLite3Util) SenderInsert(memberOpenId string, authorId string) (int64, error) {
	selectSQL := "SELECT * FROM TencentAuthor WHERE member_openid = ? AND author_id = ?"
	var tas []TencentAuthor
	err := s.QueryToStructs(&tas, selectSQL, memberOpenId, authorId)
	if tas != nil {
		return tas[0].ID, err
	}

	insertSQL := "INSERT INTO TencentAuthor (member_openid, author_id) VALUES (?, ?)"
	id, err := s.Insert(insertSQL, memberOpenId, authorId)

	if err != nil {
		return 0, err
	}

	return id, nil
}
func (s *SQLite3Util) GroupMessageInsert(messageId string, selfGroupId int32, selfSenderId int64) (int32, error) {
	selectSQL := "SELECT * FROM TencentGroupMessage WHERE message_id = ? and self_group_id = ? AND self_sender_id = ?"
	var tgms []TencentGroupMessage
	err := s.QueryToStructs(&tgms, selectSQL, messageId, selfGroupId, selfSenderId)
	if tgms != nil {
		return tgms[0].ID, err
	}

	insertSQL := "INSERT INTO TencentGroupMessage (message_id,self_group_id,self_sender_id,time_stamp) VALUES (?,?,?,?)"
	id, err := s.Insert(insertSQL, messageId, selfGroupId, selfSenderId, time.Now().Unix())

	if err != nil {
		return 0, err
	}

	return int32(id), nil
}

func (s *SQLite3Util) GetGroupMessageID(selfGroupId int32, selfSenderId int64) (string, error) {
	selectSQL := "SELECT * FROM TencentGroupMessage WHERE self_group_id = ? AND self_sender_id = ? ORDER BY id DESC"
	var tgms []TencentGroupMessage
	err := s.QueryToStructs(&tgms, selectSQL, selfGroupId, selfSenderId)
	if err == nil {
		if len(tgms) > 0 {
			return tgms[0].MessageId, nil
		}
	}
	return "", err
}
func (s *SQLite3Util) GetGroupID(messageId int32) (string, error) {
	selectSQL := "SELECT * FROM TencentGroup WHERE id = ?"
	var tgms []TencentGroup
	err := s.QueryToStructs(&tgms, selectSQL, messageId)
	if err == nil {
		if len(tgms) > 0 {
			return tgms[0].GroupOpenid, nil
		}
	}
	return "", err
}
func (s *SQLite3Util) GetSenderID(messageId int32) (string, error) {
	selectSQL := "SELECT * FROM TencentAuthor WHERE id = ?"
	var tgms []TencentAuthor
	err := s.QueryToStructs(&tgms, selectSQL, messageId)
	if err == nil {
		if len(tgms) > 0 {
			return tgms[0].MemberOpenid, nil
		}
	}
	return "", err
}
