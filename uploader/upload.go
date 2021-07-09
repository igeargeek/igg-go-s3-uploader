package uploader

import "fmt"

type Uploader struct {
	Bucket string
	Key    string
}

func (i Uploader) Print() {
	fmt.Printf("%+v", i)
}
