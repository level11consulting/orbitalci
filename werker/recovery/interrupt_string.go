// Code generated by "stringer -type Interrupt"; DO NOT EDIT.

package recovery

import "strconv"

const _Interrupt_name = "SignalPanic"

var _Interrupt_index = [...]uint8{0, 6, 11}

func (i Interrupt) String() string {
	if i < 0 || i >= Interrupt(len(_Interrupt_index)-1) {
		return "Interrupt(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Interrupt_name[_Interrupt_index[i]:_Interrupt_index[i+1]]
}
