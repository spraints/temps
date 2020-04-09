package temps

type sensorSlice []*sensor

func (s sensorSlice) Len() int {
	return len(s)
}

func (s sensorSlice) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}

func (s sensorSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
