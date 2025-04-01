package parameters

// this func makes sure that types used as an default follows YAML standards
// converts:
// int8,int16,int32,int64,int to int64
// string, []byte to string
func sanitizeDefaultValue(defaultValue interface{}) interface{} {
	switch value := defaultValue.(type) {
	case int8:
		return int64(value)
	case int16:
		return int64(value)
	case int32:
		return int64(value)
	case int64:
		return value
	case int:
		return int64(value)
	case bool:
		return value
	case string:
		return value
	case []byte:
		return string(value)
	default:
		return nil
	}
}
