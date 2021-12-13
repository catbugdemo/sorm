# sorm 

## 前言
- 本代码的核心架构为 Geeorm 架构 ，参考 7 days geeorm 和 gorm 编写而成
  - 学习并编写该项目的原因是因为 在 使用了 gorm 后，项目架构变更为 sqlx 的时候感觉用的不顺手
  - 该项目依赖核心包 golang "database/sql"
  - 目前还没写 mysql 支持
- 推荐使用 gorm 或者 sqlx ，因为前两个项目更加成熟
- 其中添加了一些自己的思考，新增了 postgres 数据库支持  
- 新增批量插入,同时返回插入后的数据
- 支持 where 重叠 以及 in 的支持

## 参考
- 7 days golang programs from scratch  https://github.com/geektutu/7days-golang
- gorm https://github.com/go-gorm/gorm
## 使用方法
### 1.连接
#### 1. 基础使用
```go
    // 获取 db
    db ,err :=sorm.Open("postgres",fmt.Sprint("host=127.0.0.1 port=5432 user=postgres password=123456 dbname=mydb sslmode=disable"))
    if err!= nil {
      return err
    }
  
    type UserTest struct {
      Id string `db:"id" sorm:"primary key"`
      CreatedTime `db:"created_time"`
      Name string `db:"name"`
    }
    
    // 设置 Model -- 该行为会自动转换 UserTest 内部信息
    // 默认设置 表名 user_test，column 映射为  tag 中 db 的 数据
    db.Model(&UserTest{})
```
#### 2. 替换 sqlx 
```go
    sqlxDB, err := sqlx.Open(driverName,dataSourceName)
    if err!= nil {
    	return err
    }
    // db 
    db, err := sorm.ReplaceSqlx(sqlxDB)
    if err!= nil {
    	return err
    }
    db.First()
```

### 2.CRUD
#### 1.新增
- Create 新增一条数据 
```go

    ut := UserTest{Name:"test"}
    if err := db.Create(&ut);err!=nil {
      return err
    }
    // INSERT INTO user_test(name) VALUES('test')
    fmt.Println(ut)
```
- Insert 新增多条数据
```go
    ut1 := UserTest{CreatedTime: time.Now(),Name:"test1"}
    ut2 := UserTest{CreatedTime: time.Now(),Name:"test2"}
    uts := make([]UserTest,0,2)
    uts = append(uts,ut1,ut2)
    if err := db.Insert(&uts);err!=nil {
    	return err
    }
    // INSERT INTO user_test(created_time,name) VALUES('2006-01-02 15:04:05','test1'),('2006-01-02 15:04:05','test2')
    fmt.Println(uts)
```

#### 2.查询
- First 查询一条数据
```go
    var ut UserTest
    if err := db.First(&ut);err!=nil {
        return err
    }
    // SELECT id,name FROM user_test Limit 1
    fmt.Println(ut)
```
- Find 查询多条记录
```go
    var uts []UserTest
    if err := db.Where("id in (?)",[]int{1,2}).Find(&uts);err!=nil{
        return err
    }
    // SELECT id,name FROM user_test WHERE id in (1,2) 
```
- Where 条件
```go
    var ut UserTest
    db.Where("id=?",1).First(&ut)
    // SELECT id,created_time,name FROM user_test WHERE id='1'
    
    // 支持 Where 重叠
    var uts []UserTest
    db.Where("name=?","test").Where("id in (?)",[]int{1,2}).Find(&uts)
    // SELECT id,created_time,name FROM user_test WHERE name='test' and id in ('1','2')
```
- Limit ,Offset
```go
    var uts []UserTest
    db.Limit(1).Offset(1).Find(&ut)
    // SELECT id,created_time,name FROM user_test LIMIT 1 OFFSET 1
```
- Select (选择需要查询的数据) 支持 数组 和 单个数据
```go
    var ut UserTest
    db.Select("name").First(&ut)
    // SELECT name FROM user_test
```
- Table (自主选择查询表名)
```go
    var ids []int
    db.Select("id").Table("user_test").Rows().Scan(&ids)
    // SELECT id FROM user_test
```
- Count 查询总数
```go
    var count int 
    db.Model(&ut).Count(&count)
```
#### 3.修改
- Update 单一参数修改
```go
    var ut UserTest
    if err:= db.Model(&ut).Update("name","test");err!=nil {
    	return err
    }
    // UPDATE user_test SET name='test'
```
- Updates 多参数修改 (支持结构体，map[string]interface{})
```go
    ut := UserTest{CreatedTime:time.Now(),Name:"test"}
    m := map[string]interface{}{
    	"created_time":time.Now(),
        "name":"test",
    }
    if err := db.Updates(&ut);err!=nil {
    	return err
    }
    if err := db.Model(&ut).Updates(m);err!=nil {
    	return err
    }
    // UPDATE user_test set created_time='2006-01-02 15:04:05',name='test'
```
#### 4.删除 
```go
    var ut UserTest
    if err:= db.Model(&ut).Where("id=?",1).Delete();err!=nil {
    	return err
    }
    // DELETE FROM user_test Where id='1'
```
#### 5.支持自主语句
```go
    var id int 
    db.Raw("select id from user_test where name=?",'test').Scan(&id)

    // or 
    db.Select("id").Table("user_test").Where("name=?",'test').Rows().Scan(&id)
    // SELECT id FROM user_test WHERE name='test'
```
### 3.支持 golang database/db 原生查询
```go
    db ,err :=sorm.Open("postgres",fmt.Sprint("host=127.0.0.1 port=5432 user=postgres password=123456 dbname=mydb sslmode=disable"))
    if err!= nil {
        return err
    }
    sql := "select * from user_test Where id=? and name=?"
    
    // Exec
    db.Raw(sql,1,"111").Exec()
    
    // QueryRow
    db.Raw(sql,1,"111").QuerRow()
    
    // QueryRows
    db.Raw(sql,1,"111").QueryRows()
```

### 4.钩子
```go
    BeforeQuery  
    AfterQuery  
    BeforeUpdate
    AfterUpdate 
    BeforeDelete
    AfterDelete
    BeforeInsert 
    AfterInsert 
```
#### 1.使用方法
```go
  type UserTest struct {
    Id string `db:"id" sorm:"primary key"`
    CreatedTime `db:"created_time"`
    Name string `db:"name"`
  }
  
  // 在查询后，返回参数 Name 都为空
  func (o *UserTest) AfterQuery(s *session.Session) {
  	o.Name = ""
  	return nil
  }
  
  func main () {
    db ,err :=sorm.Open("postgres",fmt.Sprint("host=127.0.0.1 port=5432 user=postgres password=123456 dbname=mydb sslmode=disable"))
    if err!= nil {
        return err
    }

    var ut UserTest
    if err := ut.First(&ut);err!=nil {
    	return err
    }
    fmt.Println(ut)
    // 这个时候打印的 name 参数为空
  }
```
### 待补充