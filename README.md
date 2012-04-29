GoORM
=====
GoORM is an ORM for Go. It lets you map Go structs to tables in a database. It's intended to be very lightweight, doing very little beyond what you really want. For example, when fetching data, instead of re-inventing a query syntax, we just delegate your query to the underlying database, so you can write the "where" clause of your SQL statements directly. This allows you to have more flexibility while giving you a convenience layer. But GoORM also has some smart defaults, for those times when complex queries aren't necessary.

### Installing GoORM
    go get github.com/astaxie/goorm

### How do we use it?

Open a database

	orm := goorm.NewORM("127.0.0.1", "3306", "test", "xiemengjun", "123456", "utf8")

Change Database

	orm.SelectDb("test2")  

Model a struct after a table in the db

	type Person struct {
		Id int64
		Name string
		Age int64
	}

Create an object and save it

	var someone Person
	someone.Name = "john"
	someone.Age = 20

	orm.Save(&someone)

Fetch a single object

	var person1 Person
	orm.Get(&person1, "id = ?", 3)

	var person2 Person
	orm.Get(&person2, 3) // this is shorthand for the version above

	var person3 Person
	orm.Get(&person3, "name = ?", "john") // more complex query

	var person4 Person
	orm.Get(&person4, "name = ? and age < ?", "john", 88) // even more complex

Fetch multiple objects

	var bobs []Person
	err := orm.GetAll(&bobs, "name = ?", "bob")

	var everyone []Person
	err := orm.GetAll(&everyone, "") // use empty string to omit "where" clause

Saving new and existing objects

	person2.Name = "Jack" // an already-existing person in the database, from the example above
	db.Save(&person2)

	var newGuy Person
	newGuy.Name = "that new guy"
	newGuy.Age = 27

	db.Save(&newGuy)
	// newGuy.Id is suddenly valid, and he's in the database now.
