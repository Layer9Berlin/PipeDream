package custom_io

import "io"

type BellSkipper struct{
	baseOutput io.WriteCloser
}

func NewBellSkipper(baseOutput io.WriteCloser) *BellSkipper {
	return &BellSkipper{baseOutput: baseOutput}
}

func (skipper *BellSkipper) Write(b []byte) (int, error) {
	const charBell = 7
	if len(b) == 1 && b[0] == charBell {
		return 0, nil
	}
	return skipper.baseOutput.Write(b)
}

func (skipper *BellSkipper) Close() error {
	return skipper.baseOutput.Close()
}
