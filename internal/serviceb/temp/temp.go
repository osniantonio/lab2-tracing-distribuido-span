package temp

import "math"

type Temp struct {
	C float64
	F float64
	K float64
}

func roundFloat(val float64, precision uint8) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

func CToF(c float64) float64 {
	return roundFloat(c*9/5+32, 2)
}

func CToK(c float64) float64 {
	return roundFloat(c+273.15, 2)
}

func FToC(f float64) float64 {
	return roundFloat((f-32)*5/9, 2)
}

func FToK(f float64) float64 {
	return roundFloat((f-32)*5/9+273.15, 2)
}

func KToC(k float64) float64 {
	return roundFloat(k-273.15, 2)
}

func KToF(k float64) float64 {
	return roundFloat((k-273.15)*9/5+32, 2)
}
