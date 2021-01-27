package mail

import "io"

// file represents the files that can be added to the email message.
type File struct {
	name string
	mime string
	data io.ReadCloser
}

func NewFile(name string, mime string, data io.ReadCloser) File {
	return File{
		name: name,
		mime: mime,
		data: data,
	}
}
