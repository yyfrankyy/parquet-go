package schema2struct

import (
	"fmt"
	"string"

	"github.com/xitongsys/parquet-go/parquet"
)

func ParquetTypeToParquetTypeStr(pT *parquet.ParquetType, cT *parquet.ConvertedType) (string, string) {
	var pTStr, cTStr string
	switch *pT {
	case parquet.Type_BOOLEAN:
		pTStr = "BOOLEAN"
	case parquet.Type_INT32:
		pTStr = "INT32"
	case parquet.Type_INT64:
		pTStr = "INT64"
	case parquet.Type_INT96:
		pTStr = "INT96"
	case parquet.Type_FLOAT:
		pTStr = "FLOAT"
	case parquet.Type_DOUBLE:
		pTStr = "DOUBLE"
	case parquet.Type_BYTE_ARRAY:
		pTStr = "BYTE_ARRAY"
	case parquet.Type_FIXED_LEN_BYTE_ARRAY:
		pTStr = "FIXED_LEN_BYTE_ARRAY"
	}
	switch *cT {
	case parquet.ConvertedType_UTF8:
		cTStr = "UTF8"
	case parquet.ConvertedType_INT_8:
		cTStr = "INT_8"
	case parquet.ConvertedType_INT_16:
		cTStr = "INT_16"
	case parquet.ConvertedType_INT_32:
		cTStr = "INT_32"
	case parquet.ConvertedType_INT_64:
		cTStr = "INT_64"
	case parquet.ConvertedType_UINT_8:
		cTStr = "UINT_8"
	case parquet.ConvertedType_UINT_16:
		cTStr = "UINT_16"
	case parquet.ConvertedType_UINT_32:
		cTStr = "UINT_32"
	case parquet.ConvertedType_UINT_64:
		cTStr = "UINT_64"
	case parquet.ConvertedType_DATE:
		cTStr = "DATE"
	case parquet.ConvertedType_TIME_MILLIS:
		cTStr = "TIME_MILLIS"
	case parquet.ConvertedType_TIME_MICROS:
		cTStr = "TIME_MICROS"
	case parquet.ConvertedType_TIMESTAMP_MILLIS:
		cTStr = "TIMESTAMP_MILLIS"
	case parquet.ConvertedType_TIMESTAMP_MICROS:
		cTStr = "TIMESTAMP_MICROS"
	case parquet.ConvertedType_INTERVAL:
		cTStr = "INTERVAL"
	case parquet.ConvertedType_DECIMAL:
		cTStr = "DECIMAL"
	case parquet.ConvertedType_MAP:
		cTStr = "MAP"
	case parquet.ConvertedType_LIST:
		cTStr = "LIST"
	}
	return pTStr, cTStr
}

func ParquetTypeToGoTypeStr(pT *parquet.ParquetType, cT *parquet.ConvertedType) string {
	res := ""
	switch *pT {
	case parquet.Type_BOOLEAN:
		res = "bool"
	case parquet.Type_INT32:
		res = "int32"
	case parquet.Type_INT64:
		res = "int64"
	case parquet.Type_INT96:
		res = "string"
	case parquet.Type_FLOAT:
		res = "float32"
	case parquet.Type_DOUBLE:
		res = "float64"
	case parquet.Type_BYTE_ARRAY:
		res = "string"
	case parquet.Type_FIXED_LEN_BYTE_ARRAY:
		res = "string"
	}
	if cT != nil {
		switch *cT {
		case parquet.ConvertedType_UINT_8:
			res = "uint32"
		case parquet.ConvertedType_UINT_16:
			res = "uint32"
		case parquet.ConvertedType_UINT_32:
			res = "uint32"
		case parquet.ConvertedType_UINT_64:
			res = "uint64"
		}
	}
}

type Node struct {
	SE       *parquet.SchemaElement
	Children []*Node
}

func NewNode(schema *parquet.SchemaElement) *Node {
	node := &(Node{
		SE:       schema,
		Children: []*Node{},
	})
	return node
}

func (self *Node) OutputJsonSchema() string {
	return ""
}

func (self *Node) OutputStruct() string {
	name := self.SE.GetName()
	res := name
	pT, cT := self.SE.GetType(), self.SE.GetConvertedType()
	pTStr, cTStr := ParquetTypeToParquetTypeStr(pT, cT)
	rTStr := ""
	if self.SE.GetRepetitionType == parquet.FieldRepetitionType_OPTIONAL {
		rTStr = "*"
	} else if self.SE.GetConvertedType == parquet.FieldRepetitionType_REPEATED {
		rTStr = "[]"
	}

	if pT == nil && cT == nil {
		res += " struct {\n"
		for _, cNode := range self.Children {
			res += "\t" + cNode.OutputStruct() + "\n"
		}
		res += "}"

	} else if cT != nil && *cT == parquet.ConvertedType_MAP {
		keyNode := self.Children[0].Children[0]
		keyPT, keyCT := keyNode.SE.GetType(), keyNode.SE.GetConvertedType()
		keyGoTypeStr := ParquetTypeToGoTypeStr(pT, cT)
		valNode := self.Children[0].Children[1]
		res += " " + "map[" + keyGoTypeStr + "]" + valNode.OutputStruct()

	} else if cT != nil && *cT == parquet.ConvertedType_LIST {
		cNode := self.Children[0].Children[0]
		res += " " + rTStr + " " + cNode.OutputStruct()

	} else {
		goTypeStr := ParquetTypeToGoTypeStr(pT, cT)
		res += " " + rTStr + " " + goTypeStr
	}
	return res

}

func CreateTree(schemas []*parquet.SchemaElement) *Node {
	ln := len(schemas)
	pos := 0
	stack := make([]*Node, 0)
	root := NewNode(schemas[0])
	stack = append(stack, root)
	pos++

	for len(stack) > 0 {
		node := stack[len(stack)-1]
		numChildren := int(node.SE.GetNumChildren())
		lc := len(node.Children)
		if lc < numChildren {
			newNode := NewNode(schemas[pos])
			node.Children = append(node.Children, newNode)
			stack = append(stack, newNode)
			pos++
		} else {
			stack = stack[:len(stack)-1]
		}
	}
	return root

}
