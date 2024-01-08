package scaffold

import (
	"fmt"
	"reflect"
	"strings"
)

func generateYaml(yamlBuilder *strings.Builder, reflectValue reflect.Value, indentLevel int, commentPrefix string) {
	t := reflectValue.Type()

	indentPrefix := strings.Repeat("  ", indentLevel)
	originalCommentPrefix := commentPrefix
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := reflectValue.Field(i)
		tag := field.Tag.Get("yaml")
		comment := field.Tag.Get("comment")

		if value.IsZero() && !strings.Contains(comment, "required") {
			commentPrefix = "# "
		}

		if value.Kind() == reflect.Struct {
			yamlBuilder.WriteString(fmt.Sprintf("%s%s%s: # %s\n", commentPrefix, indentPrefix, tag, comment))
			generateYaml(yamlBuilder, value, indentLevel+1, commentPrefix)
			continue
		}

		if value.Kind() == reflect.Slice || value.Kind() == reflect.Map {
			yamlBuilder.WriteString(fmt.Sprintf("%s%s%s: # %s\n", commentPrefix, indentPrefix, tag, comment))
			for j := 0; j < value.Len(); j++ {
				valueStr := getValueStr(value.Index(j))
				yamlBuilder.WriteString(fmt.Sprintf("%s%s  - %s\n", commentPrefix, indentPrefix, valueStr))
			}
			if value.Len() == 0 {
				yamlBuilder.WriteString(fmt.Sprintf("%s%s  - \n", commentPrefix, indentPrefix))
			}
			continue
		}

		valueStr := getValueStr(value)
		yamlBuilder.WriteString(fmt.Sprintf("%s%s%s: %s # %s\n", commentPrefix, indentPrefix,
			tag, valueStr, comment))

		commentPrefix = originalCommentPrefix
	}
}

func getValueStr(value reflect.Value) string {
	valueStr := ""
	if value.Kind() == reflect.String {
		valueStr = fmt.Sprintf("\"%v\"", value.Interface())
	} else {
		valueStr = fmt.Sprintf("%v", value.Interface())
	}
	return valueStr
}
