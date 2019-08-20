package main

import (
	"db/orm"
	"fmt"
	"reflect"
)

type A struct {
	Name string
	Id   uint32
}

type B struct {
	orm.BaseModel
}

func test(a *A) {
	valuePtr := reflect.ValueOf(a)
	value := valuePtr.Elem()
	value.Set(reflect.Append(value, reflect.ValueOf(A{Name: "Test"})))
}

func main() {

	//ctx := context.WithValue(context.Background(), "id", "192jj2jj2j")
	//ctx, _ = context.WithTimeout(ctx, 2*time.Second)
	//ctx = context.WithValue(ctx, "id1", "zzzz")
	//
	//select {
	//case c := <-ctx.Done():
	//	fmt.Println(ctx.Value("id"))
	//	fmt.Println(ctx.Value("id"))
	//	fmt.Println(ctx.Err())
	//	fmt.Println(c, "Done!")
	//}

	//var arr []A
	//fmt.Println(db.ElementName(&arr))

	//var arr A
	//
	//m := &arr
	//
	//fmt.Println(reflect.TypeOf(m).Kind())
	//
	//a := reflect.New(reflect.TypeOf(m).Elem().Elem()).Interface().(*A)
	//fmt.Println(a)

	var b B
	b.Id = 1
	fmt.Println(b.TableName())
}
