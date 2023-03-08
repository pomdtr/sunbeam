package utils

func SplitCommand(fields []string) (string, []string) {
	if len(fields) == 0 {
		return "", nil
	}

	if len(fields) == 1 {
		return fields[0], nil
	}

	return fields[0], fields[1:]

}
