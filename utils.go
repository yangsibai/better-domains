package main

func index(vs []string, t string) int {
	for i, v := range vs {
		if v == t {
			return i
		}
	}
	return -1
}

func filter(vs []string, f func(string) bool) []string {
	vsf := make([]string, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func contains(vs []string, t string) bool {
	return index(vs, t) != -1
}
