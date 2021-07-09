package main

import (
	"fmt"
	"igg-go-s3-uploader/uploader"
)

func main() {
	fmt.Println("run 1")

	uploader := uploader.Uploader{
		Bucket: "bucket",
		Key:    "key",
	}
	uploader.Print()
}
