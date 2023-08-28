package types

import "encoding/json"

type Temperature interface {
	Celsius() Celsius
	Fahrenheit() Fahrenheit
}

type Fahrenheit float64

var _ json.Marshaler = Fahrenheit(0.0)

func (f Fahrenheit) Celsius() Celsius       { return Celsius((f - 32.0) * 5.0 / 9.0) }
func (f Fahrenheit) Fahrenheit() Fahrenheit { return f }

func (f Fahrenheit) MarshalJSON() ([]byte, error) {
	return tempAsJSON(f)
}

type Celsius float64

var _ json.Marshaler = Celsius(0.0)

func (c Celsius) Celsius() Celsius       { return c }
func (c Celsius) Fahrenheit() Fahrenheit { return Fahrenheit(32.0 + c*9.0/5.0) }

func (c Celsius) MarshalJSON() ([]byte, error) {
	return tempAsJSON(c)
}

func tempAsJSON(t Temperature) ([]byte, error) {
	type temp struct {
		F float64 `json:"f"`
		C float64 `json:"c"`
	}
	return json.Marshal(temp{F: float64(t.Fahrenheit()), C: float64(t.Celsius())})
}
