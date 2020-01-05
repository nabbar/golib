package njs_ioutils

import "io"

type IOWrapper struct {
	iow 	interface{}
	read func(p []byte) []byte
	write func(p []byte) []byte
}

func NewIOWrapper(ioInput interface{}) *IOWrapper {
	return &IOWrapper{
		iow: ioInput,
	}
}

func (I *IOWrapper) SetWrapper(read func(p []byte) []byte, write func(p []byte) []byte) {
	if read != nil {
		I.read = read
	}
	if write != nil {
		I.write = write
	}
}

func (I IOWrapper) Read(p []byte) (n int, err error) {
	if I.read != nil {
		return I.iow.(io.Reader).Read(I.read(p))
	}

	return I.iow.(io.Reader).Read(p)
}

func (I IOWrapper) Write(p []byte) (n int, err error) {
	if I.write != nil {
		return I.iow.(io.Writer).Write(I.write(p))
	}

	return I.iow.(io.Writer).Write(p)
}

func (I IOWrapper) Seek(offset int64, whence int) (int64, error) {
	return I.iow.(io.Seeker).Seek(offset, whence)
}

func (I IOWrapper) Close() error {
	return I.iow.(io.Closer).Close()
}



