package main

import (
	"fmt"
	"github.com/astaxie/goorm"
)

type User struct {
	Id       int
	Username string
}

func main() {	
	orm := goorm.NewORM("127.0.0.1", "3306", "dds", "xiemengjun", "123456", "utf8")

	var user []User	
	//getall
	orm.GetAll(&user, "id>?", "1")
	fmt.Println(user)
}
