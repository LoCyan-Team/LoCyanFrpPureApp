package api

func BoolToString(val bool) (str string) {
	if val {
		return "true"
	}
	return "false"
}
