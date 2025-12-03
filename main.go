package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

type structModel struct {
	StructName     string        // 结构体名称
	FieldInfo      []*fieldModel // 字段信息
	IsTransFromMap bool
}

type fieldModel struct {
	FieldName    string // 字段名称
	FieldTypeStr string // 字段类型
	SourceKey    string // 原始key
	SourceValue  any    // 原始value
}

var (
	inputDir, file, pkg, outputDir string // 输入目录、文件名、包名、输出目录
	ResultArr                      []*structModel
	GenerationStartFlag            = "// ==== auto-generated-start ====\n" // 自动生成开始
	GenerationEndFlag              = "// ==== auto-generated-end ====\n"   // 自动生成结束
	GenerationStartStr             = ""
	GenerationEndStr               = ""
	CurFileName                    = ""
	CurInitSModel                  = &structModel{}
)

type replaceModel struct {
	old string        // 要替换的字符串
	new func() string // 替换后的字符串
}

var replaceAll = []replaceModel{
	{old: "${pkg}", new: func() string {
		return pkg
	}},
	{old: "${StartFlag}", new: func() string {
		return GenerationStartFlag
	}},
	{old: "${EndFlag}", new: func() string {
		if GenerationEndStr == "" {
			return GenerationEndFlag
		}
		return GenerationEndStr
	}},
	{old: "${imports}", new: func() string {
		if GenerationStartStr != "" {
			return GenerationStartStr
		}
		return "import (\n\t\"encoding/json\"\n\t\"fmt\"\n\t\"os\"\n\t\"runtime/debug\"\n)\n\n"
	}},
	{old: "${fileName}", new: func() string {
		return CurFileName
	}},
	{old: "${structs}", new: func() string {
		str := ""
		for _, v := range ResultArr {
			str += printStruct(v)
		}
		return str
	}},
	{old: "${isMap}", new: func() string {
		if CurInitSModel.IsTransFromMap {
			return "Data map[int]"
		}
		return "Data"
	}},
	{old: "${structName}", new: func() string {
		return CurInitSModel.StructName
	}},
}

func main() {
	// 定义参数
	i := flag.String("i", "./", "输入目录路径")
	o := flag.String("o", "./", "输出目录路径")
	f := flag.String("f", "", "指定文件名")
	p := flag.String("p", "main", "自定义包名")
	flag.Parse()
	inputDir, outputDir, file, pkg = *i, *o, *f, *p

	// 确保目录路径以斜杠结尾
	if !strings.HasSuffix(inputDir, "/") && !strings.HasSuffix(inputDir, "\\") {
		inputDir += "/"
	}

	files, err := os.ReadDir(inputDir)
	if err != nil {
		log.Printf("Failed to read inputDir %s: %v", inputDir, err)
		return
	}

	for _, f := range files {
		if !strings.Contains(f.Name(), ".json") {
			continue
		}

		if file != "" && file != f.Name() {
			continue
		}

		GenerationStartStr, GenerationEndStr = "", ""
		CurFileName = f.Name()
		readFile, err := os.ReadFile(inputDir + CurFileName)
		if err != nil {
			log.Printf("Failed to read file %s: %v", inputDir+CurFileName, err)
			return
		}
		outputFile := outputDir + strings.Replace(CurFileName, ".json", ".go", 1)
		if goFileContent, _ := os.ReadFile(outputFile); goFileContent != nil {
			if a, b := strings.Index(string(goFileContent), "import"), strings.Index(string(goFileContent), GenerationStartFlag); a >= 0 && b >= 0 {
				GenerationStartStr = string(goFileContent)[a:b]
			}
			if c := strings.Index(string(goFileContent), GenerationEndFlag); c >= 0 {
				GenerationEndStr = string(goFileContent)[c:]
			}
		}

		goContent, err := makeGoFile(string(readFile))
		if err != nil {
			log.Printf("Failed to generate Go file: %v", err)
			return
		}

		err = os.WriteFile(outputFile, []byte(goContent), 0644)
		if err != nil {
			log.Printf("Failed to write file: %v", err)
			return
		}

		log.Printf("Successfully generated Go file: %s", outputFile)
	}
}

