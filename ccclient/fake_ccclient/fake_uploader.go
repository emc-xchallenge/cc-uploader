// This file was generated by counterfeiter
package fake_ccclient

import (
	"net/http"
	"net/url"
	"sync"

	"github.com/cloudfoundry-incubator/cc-uploader/ccclient"
)

type FakeUploader struct {
	UploadStub        func(uploadURL *url.URL, filename string, r *http.Request, cancelChan <-chan struct{}) (*http.Response, error)
	uploadMutex       sync.RWMutex
	uploadArgsForCall []struct {
		uploadURL  *url.URL
		filename   string
		r          *http.Request
		cancelChan <-chan struct{}
	}
	uploadReturns struct {
		result1 *http.Response
		result2 error
	}
}

func (fake *FakeUploader) Upload(uploadURL *url.URL, filename string, r *http.Request, cancelChan <-chan struct{}) (*http.Response, error) {
	fake.uploadMutex.Lock()
	fake.uploadArgsForCall = append(fake.uploadArgsForCall, struct {
		uploadURL  *url.URL
		filename   string
		r          *http.Request
		cancelChan <-chan struct{}
	}{uploadURL, filename, r, cancelChan})
	fake.uploadMutex.Unlock()
	if fake.UploadStub != nil {
		return fake.UploadStub(uploadURL, filename, r, cancelChan)
	} else {
		return fake.uploadReturns.result1, fake.uploadReturns.result2
	}
}

func (fake *FakeUploader) UploadCallCount() int {
	fake.uploadMutex.RLock()
	defer fake.uploadMutex.RUnlock()
	return len(fake.uploadArgsForCall)
}

func (fake *FakeUploader) UploadArgsForCall(i int) (*url.URL, string, *http.Request, <-chan struct{}) {
	fake.uploadMutex.RLock()
	defer fake.uploadMutex.RUnlock()
	return fake.uploadArgsForCall[i].uploadURL, fake.uploadArgsForCall[i].filename, fake.uploadArgsForCall[i].r, fake.uploadArgsForCall[i].cancelChan
}

func (fake *FakeUploader) UploadReturns(result1 *http.Response, result2 error) {
	fake.UploadStub = nil
	fake.uploadReturns = struct {
		result1 *http.Response
		result2 error
	}{result1, result2}
}

var _ ccclient.Uploader = new(FakeUploader)
