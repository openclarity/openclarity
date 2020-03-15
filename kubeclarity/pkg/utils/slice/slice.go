package slice

func AppendStringIfMissing(list []string, candidate string) []string {
	for _, ele := range list {
		if ele == candidate {
			return list
		}
	}
	list = append(list, candidate)
	return list
}

func ContainsString(list []string, imageName string) bool {
	for _, a := range list {
		if a == imageName {
			return true
		}
	}
	return false
}
