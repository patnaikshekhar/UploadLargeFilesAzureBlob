package main

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/Azure/azure-storage-blob-go/azblob"
)

var progress sync.Map

func getCredentials() (string, string, *azblob.SharedKeyCredential, *azblob.ContainerURL, error) {
	accountName := os.Getenv("AZ_ACCOUNT_NAME")
	accountKey := os.Getenv("AZ_ACCOUNT_KEY")
	containerName := os.Getenv("AZ_CONTAINER_NAME")

	credential, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		return "", "", nil, nil, err
	}

	p := azblob.NewPipeline(credential, azblob.PipelineOptions{})

	URL, _ := url.Parse(
		fmt.Sprintf("https://%s.blob.core.windows.net/%s", accountName, containerName))

	containerURL := azblob.NewContainerURL(*URL, p)

	return accountName, containerName, credential, &containerURL, nil
}

func uploadToStorageBlob(fileName string, file multipart.File) error {

	log.Printf("Starting to upload file %s", fileName)

	progress.Store(fileName, 0)

	_, _, _, containerURL, err := getCredentials()
	if err != nil {
		return err
	}

	ctx := context.Background()

	blobURL := containerURL.NewBlockBlobURL(fileName)

	bufferSize := 2 * 1024 * 1024
	maxBuffers := 3

	_, err = azblob.UploadStreamToBlockBlob(ctx, file, blobURL, azblob.UploadStreamToBlockBlobOptions{
		BufferSize: bufferSize,
		MaxBuffers: maxBuffers,
	})

	// _, err = azblob.UploadBufferToBlockBlob(ctx, contents, blobURL, azblob.UploadToBlockBlobOptions{
	// 	BlockSize:   4 * 1024 * 1024,
	// 	Parallelism: 16,
	// 	Progress: func(bytesTransferred int64) {
	// 		percentage := float64(bytesTransferred) / float64(len(contents)) * 100
	// 		progress.Store(fileName, int(percentage))
	// 	},
	// })

	progress.Delete(fileName)

	log.Printf("Completed uploading. Errors are %v", err)

	return err
}

func listBlobs() ([]azblob.BlobItem, error) {

	ctx := context.Background()

	_, _, _, containerURL, err := getCredentials()
	if err != nil {
		return nil, err
	}

	var result []azblob.BlobItem

	for marker := (azblob.Marker{}); marker.NotDone(); {
		listBlob, err := containerURL.ListBlobsFlatSegment(ctx, marker, azblob.ListBlobsSegmentOptions{})
		if err != nil {
			return nil, err
		}

		marker = listBlob.NextMarker

		// Process the blobs returned in this result segment (if the segment is empty, the loop body won't execute)
		for _, blobInfo := range listBlob.Segment.BlobItems {
			result = append(result, blobInfo)
		}
	}

	log.Printf("Completed listBlobs. Found %d blobs.", len(result))

	return result, nil
}

func getSAS(filename string) string {
	accountName, containerName, credential, _, err := getCredentials()
	if err != nil {
		log.Println(err.Error())
	}

	sasQueryParams, err := azblob.BlobSASSignatureValues{
		Protocol:      azblob.SASProtocolHTTPS,              // Users MUST use HTTPS (not HTTP)
		ExpiryTime:    time.Now().UTC().Add(48 * time.Hour), // 48-hours before expiration
		ContainerName: containerName,
		BlobName:      filename,

		Permissions: azblob.BlobSASPermissions{Add: true, Read: true, Write: true}.String(),
	}.NewSASQueryParameters(credential)

	if err != nil {
		log.Fatal(err)
	}

	qp := sasQueryParams.Encode()

	url := fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s?%s",
		accountName, containerName, filename, qp)

	return url
}
