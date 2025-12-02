package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"
)

// ==== auto-generated-start ====
// from Test.json

type TestJson struct {
	A        string   `json:"A,omitempty"`
	Array2   []string `json:"Array2,omitempty"`
	Array3   []Array3 `json:"array3,omitempty"`
	Arrayid1 []int    `json:"Arrayid1,omitempty"`
	Bool     bool     `json:"bool,omitempty"`
	Int      int      `json:"int,omitempty"`
	Null     any      `json:"null,omitempty"`
	Object   Object   `json:"oBject,omitempty"`
	String   string   `json:"string,omitempty"`
}

type Array3 struct {
	Id   int    `json:"ID,omitempty"`
	Name string `json:"Name,omitempty"`
}

type Object struct {
	Obj1 Obj1 `json:"Obj1,omitempty"`
	Obj2 Obj2 `json:"Obj2,omitempty"`
	Obj3 Obj3 `json:"Obj3,omitempty"`
}

type Obj1 struct {
	Id   int    `json:"ID,omitempty"`
	Name string `json:"Name,omitempty"`
}

type Obj2 struct {
	Id   int    `json:"ID,omitempty"`
	Name string `json:"Name,omitempty"`
}

type Obj3 struct {
	Id   int    `json:"ID,omitempty"`
	Name string `json:"Name,omitempty"`
}

var TestJsonData TestJson

func LoadTestJson(dirPath string) {
	data, err := os.ReadFile(dirPath + "Test.json")
	if err != nil {
		fmt.Printf("%v\n%v", err, string(debug.Stack()))
		return
	}
	TestJsonData = TestJson{}
	err = json.Unmarshal(data, &TestJsonData)
	if err != nil {
		fmt.Printf("%v\n%v", err, string(debug.Stack()))
		return
	}
}

// ==== auto-generated-end ====
