package sm_snapshots

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type downloaderImpl struct {
	ctx         context.Context
	chunkReader io.ReadCloser
	filePath    string
	fileSize    uint64
	chunkEnd    uint64
	chunkSize   uint64
	onCloseFun  func()
}

var (
	_ io.Reader  = &downloaderImpl{}
	_ io.Closer  = &downloaderImpl{}
	_ Downloader = &downloaderImpl{}
)

const (
	defaultChunkSizeConst = uint64(1024 * 1024) // 1Mb
	tempFileSuffixConst   = ".part"
)

// Downloader is a reader, which reads from network URL in chunks (if webserver
// supports that). Only request to read the first chunk is made on downloader
// creation. Requests to read other chunks are made once they are needed by
// `Read` calls. The user of downloader can use `Read` independently of how many
// chunks will have to be downloaded to complete the operation.
func NewDownloader(
	ctx context.Context,
	filePath string,
	chunkSize ...uint64,
) (Downloader, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodHead, filePath, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to make head request to %s: %w", filePath, err)
	}
	head, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to receive header for url %s: %w", filePath, err)
	}
	defer head.Body.Close()

	if head.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("head request to %s got status code %v", filePath, head.StatusCode)
	}

	acceptRanges := head.Header.Get("Accept-Ranges")
	fileSizeStr := head.Header.Get("Content-Length")
	fileSize, err := strconv.ParseUint(fileSizeStr, 10, 64)
	result := &downloaderImpl{
		ctx:        ctx,
		filePath:   filePath,
		fileSize:   fileSize,
		onCloseFun: func() {},
	}
	if err != nil {
		return nil, fmt.Errorf("failed to convert file length %v to integer: %w", fileSizeStr, err)
	}
	if acceptRanges == "" || strings.ToLower(acceptRanges) == "none" {
		result.chunkSize = 0
	} else {
		if len(chunkSize) > 0 {
			result.chunkSize = chunkSize[0]
		} else {
			result.chunkSize = defaultChunkSizeConst
		}
		result.chunkEnd = 0
	}
	err = result.setReader()
	if err != nil {
		if result.chunkReader != nil {
			result.chunkReader.Close()
		}
		return nil, err
	}
	return result, nil
}

func NewDownloaderWithTimeout(ctx context.Context,
	filePath string,
	timeout time.Duration,
	chunkSize ...uint64,
) (Downloader, error) {
	ctxWithTimeout, ctxWithTimeoutCancel := context.WithTimeout(ctx, timeout)
	result, err := NewDownloader(ctxWithTimeout, filePath, chunkSize...)
	if err != nil {
		ctxWithTimeoutCancel()
		return nil, err
	}
	result.(*downloaderImpl).onCloseFun = ctxWithTimeoutCancel
	return result, nil
}

func (d *downloaderImpl) setReader() error {
	request, err := http.NewRequestWithContext(d.ctx, http.MethodGet, d.filePath, http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to make get request to %s: %w", d.filePath, err)
	}
	chunkPartStr := ""
	var expectedStatusCode int
	if d.chunkSize > 0 {
		start := d.chunkEnd
		end := start + d.chunkSize
		if end > d.fileSize {
			end = d.fileSize
		}
		request.Header.Add("Range", "bytes="+strconv.FormatUint(start, 10)+"-"+strconv.FormatUint(end-1, 10))
		chunkPartStr = fmt.Sprintf(" byte %v to %v", start, end)
		d.chunkEnd = end
		expectedStatusCode = http.StatusPartialContent
	} else {
		d.chunkEnd = d.fileSize
		expectedStatusCode = http.StatusOK
	}
	chunk, err := http.DefaultClient.Do(request) //nolint:bodyclose// closing is handled differently; linter cannot understand that
	if err != nil {
		return fmt.Errorf("failed to get file%s from %s: %w", chunkPartStr, d.filePath, err)
	}
	d.chunkReader = chunk.Body
	if chunk.StatusCode != expectedStatusCode {
		return fmt.Errorf("get%s request to %s got status code %v", chunkPartStr, d.filePath, chunk.StatusCode)
	}
	return nil
}

func (d *downloaderImpl) Read(b []byte) (int, error) {
	n, err := d.chunkReader.Read(b)
	if err == io.EOF {
		if d.chunkEnd >= d.fileSize {
			return n, err
		}
		d.chunkReader.Close()
		err = d.setReader()
		if err != nil {
			return n, err
		}
		var nn int
		nn, err = d.Read(b[n:])
		return n + nn, err
	}
	return n, err
}

func (d *downloaderImpl) Close() error {
	d.onCloseFun()
	return d.chunkReader.Close()
}

func (d *downloaderImpl) GetLength() uint64 {
	return d.fileSize
}

// Use downloader to download data to file. Temporary file is created while
// download is in progress and only on finishing the download, the file is
// renamed to provided name. Progress reporter wrapped in `TeeReader` (or any
// other reader) can be wrapped around downloader (provided by parameter of the
// function). If it is not needed, `addProgressReporter` function should return
// the reader in the parameter.
func DownloadToFile(
	ctx context.Context,
	filePathNetwork string,
	filePathLocal string,
	timeout time.Duration,
	addProgressReporter func(io.Reader, string, uint64) io.Reader,
) error {
	filePathTemp := filePathLocal + tempFileSuffixConst
	err := func() error { // Function is used to make deferred close occur when it is needed even if write is successful
		downloader, e := NewDownloaderWithTimeout(ctx, filePathNetwork, timeout)
		if e != nil {
			return fmt.Errorf("failed to start downloading %s: %w", filePathNetwork, e)
		}
		defer downloader.Close()
		r := addProgressReporter(downloader, filePathNetwork, downloader.GetLength())

		f, e := os.OpenFile(filePathTemp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o666)
		if e != nil {
			return fmt.Errorf("failed to create temporary file %s: %w", filePathTemp, e)
		}
		defer f.Close()

		n, e := io.Copy(f, r)
		if e != nil {
			return fmt.Errorf("error downloading and saving url %s to file %s: %w", filePathNetwork, filePathTemp, e)
		}
		if n != int64(downloader.GetLength()) {
			return fmt.Errorf("downloaded file %s was not written completely: of %v bytes to download only %v byte written",
				filePathNetwork, downloader.GetLength(), n)
		}
		return nil
	}()
	if err != nil {
		return err
	}
	err = os.Rename(filePathTemp, filePathLocal)
	if err != nil {
		return fmt.Errorf("failed to move temporary file %s to permanent location %s: %v",
			filePathTemp, filePathLocal, err)
	}
	return nil
}
