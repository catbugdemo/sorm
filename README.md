## sorm 

- 参考 7 days Geeorm 编写而成
- 其中添加了一些自己的思考，新增了 postgres 数据库支持  
- 新增批量插入,同时返回插入后的数据
- 支持 where 的重叠
- 优化：如果为空返回错误
- 支持 where in 中 的数据插入

- https://github.com/geektutu

## 使用方法
### 1. 直接使用
```go
    db ,err :=sorm.Open("postgres",fmt.Sprint("host=127.0.0.1 port=5432 user=postgres password=123456 dbname=mydb sslmode=disable"))
    if err!= nil {
    	return err
    }

    type UserTest struct {
        Id string `db:"id" sorm:"primary key"`
        Name string `db:"name"`
    }
    
    // Model -- 存储相应结构体内容
    db = db.Model(&UserTest{})
    
    // 单数量查询
    var ut UserTest
    if err := db.First(&ut);err!=nil {  
    	return err
    }
    // SELECT id,name FROM user_test Limit 1
    fmt.Println(ut)
    
    // Where + Limit + Offset + Find 
    var uts []UserTest
	if err := db.Where("id in (?)",[]int{1,2}).Limit(10).Offset(1).Find(&uts);err!=nil{
		return err
    }
    // SELECT id,name FROM user_test WHERE id in (1,2)  LIMIT 10 OFFSET 1
    
    // Select + Table + Find
    db.Select("name").Table("test").Find(&)
    
    // Create  -- 单数据创建
	ut2 :=UserTest{
		Name: "test"
    }
    if err := db.Create(&ut);err!=nil {
    	return err  
    }
    // INSERT INTO user_test(name) VALUES('test') RETURNING id,name
    
    // Insert --  批量创建
    if err := db.Insert(&uts);err!=nil {
    	return err
    }
    // INSERT INTO user_test(id,name) VALUES('?','?'),('?','?')
    
    // Update -- 更新
    
    
    
```