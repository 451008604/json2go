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
	Array1 []int    `json:"Array1,omitempty"`
	Array2 []string `json:"Array2,omitempty"`
	Array3 []Array3 `json:"Array3,omitempty"`
	Bool   bool     `json:"Bool,omitempty"`
	Int    int      `json:"int,omitempty"`
	NULl   any      `json:"NULL,omitempty"`
	Object Object   `json:"Object,omitempty"`
	String string   `json:"String,omitempty"`
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
