package snapshots

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/runtime/ioutils"
	"github.com/iotaledger/wasp/v2/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/v2/packages/util/rwutil"
)

const downloaderServerPathConst = "testDownloader"

// TODO: test reading without chunks. How to create a file server which pretends to not support `Range` header?

func TestDownloaderSimple(t *testing.T) {
	uintToWrite := uint32(123456)
	stringToWrite := "Lorem ipsum"
	bytesToWrite := []byte{1, 2, 3, 4, 5, 6}

	testDownloader(t,
		func(w io.Writer) {
			ww := rwutil.NewWriter(w)
			ww.WriteUint32(uintToWrite)
			ww.WriteString(stringToWrite)
			ww.WriteN(bytesToWrite)
			require.NoError(t, ww.Err)
		},
		func(r io.Reader) {
			rr := rwutil.NewReader(r)
			require.Equal(t, uintToWrite, rr.ReadUint32())
			require.Equal(t, stringToWrite, rr.ReadString())
			bytes := make([]byte, len(bytesToWrite))
			rr.ReadN(bytes)
			require.Equal(t, bytesToWrite, bytes)
			require.NoError(t, rr.Err)
		},
	)
}

func TestDownloaderPartialRead(t *testing.T) {
	uintToWrite := uint32(123456)

	testDownloader(t,
		func(w io.Writer) {
			ww := rwutil.NewWriter(w)
			ww.WriteUint32(uintToWrite)
			ww.WriteString("SomeLongStringToFillSeveralSmallChunks")
			ww.WriteN([]byte{2, 2, 3, 3, 2, 2})
			require.NoError(t, ww.Err)
		},
		func(r io.Reader) {
			rr := rwutil.NewReader(r)
			require.Equal(t, uintToWrite, rr.ReadUint32())
			require.NoError(t, rr.Err)
		},
		10,
	)
}

func TestDownloaderManyChunksInOneRead(t *testing.T) {
	bytesToWrite1 := make([]byte, 5)
	bytesToWrite2 := make([]byte, 10)
	bytesToWrite3 := make([]byte, 50)
	bytesToWrite4 := make([]byte, 5)
	for i := range bytesToWrite1 {
		bytesToWrite1[i] = byte(i) + 1
	}
	for i := range bytesToWrite2 {
		bytesToWrite2[i] = byte(i) + 10
	}
	for i := range bytesToWrite3 {
		bytesToWrite3[i] = byte(i) + 100
	}
	for i := range bytesToWrite4 {
		bytesToWrite4[i] = byte(i) + 200
	}

	testDownloader(t,
		func(w io.Writer) {
			writeFun := func(b []byte) {
				n, err := w.Write(b)
				require.Equal(t, len(b), n)
				require.NoError(t, err)
			}
			writeFun(bytesToWrite1)
			writeFun(bytesToWrite2)
			writeFun(bytesToWrite3)
			writeFun(bytesToWrite4)
		},
		func(r io.Reader) {
			readFun := func(b []byte) {
				buf := make([]byte, len(b))
				n, err := io.ReadFull(r, buf)
				require.Equal(t, len(b), n)
				require.NoError(t, err)
				require.Equal(t, b, buf)
			}
			readFun(bytesToWrite1)
			readFun(bytesToWrite2)
			readFun(bytesToWrite3)
			readFun(bytesToWrite4)
		},
		10,
	)
}

func TestDownloaderReadTooMuch(t *testing.T) {
	arraySize := 25
	bytesToWrite := make([]byte, arraySize)
	for i := range bytesToWrite {
		bytesToWrite[i] = byte(i) + 1
	}

	testDownloader(t,
		func(w io.Writer) {
			n, err := w.Write(bytesToWrite)
			require.Equal(t, len(bytesToWrite), n)
			require.NoError(t, err)
		},
		func(r io.Reader) {
			buf := make([]byte, arraySize+5)
			n, err := r.Read(buf)
			require.Equal(t, arraySize, n)
			require.Error(t, err)
			require.Equal(t, io.EOF, err)
		},
		10,
	)
}

func startServer(t *testing.T, port string, handler http.Handler) {
	listener, err := net.Listen("tcp", port)
	require.NoError(t, err)
	srv := &http.Server{Addr: port, Handler: handler}
	go func() {
		err := srv.Serve(listener)
		if !errors.Is(err, http.ErrServerClosed) {
			require.NoError(t, err)
		}
	}()
	t.Cleanup(func() { srv.Shutdown(context.Background()) })
}

func testDownloader(t *testing.T, writeFun func(io.Writer), readFun func(io.Reader), chunkSize ...uint64) {
	log := testlogger.NewLogger(t)
	defer log.Shutdown()
	defer cleanupAfterDownloaderTest(t)

	err := ioutils.CreateDirectory(downloaderServerPathConst, 0o777)
	require.NoError(t, err)

	port := ":9998"
	startServer(t, port, http.FileServer(http.Dir(downloaderServerPathConst)))

	fileName := "TestFile.bin"
	filePathLocal := filepath.Join(downloaderServerPathConst, fileName)
	f, err := os.OpenFile(filePathLocal, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o666)
	require.NoError(t, err)
	writeFun(f)
	err = f.Close()
	require.NoError(t, err)

	filePathURL, err := url.JoinPath("http://localhost"+port+"/", fileName)
	require.NoError(t, err)
	d, err := NewDownloader(context.Background(), filePathURL, chunkSize...)
	require.NoError(t, err)
	readFun(d)
	err = d.Close()
	require.NoError(t, err)

	// NOTE: if downloading by chunks is not supported, chunkSize is going to be 0
	if len(chunkSize) > 0 {
		require.Equal(t, chunkSize[0], d.(*downloaderImpl).chunkSize)
	} else {
		require.Equal(t, defaultChunkSizeConst, d.(*downloaderImpl).chunkSize)
	}
}

func cleanupAfterDownloaderTest(t *testing.T) {
	err := os.RemoveAll(downloaderServerPathConst)
	require.NoError(t, err)
}
