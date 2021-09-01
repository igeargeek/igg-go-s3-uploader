package uploader

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/nfnt/resize"
	"github.com/oliamb/cutter"
	"github.com/spf13/cast"
)

type config struct {
	accessKeyID     string
	secretAccessKey string
	region          string
	bucket          string
}

type resultUploadImage struct {
	Filename string
}

type SizeOfWidthHeight struct {
	Width  int
	Height int
}

func New(accessKeyID string, secretAccessKey string, region string, bucket string) config {
	config := config{
		accessKeyID,
		secretAccessKey,
		region,
		bucket,
	}

	return config
}

func connectAws(accessKeyID string, secretAccessKey string, region string, bucket string) (*s3.S3, string, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
	})
	if err != nil {
		return nil, bucket, err
	}
	s3Client := s3.New(sess)
	return s3Client, bucket, nil
}

func (cf config) UploadFile(c *fiber.Ctx, inputField string, path string) (resultUploadImage, error) {
	fileHeader, errFileHeader := c.FormFile(inputField)
	if errFileHeader != nil {
		return resultUploadImage{}, errFileHeader
	}

	var pathInner = ""
	if path != "" {
		pathInner = path + "/"
	}

	splitType := strings.Split(fileHeader.Filename, ".")
	fileType := splitType[len(splitType)-1]
	filename := uuid.New().String() + "." + fileType
	fullPath := pathInner + filename
	contentType := fileHeader.Header["Content-Type"][0]

	file, _ := fileHeader.Open()
	buf := bytes.NewBuffer(nil)
	if _, errIoCopy := io.Copy(buf, file); errIoCopy != nil {
		return resultUploadImage{}, errIoCopy
	}

	s3Client, bucketName, _ := connectAws(cf.accessKeyID, cf.secretAccessKey, cf.region, cf.bucket)
	_, errUpload := s3Client.PutObject(&s3.PutObjectInput{
		Bucket:       aws.String(bucketName),
		Key:          aws.String(fullPath),
		Body:         bytes.NewReader(buf.Bytes()),
		ContentType:  aws.String(contentType),
		CacheControl: aws.String("max-age=31557600"),
		ACL:          aws.String("public-read"),
	})

	if errUpload != nil {
		return resultUploadImage{}, errUpload
	}

	return resultUploadImage{
		Filename: filename,
	}, nil
}

func (cf config) UploadImage(c *fiber.Ctx, inputField string, path string, crop [4]int, sizeOf map[string]SizeOfWidthHeight, quality int) (resultUploadImage, error) {
	fileHeader, errFileHeader := c.FormFile(inputField)
	if errFileHeader != nil {
		return resultUploadImage{}, errFileHeader
	}

	splitType := strings.Split(fileHeader.Filename, ".")
	fileType := splitType[len(splitType)-1]
	filename := uuid.New().String() + "." + fileType
	contentType := fileHeader.Header["Content-Type"][0]

	file, _ := fileHeader.Open()
	buf := bytes.NewBuffer(nil)
	if _, errIoCopy := io.Copy(buf, file); errIoCopy != nil {
		return resultUploadImage{}, errIoCopy
	}

	imageOriginalByteToImage, _, _ := image.Decode(bytes.NewReader(buf.Bytes()))

	g := imageOriginalByteToImage.Bounds()
	dyHeight := g.Dy()
	dxWidth := g.Dx()

	height := cast.ToInt(crop[1])
	width := cast.ToInt(crop[0])
	if width > dxWidth {
		width = dxWidth - cast.ToInt(crop[2])
	}
	if height > dyHeight {
		height = dyHeight - cast.ToInt(crop[3])
	}

	extractImage, errCrop := cutter.Crop(imageOriginalByteToImage, cutter.Config{
		Width:  width,
		Height: height,
		Anchor: image.Point{cast.ToInt(crop[2]), cast.ToInt(crop[3])},
		Mode:   cutter.TopLeft,
	})
	if errCrop != nil {
		return resultUploadImage{}, errCrop
	}

	var pathInner = ""
	if path != "" {
		pathInner = path + "/"
	}

	for key, size := range sizeOf {
		fullPath := pathInner + "/" + key + "/" + filename
		var width uint
		var height uint

		width = cast.ToUint(size.Width)
		height = cast.ToUint(size.Height)

		resizeImage := resize.Resize(width, height, extractImage, resize.Lanczos3)
		bufCrop := new(bytes.Buffer)

		if contentType == "image/jpeg" {
			var opt jpeg.Options
			opt.Quality = quality
			jpeg.Encode(bufCrop, resizeImage, &opt)
		} else if contentType == "image/png" {
			png.Encode(bufCrop, resizeImage)
		}

		s3Client, bucketName, _ := connectAws(cf.accessKeyID, cf.secretAccessKey, cf.region, cf.bucket)
		_, errUpload := s3Client.PutObject(&s3.PutObjectInput{
			Bucket:       aws.String(bucketName),
			Key:          aws.String(fullPath),
			Body:         bytes.NewReader(bufCrop.Bytes()),
			ContentType:  aws.String(contentType),
			CacheControl: aws.String("max-age=31557600"),
			ACL:          aws.String("public-read"),
		})

		if errUpload != nil {
			return resultUploadImage{}, errUpload
		}
	}

	return resultUploadImage{
		Filename: filename,
	}, nil
}

func (cf config) DeleteFile(filename string, path string) (bool, error) {
	var pathInner = ""
	if path != "" {
		pathInner = path + "/"
	}

	fullPath := pathInner + filename

	s3Client, bucketName, _ := connectAws(cf.accessKeyID, cf.secretAccessKey, cf.region, cf.bucket)
	_, errDelete := s3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fullPath),
	})

	if errDelete != nil {
		return false, errDelete
	}

	return true, nil
}

func (cf config) DeleteImage(filename string, path string, sizeOf map[string]SizeOfWidthHeight) (bool, error) {
	var pathInner = ""
	if path != "" {
		pathInner = path + "/"
	}

	for key, _ := range sizeOf {
		fullPath := pathInner + "/" + key + "/" + filename

		s3Client, bucketName, _ := connectAws(cf.accessKeyID, cf.secretAccessKey, cf.region, cf.bucket)
		_, errDelete := s3Client.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(fullPath),
		})

		if errDelete != nil {
			return false, errDelete
		}
	}

	return true, nil
}
