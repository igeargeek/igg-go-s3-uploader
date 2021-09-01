# IGG-GO-S3-UPLOADER

## Requirement
- ชื่อ Module คือ ```github.com/igeargeek/igg-go-s3-uploader```
- Flexible ต่อการใช้งานและใช้งานได้ง่าย
- สามารถกำหนด Config หรือ Connection ได้ เช่น ACCESS_KEY, SECRET_KEY, BUCKET และอื่นๆ
- รองรับการอัพโหลด Image
- Lib สามารถ Auto compression รูปได้ และระบุ Quality ได้ และสามารถกำหนดการ Resize และ Crop ของรูปได้ (เหมือนกับการทำงานของ Shape ที่เราทำ Javascript ก่อนหน้านี้)
- มี Unit test
- อัพเป็น tag version เช่น v1.0.0

## Install
```
go get github.com/igeargeek/igg-go-s3-uploader/uploader
```

## Example
### Config
```go
accessKeyID := ""
secretAccessKey := ""
region := ""
bucket := ""

s3 := uploader.New(accessKeyID, secretAccessKey, region, bucket)
```

### UploadImage
```go
resp, err := s3.UploadFile(c, "imageFile", "test")
```

### UploadImage
```go
crop := [4]int{300, 300, 299, 0}
sizeOf := map[string]uploader.SizeOfWidthHeight{
  "tiny": {
  	Width:  100,
  	Height: 100,
  },
  "small": {
  	Width:  200,
  	Height: 200,
  },
  "medium": {
  	Width:  300,
  	Height: 300,
  },
  "large": {
  	Width:  400,
  	Height: 400,
  },
}

resp, err := s3.UploadImage(c, "imageFile", "product", crop, sizeOf, 75)
```

### DeleteFile
```go
_, err := s3.DeleteFile("5a84e813-8e37-40b6-8687-f79fe1fffa83.jpeg", "test")
```

### DeleteImage
```go
sizeOf := map[string]uploader.SizeOfWidthHeight{
  "tiny": {
  	Width:  100,
  	Height: 100,
  },
  "small": {
  	Width:  200,
  	Height: 200,
  },
  "medium": {
  	Width:  300,
  	Height: 300,
  },
  "large": {
  	Width:  400,
  	Height: 400,
  },
}

_, err := s3.DeleteImage("4392e491-9429-4859-85cf-11a5af90cb37.jpeg", "product", sizeOf)
```

### Example with fiber
```go
package controller

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/igeargeek/igg-go-s3-uploader/uploader"
)

type uploadController interface {
	UploadPhoto(c *fiber.Ctx) error
	DeleteFile(c *fiber.Ctx) error
}

type UploadController struct{}

type reSize struct {
	width  int
	height int
}

func NewUpload() uploadController {
	return &UploadController{}
}

func (u *UploadController) UploadPhoto(c *fiber.Ctx) error {
	accessKeyID := ""
	secretAccessKey := ""
	region := ""
	bucket := ""

	s3 := uploader.New(accessKeyID, secretAccessKey, region, bucket)

	crop := [4]int{300, 300, 299, 0}
	sizeOf := map[string]uploader.SizeOfWidthHeight{
		"tiny": {
			Width:  100,
			Height: 100,
		},
		"small": {
			Width:  200,
			Height: 200,
		},
		"medium": {
			Width:  300,
			Height: 300,
		},
		"large": {
			Width:  400,
			Height: 400,
		},
	}

	resp, err := s3.UploadImage(c, "imageFile", "product", crop, sizeOf, 75)

	if err != nil {
		fmt.Println(err)
		return c.Status(200).JSON("Error")
	}
	return c.Status(200).JSON(resp)
}
```