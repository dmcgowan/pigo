package pigo

type Value byte
type PinFunction byte
type Direction PinFunction
type PinNumber uint
const (
	Low  = Value(0x00)
	High = Value(0x01)
	Off  = Low
	On   = High
	
	Input = Direction(0x00)
	Output = Direction(0x01)
	In = Input
	Out = Output
	Alt0   = PinFunction(0x04) // 100 GPIO Pin takes alternate function 0
	Alt1   = PinFunction(0x05) // 101 GPIO Pin takes alternate function 1
	Alt2   = PinFunction(0x06) // 110 GPIO Pin takes alternate function 2
	Alt3   = PinFunction(0x07) // 111 GPIO Pin takes alternate function 3
	Alt4   = PinFunction(0x03) // 011 GPIO Pin takes alternate function 4
	Alt5   = PinFunction(0x02) // 010 GPIO Pin takes alternate function 5
)

type GPIO interface {
	SetDirection(PinNumber, Direction)
	GetValue(PinNumber) Value
	SetValue(PinNumber, Value)
	Close() error
}

type FunctionalGPIO interface {
	GPIO
	SetFunction(PinNumber, PinFunction)
	GetFunction(PinNumber) PinFunction
}
