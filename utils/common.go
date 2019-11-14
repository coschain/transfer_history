package utils

func CheckIsNotEmptyStr(str string) bool {
	if str != "" && len(str) > 0 {
		return true
	}
	return false
}