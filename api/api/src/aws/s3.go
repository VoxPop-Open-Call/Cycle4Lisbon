package aws

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	signer "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	userFilesPrefix  = "user-files/"
	profilePicSuffix = "/profilepic"

	initiativeFilesPrefix = "initiative-files/"
	initiativeImgSuffix   = "/banner-img"

	institutionFilesPrefix = "institution-files/"
	institutionLogoSuffix  = "/logo"

	presignedUrlExpiration = 20 * time.Minute
)

// S3 embeds both s3.Client and s3.PresignClient, and implements helper methods
// to work with objects in the user files bucket.
type S3 struct {
	*s3.Client
	*s3.PresignClient

	bucketName string
}

func newS3(config aws.Config, bucketName string) *S3 {
	client := s3.NewFromConfig(config)
	presignClient := s3.NewPresignClient(client)
	return &S3{
		Client:        client,
		PresignClient: presignClient,
		bucketName:    bucketName,
	}
}

func presignExpiration(opts *s3.PresignOptions) {
	opts.Expires = presignedUrlExpiration
}

// PresignGet generates a pre-signed request to retrieve an object from the
// user files bucket.
func (c *S3) PresignGet(key string) (*signer.PresignedHTTPRequest, error) {
	return c.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: &c.bucketName,
		Key:    &key,
	}, presignExpiration)
}

// PresignPut generates a pre-signed request to put an object in the user files
// bucket.
func (c *S3) PresignPut(key string) (*signer.PresignedHTTPRequest, error) {
	return c.PresignPutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: &c.bucketName,
		Key:    &key,
	}, presignExpiration)
}

// PresignDelete generates a pre-signed request to delete an object from the
// user files bucket.
func (c *S3) PresignDelete(key string) (*signer.PresignedHTTPRequest, error) {
	return c.PresignDeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: &c.bucketName,
		Key:    &key,
	}, presignExpiration)
}

func wrapPresign(
	key string,
	f func(string) (*signer.PresignedHTTPRequest, error),
) (url, method string, err error) {
	req, err := f(key)
	if err != nil { // avoid nil pointer dereference
		return "", "", err
	}
	return req.URL, req.Method, err
}

func profilePicKey(userID string) string {
	return userFilesPrefix + userID + profilePicSuffix
}

func wrapProfilePicPresign(
	userID string,
	f func(string) (*signer.PresignedHTTPRequest, error),
) (url, method string, err error) {
	return wrapPresign(profilePicKey(userID), f)
}

func (c *S3) PresignGetProfilePicture(
	userID string,
) (url, method string, err error) {
	return wrapProfilePicPresign(userID, c.PresignGet)
}

func (c *S3) PresignPutProfilePicture(
	userID string,
) (url, method string, err error) {
	return wrapProfilePicPresign(userID, c.PresignPut)
}

func (c *S3) PresignDeleteProfilePicture(
	userID string,
) (url, method string, err error) {
	return wrapProfilePicPresign(userID, c.PresignDelete)
}

func initiativeImgKey(initiativeID string) string {
	return initiativeFilesPrefix + initiativeID + initiativeImgSuffix
}

func wrapInitiativeImgPresign(
	initiativeID string,
	f func(string) (*signer.PresignedHTTPRequest, error),
) (url, method string, err error) {
	return wrapPresign(initiativeImgKey(initiativeID), f)
}

func (c *S3) PresignGetInitiativeImg(
	initiativeID string,
) (url, method string, err error) {
	return wrapInitiativeImgPresign(initiativeID, c.PresignGet)
}

func (c *S3) PresignPutInitiativeImg(
	initiativeID string,
) (url, method string, err error) {
	return wrapInitiativeImgPresign(initiativeID, c.PresignPut)
}

func (c *S3) PresignDeleteInitiativeImg(
	initiativeID string,
) (url, method string, err error) {
	return wrapInitiativeImgPresign(initiativeID, c.PresignDelete)
}

func institutionLogoKey(institutionID string) string {
	return institutionFilesPrefix + institutionID + institutionLogoSuffix
}

func wrapInstitutionLogoPresign(
	institutionID string,
	f func(string) (*signer.PresignedHTTPRequest, error),
) (url, method string, err error) {
	return wrapPresign(institutionLogoKey(institutionID), f)
}

func (c *S3) PresignGetInstitutionLogo(
	institutionID string,
) (url, method string, err error) {
	return wrapInstitutionLogoPresign(institutionID, c.PresignGet)
}

func (c *S3) PresignPutInstitutionLogo(
	institutionID string,
) (url, method string, err error) {
	return wrapInstitutionLogoPresign(institutionID, c.PresignPut)
}

func (c *S3) PresignDeleteInstitutionLogo(
	institutionID string,
) (url, method string, err error) {
	return wrapInstitutionLogoPresign(institutionID, c.PresignDelete)
}
