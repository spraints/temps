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

	if a, ok := s.getIntFromName(i); ok {
		if b, ok := s.getIntFromName(j); ok {
			return a < b
		}
	}

	return s[i].Name < s[j].Name
}

func (s sensorSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s sensorSlice) getIntFromName(i int) (n int, ok bool) {
	str := s[i].Name
	for j := 0; j < len(str); j++ {
		c := str[j]
		if c >= '0' && c <= '9' {
			n = 10*n + int(c-'0')
			ok = true
		} else {
			return
		}
	}
	return
}
