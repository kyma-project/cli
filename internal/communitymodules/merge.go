package communitymodules

func MergeRowMaps(moduleMaps ...moduleMap) moduleMap {
	result := make(moduleMap)
	for _, moduleMap := range moduleMaps {
		for name, value := range moduleMap {
			if resultValue, ok := result[name]; ok {
				result[name] = mergeTwoRows(resultValue, value)
			} else {
				result[name] = value
			}
		}
	}
	return result
}

func mergeTwoRows(a row, b row) row {
	result := a
	if result.Name == "" {
		result.Name = b.Name
	}
	if result.Repository == "" {
		result.Repository = b.Repository
	}
	if result.LatestVersion == "" {
		result.LatestVersion = b.LatestVersion
	}
	if result.Version == "" {
		result.Version = b.Version
	}
	if result.Channel == "" {
		result.Channel = b.Channel
	}
	return result
}
