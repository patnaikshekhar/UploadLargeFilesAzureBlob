package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/Azure/azure-storage-blob-go/azblob"
)

const containerName = "bigfiles"

func getCredentials() (string, *azblob.SharedKeyCredential, *azblob.ContainerURL, error) {
	accountName := os.Getenv("AZ_ACCOUNT_NAME")
	accountKey := os.Getenv("AZ_ACCOUNT_KEY")

	credential, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		return "", nil, nil, err
	}

	p := azblob.NewPipeline(credential, azblob.PipelineOptions{})

	URL, _ := url.Parse(
		fmt.Sprintf("https://%s.blob.core.windows.net/%s", accountName, containerName))

	containerURL := azblob.NewContainerURL(*URL, p)

	return accountName, credential, &containerURL, nil
}

func uploadToStorageBlob(fileName string, contents []byte) error {

	_, _, containerURL, err := getCredentials()
	if err != nil {
		return err
	}

	ctx := context.Background()

	blobURL := containerURL.NewBlockBlobURL(fileName)

	_, err = azblob.UploadBufferToBlockBlob(ctx, contents, blobURL, azblob.UploadToBlockBlobOptions{
		BlockSize:   4 * 1024 * 1024,
		Parallelism: 16,
	})

	return err
}

func listBlobs() ([]azblob.BlobItem, error) {

	ctx := context.Background()

	_, _, containerURL, err := getCredentials()
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

	return result, nil
}

func getSAS(filename string) string {
	accountName, credential, _, err := getCredentials()
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