func makeGoFile(fileData string) (string, error) {
	// 解析文件内容
	var data map[string]any
	if err := json.Unmarshal([]byte(fileData), &data); err != nil {
		return "", fmt.Errorf("failed to parse JSON: %v", err)
	}

	ResultArr = make([]*structModel, 0)
	// 将json数据转换为结构体
	CurInitSModel = &structModel{StructName: strings.Replace(CurFileName, ".json", "Json", 1)}
	// 开始处理数据
	pullStructModel2ResultArr(CurInitSModel, data)

	// 生成go文件内容
	goContent := "package ${pkg}\n\n" +
		"${imports}" +
		"${StartFlag}" +
		"// from ${fileName}\n\n" +
		"${structs}" +
		"var ${structName}${isMap} ${structName}\n\n" +
		"func Load${structName}(dirPath string) {\n" +
		"\tdata, err := os.ReadFile(dirPath + \"${fileName}\")\n" +
		"\tif err != nil {\n" +
		"\t\tfmt.Printf(\"%v\\n%v\", err, string(debug.Stack()))\n" +
		"\t\treturn\n" +
		"\t}\n" +
		"\t${structName}${isMap} = ${structName}{}\n" +
		"\terr = json.Unmarshal(data, &${structName}Data)\n" +
		"\tif err != nil {\n" +
		"\t\tfmt.Printf(\"%v\\n%v\", err, string(debug.Stack()))\n" +
		"\t\treturn\n" +
		"\t}\n" +
		"}\n\n" +
		"${EndFlag}"

	for _, v := range replaceAll {
		goContent = strings.ReplaceAll(goContent, v.old, v.new())
	}

	return goContent, nil
}

func analyzeType(fModel *fieldModel) *fieldModel {
	if fModel.FieldName == "" {
		fModel.FieldName = toFieldName(fModel.SourceKey)
	}
	switch data := fModel.SourceValue.(type) {
	case map[string]any:
		if checkMapIsSubMap(data) {
			fModel.FieldTypeStr = "" + fModel.FieldName
			// 如果是类数组形式，则以map形式解析
			if checkIsArrMap(data) {
				fModel.FieldTypeStr = "map[int]" + fModel.FieldName
			}
			pullStructModel2ResultArr(&structModel{StructName: fModel.FieldName}, data)

		} else if t := checkMapFieldSameType(data); t != "" {
			fModel.FieldTypeStr = "map[int]" + t

		} else {
			fModel.FieldTypeStr = "" + fModel.FieldName
			pullStructModel2ResultArr(&structModel{StructName: fModel.FieldName}, data)
		}

	case []any:
		if len(data) > 0 {
			switch data[0].(type) {
			case map[string]any:
				fModel.FieldTypeStr = "[]" + fModel.FieldName
				pullStructModel2ResultArr(&structModel{StructName: fModel.FieldName}, data[0].(map[string]any))

			default:
				fModel.FieldTypeStr = "[]" + analyzeType(&fieldModel{SourceValue: data[0]}).FieldTypeStr
			}
		} else {
			fModel.FieldTypeStr = "[]any"
		}

	case float64:
		if math.Mod(data, 1) == 0 {
			fModel.FieldTypeStr = "int"
		} else {
			fModel.FieldTypeStr = "float64"
		}

	case bool:
		fModel.FieldTypeStr = "bool"

	case string:
		fModel.FieldTypeStr = "string"

	default:
		fModel.FieldTypeStr = "any"
	}
	return fModel
}

func pullStructModel2ResultArr(sModel *structModel, data map[string]any) *structModel {
	isExist := false
	for _, v := range ResultArr {
		if v.StructName == sModel.StructName {
			sModel = v
			isExist = true
			break
		}
	}
	if !isExist {
		ResultArr = append(ResultArr, sModel)
	}

	// map[string]map[string]any => map[string]any
	if checkIsArrMap(data) && checkMapIsSubMap(data) {
		data = getMapSubFields(data)
		sModel.IsTransFromMap = true
	}

	// 将map转换为切片
	var dataArr []*fieldModel
	for k, v := range data {
		dataArr = append(dataArr, &fieldModel{SourceKey: k, SourceValue: v})
	}
	// 按字段名排序
	sort.Slice(dataArr, func(i, j int) bool {
		return toFieldName(dataArr[i].SourceKey) < toFieldName(dataArr[j].SourceKey)
	})
	// 将切片转换为结构体
	for _, fModel := range dataArr {
		isExist = false
		// 字段去重
		for _, model := range sModel.FieldInfo {
			if model.SourceKey == fModel.SourceKey {
				isExist = true
				break
			}
		}
		if isExist {
			continue
		}
		sModel.FieldInfo = append(sModel.FieldInfo, analyzeType(fModel))
	}

	return sModel
}

