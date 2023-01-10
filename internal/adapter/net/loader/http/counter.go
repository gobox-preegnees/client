package http

import (
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/sirupsen/logrus"
)

// WriteCounter counts the number of bytes written to it. By implementing the Write method,
// it is of the io.Writer interface and we can pass this into io.TeeReader()
// Every write to this writer, will print the progress of the file write.
type WriteCounter struct {
	Log          *logrus.Logger
	Total        uint64
	withProgress bool
}

func (ws *WriteCounter) WithProgress() *WriteCounter {

	ws.withProgress = true
	return ws
}

func (wc *WriteCounter) Write(p []byte) (int, error) {

	n := len(p)
	if !wc.withProgress {
		return n, nil
	}

	wc.Total += uint64(n)
	wc.PrintProgress()
	return n, nil
}

// PrintProgress prints the progress of a file write
func (wc *WriteCounter) PrintProgress() {

	if !wc.withProgress {
		return
	}
	// Clear the line by using a character return to go back to the start and remove
	// the remaining characters by filling it with spaces
	wc.Log.Infof("\r%s", strings.Repeat(" ", 50))

	// Return again and print current status of download
	// We use the humanize package to print the bytes in a meaningful way (e.g. 10 MB)
	wc.Log.Infof("\rDownloading... %s complete", humanize.Bytes(wc.Total))
}
