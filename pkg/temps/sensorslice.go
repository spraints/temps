package temps

import (
	"github.com/spraints/temps/pkg/types"
)

type sensorSlice []types.Measurement

func (s sensorSlice) Len() int {
	return len(s)
}

func (s sensorSlice) Less(i, j int) bool {
	if s[i].ID == outsideID {
		return true
	}
	if s[j].ID == outsideID {
		return false
	}
	return s[i].Name < s[j].Name
}

func (s sensorSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