// 打印结构体
func printStruct(sModel *structModel) string {
	goContent := "type " + sModel.StructName + " struct {\n"

	// 计算最长的字段名和类型名，用于格式化对齐
	maxFieldLen, maxTypeLen := 0, 0
	for _, field := range sModel.FieldInfo {
		if len(field.FieldName) > maxFieldLen {
			maxFieldLen = len(field.FieldName)
		}
		if len(field.FieldTypeStr) > maxTypeLen {
			maxTypeLen = len(field.FieldTypeStr)
		}
	}

	// 按字段名排序
	sort.Slice(sModel.FieldInfo, func(i, j int) bool {
		return toFieldName(sModel.FieldInfo[i].SourceKey) < toFieldName(sModel.FieldInfo[j].SourceKey)
	})

	for _, fieldInfo := range sModel.FieldInfo {
		goContent += fmt.Sprintf("\t%-*s %-*s `json:\"%s,omitempty\"`\n", maxFieldLen, toFieldName(fieldInfo.FieldName), maxTypeLen, fieldInfo.FieldTypeStr, fieldInfo.SourceKey)
	}

	goContent += "}\n\n"
	return goContent
}

// 将任意 JSON 键名转换为合法的 Go 字段名（导出形式）。
func toFieldName(s string) string {
	// 如果输入空字符串，直接返回空
	if s == "" {
		return ""
	}

	// 将非字母数字的字符作为分隔符进行分词
	words := strings.FieldsFunc(s, func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsDigit(r))
	})

	// 如果分词失败（例如全是特殊字符），保留原字符串作为唯一词
	if len(words) == 0 {
		words = []string{s}
	}

	// 逐词转换成首字母大写，其余小写（Title Case）
	var result strings.Builder
	for _, word := range words {
		if word != "" {
			// word[:1] 是首字母；word[1:] 是剩余部分
			result.WriteString(strings.ToUpper(word[:1]) + strings.ToLower(word[1:]))
		}
	}

	// 拼接后的字段名
	name := result.String()

	// 如果结果仍为空，返回默认字段名
	if name == "" {
		return "Field"
	}

	// 判断开头字符是否为数字，若是数字，前面加上 "Field"
	if _, err := strconv.ParseFloat(name[0:1], 64); err == nil {
		return "Field" + name
	}

	// 返回最终合法的 Go 字段名
	return name
}

// 检查是否为类数组形式的map
func checkIsArrMap(m map[string]any) bool {
	for k := range m {
		if _, err := strconv.ParseFloat(k, 64); err != nil {
			return false
		}

	}
	return true
}

// 检查map的子级是否全部为map
func checkMapIsSubMap(m map[string]any) bool {
	for _, v := range m {
		switch v.(type) {
		case map[string]any:
		default:
			return false
		}
	}
	return true
}

// 获取map下的全部子map的汇总字段
func getMapSubFields(m map[string]any) (fields map[string]any) {
	if !checkMapIsSubMap(m) {
		return m
	}
	fields = map[string]any{}
	for _, v := range m {
		for k1, v1 := range v.(map[string]any) {
			fields[k1] = v1
		}
	}
	return fields
}

// 检查map字段是否为相同类型
func checkMapFieldSameType(m map[string]any) string {
	if !checkIsArrMap(m) {
		return ""
	}
	lastType := ""
	for _, v := range m {
		t := ""
		switch v.(type) {
		case int:
			t = "int"
		case float64:
			t = "float64"
		case string:
			t = "string"
		case bool:
			t = "bool"
		default:
			return ""
		}

		if lastType != "" && t != lastType {
			return ""
		}
		lastType = t
	}
	return lastType
}
