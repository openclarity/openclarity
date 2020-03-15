package string

func TruncateString(str string, maxsize int) string {
	if len(str) > maxsize {
		return str[0:maxsize]
	}

	return str
}
