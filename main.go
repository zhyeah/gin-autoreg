// package main

// import (
// 	"fmt"

// 	"github.com/zhyeah/gin-autoreg/util"
// )

// type A struct {
// 	Id    *int `json:"id,omitempty"`
// 	Age   int
// 	Score string `json:"score"`
// 	Money bool   `json:"money"`
// 	Empty string
// 	Info  *B `json:"info"`
// }

// type B struct {
// 	Cool  []int  `json:"cool"`
// 	Buddy int    `json:"buddy"`
// 	Laugh string `json:"laugh"`
// }

// func main() {
// 	jsonText := "{\"id\":\"12332\",\"age\":\"10\",\"score\":123,\"money\":true,\"info\":{\"cool\":[1,2,3],\"buddy\":\"567\",\"laugh\":765}}"
// 	a := &A{}
// 	err := util.AdaptJSONForDTO(jsonText, a)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	fmt.Printf("%v %v", *a.Id, a.Info)
// }
