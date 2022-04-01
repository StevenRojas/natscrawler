package crawler

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"

	"testing"
)

func TestDoCrawler(t *testing.T) {
	requestID := "123"
	url := "https://channelstore.roku.com/en-gb/details/517004dee775521b2464219f2d9dacb3/the-legal-channel"
	info, err := doCrawler(context.TODO(), url, requestID)
	assert.Nil(t, err)
	assert.Equal(t, requestID, info.RequestID)
	fmt.Printf("URL info: %+v\n", info)
}
