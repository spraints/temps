package types

type Temperature interface {
	Celsius() Celsius
	Fahrenheit() Fahrenheit
}

type Fahrenheit float64

func (f Fahrenheit) Celsius() Celsius       { return Celsius((f - 32.0) * 5.0 / 9.0) }
func (f Fahrenheit) Fahrenheit() Fahrenheit { return f }

type Celsius float64

func (c Celsius) Celsius() Celsius       { return c }
func (c Celsius) Fahrenheit() Fahrenheit { return Fahrenheit(32.0 + c*9.0/5.0) }
