package main

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/Azure/azure-storage-blob-go/azblob"
)

const containerName = "bigfiles"

func getContainerURL() (*azblob.ContainerURL, error) {
	accountName := os.Getenv("AZ_ACCOUNT_NAME")
	accountKey := os.Getenv("AZ_ACCOUNT_KEY")

	credential, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		return nil, err
	}

	p := azblob.NewPipeline(credential, azblob.PipelineOptions{})

	URL, _ := url.Parse(
		fmt.Sprintf("https://%s.blob.core.windows.net/%s", accountName, containerName))

	containerURL := azblob.NewContainerURL(*URL, p)

	return &containerURL, nil
}
func uploadToStorageBlob(fileName string, contents []byte) error {

	containerURL, err := getContainerURL()
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

	containerURL, err := getContainerURL()
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
