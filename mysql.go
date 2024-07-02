package abc

import (
	"github.com/chenqgp/abc/conf"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

//var (
//	mysqlDB     = "ex3"
//	mysqlServer = "154.23.187.172:3306"
//	mysqlUser   = "ex3"
//	mysqlPasswd = "helloex"
//)

//var (
//	mysqlDB     = "hello_x"
//	mysqlServer = "152.32.169.123:3306"
//	mysqlUser   = "hello_x"
//	mysqlPasswd = "Rm3XT7rYjG5X3JW7"
//)

func bootSql() {
	//open a db connection
	log.Println(conf.MysqlUser + ":" + conf.MysqlPasswd + "(" + conf.MysqlServer + ")/" + conf.MysqlDB)
	var err error
	db, err = gorm.Open(mysql.Open(conf.MysqlUser+":"+conf.MysqlPasswd+"@tcp("+conf.MysqlServer+")/"+conf.MysqlDB), &gorm.Config{})
	if err != nil {
		panic(err)
	}
}
