package bn

import "math/big"

func convertIntNumberToFloat(x *IntNumber) *FloatNumber {
	return &FloatNumber{x: x.BigFloat()}
}

func convertBigIntToFloat(x *big.Int) *FloatNumber {
	return &FloatNumber{x: new(big.Float).SetInt(x)}
}

func convertBigFloatToFloat(x *big.Float) *FloatNumber {
	return &FloatNumber{x: x}
}

func convertInt64ToFloat(x int64) *FloatNumber {
	return &FloatNumber{x: new(big.Float).SetInt64(x)}
}

func convertUint64ToFloat(x uint64) *FloatNumber {
	return &FloatNumber{x: new(big.Float).SetUint64(x)}
}

func convertFloat64ToFloat(x float64) *FloatNumber {
	return &FloatNumber{x: big.NewFloat(x)}
}

func convertStringToFloat(x string) *FloatNumber {
	if f, ok := new(big.Float).SetString(x); ok {
		return &FloatNumber{x: f}
	}
	return nil
}

func convertFloatNumberToInt(x *FloatNumber) *IntNumber {
	return &IntNumber{x: x.BigInt()}
}

func convertBigIntToInt(x *big.Int) *IntNumber {
	return &IntNumber{x: x}
}

func convertBigFloatToInt(x *big.Float) *IntNumber {
	i, _ := x.Int(nil)
	return &IntNumber{x: i}
}

func convertInt64ToInt(x int64) *IntNumber {
	return &IntNumber{x: new(big.Int).SetInt64(x)}
}

func convertUint64ToInt(x uint64) *IntNumber {
	return &IntNumber{x: new(big.Int).SetUint64(x)}
}

func convertFloat64ToInt(x float64) *IntNumber {
	f, _ := big.NewFloat(x).Int(nil)
	return &IntNumber{x: f}
}

func convertStringToInt(x string) *IntNumber {
	if i, ok := new(big.Int).SetString(x, 0); ok {
		return &IntNumber{x: i}
	}
	return nil
}

func convertBytesToInt(x []byte) *IntNumber {
	return &IntNumber{x: new(big.Int).SetBytes(x)}
}

func anyToInt64(x any) int64 {
	switch x := x.(type) {
	case int:
		return int64(x)
	case int8:
		return int64(x)
	case int16:
		return int64(x)
	case int32:
		return int64(x)
	case int64:
		return x
	}
	return 0
}

func anyToUint64(x any) uint64 {
	switch x := x.(type) {
	case uint:
		return uint64(x)
	case uint8:
		return uint64(x)
	case uint16:
		return uint64(x)
	case uint32:
		return uint64(x)
	case uint64:
		return x
	}
	return 0
}

func anyToFloat64(x any) float64 {
	switch x := x.(type) {
	case float32:
		return float64(x)
	case float64:
		return x
	}
	return 0
}
