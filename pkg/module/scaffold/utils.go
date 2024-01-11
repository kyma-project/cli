package scaffold

import (
	"fmt"
	"reflect"
	"strings"
)

func generateYaml(obj interface{}) string {
	reflectValue := reflect.ValueOf(obj)
	var yamlBuilder strings.Builder
	generateYamlWithComments(&yamlBuilder, reflectValue, 0, "")
	return yamlBuilder.String()
}

// generateYamlWithComments uses a "comment" tag in the struct definition to generate YAML with comments on corresponding lines.
// Note: Map support is missing!
func generateYamlWithComments(yamlBuilder *strings.Builder, obj reflect.Value, indentLevel int, commentPrefix string) {
	t := obj.Type()

	indentPrefix := strings.Repeat("  ", indentLevel)
	originalCommentPrefix := commentPrefix
	for i := 0; i < t.NumField(); i++ {
		commentPrefix = originalCommentPrefix
		field := t.Field(i)
		value := obj.Field(i)
		yamlTag := field.Tag.Get("yaml")
		commentTag := field.Tag.Get("comment")

		// comment-out non-required empty attributes
		if value.IsZero() && !strings.Contains(commentTag, "required") {
			commentPrefix = "# "
		}

		if value.Kind() == reflect.Struct {
			if commentTag == "" {
				yamlBuilder.WriteString(fmt.Sprintf("%s%s%s:\n", commentPrefix, indentPrefix, yamlTag))
			} else {
				yamlBuilder.WriteString(fmt.Sprintf("%s%s%s: # %s\n", commentPrefix, indentPrefix, yamlTag, commentTag))
			}
			generateYamlWithComments(yamlBuilder, value, indentLevel+1, commentPrefix)
			continue
		}

		if value.Kind() == reflect.Slice {
			if commentTag == "" {
				yamlBuilder.WriteString(fmt.Sprintf("%s%s%s:\n", commentPrefix, indentPrefix, yamlTag))
			} else {
				yamlBuilder.WriteString(fmt.Sprintf("%s%s%s: # %s\n", commentPrefix, indentPrefix, yamlTag, commentTag))
			}

			if value.Len() == 0 {
				yamlBuilder.WriteString(fmt.Sprintf("%s%s  -\n", commentPrefix, indentPrefix))
			}
			for j := 0; j < value.Len(); j++ {
				valueStr := getValueStr(value.Index(j))
				yamlBuilder.WriteString(fmt.Sprintf("%s%s  - %s\n", "", indentPrefix, valueStr))
			}
			continue
		}

		valueStr := getValueStr(value)
		if commentTag == "" {
			yamlBuilder.WriteString(fmt.Sprintf("%s%s%s: %s\n", commentPrefix, indentPrefix,
				yamlTag, valueStr))
		} else {
			yamlBuilder.WriteString(fmt.Sprintf("%s%s%s: %s # %s\n", commentPrefix, indentPrefix,
				yamlTag, valueStr, commentTag))
		}
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
