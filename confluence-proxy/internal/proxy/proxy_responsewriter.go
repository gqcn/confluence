package proxy

import (
	"bufio"
	"bytes"
	"net"
	"net/http"
)

// ResponseWriter is the custom writer for http response.
type ResponseWriter struct {
	status      int                 // HTTP status.
	writer      http.ResponseWriter // The underlying ResponseWriter.
	buffer      *bytes.Buffer       // The output buffer.
	hijacked    bool                // Mark this request is hijacked or not.
	wroteHeader bool                // Is header wrote or not, avoiding error: superfluous/multiple response.WriteHeader call.
}

// NewResponseWriter creates and return a ResponseWriter.
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		buffer: bytes.NewBuffer(nil),
		writer: w,
	}
}

// RawWriter returns the underlying ResponseWriter.
func (w *ResponseWriter) RawWriter() http.ResponseWriter {
	return w.writer
}

// Status returns the status of ResponseWriter.
func (w *ResponseWriter) Status() int {
	return w.status
}

// Header implements the interface function of http.ResponseWriter.Header.
func (w *ResponseWriter) Header() http.Header {
	return w.writer.Header()
}

// Write implements the interface function of http.ResponseWriter.Write.
func (w *ResponseWriter) Write(data []byte) (int, error) {
	w.buffer.Write(data)
	return len(data), nil
}

// WriteString appends the contents of s to the buffer, growing the buffer as
// needed. The return value n is the length of s; err is always nil. If the
// buffer becomes too large, WriteString will panic with ErrTooLarge.
func (w *ResponseWriter) WriteString(s string) (n int, err error) {
	return w.buffer.WriteString(s)
}

func (w *ResponseWriter) Overwrite(s []byte) (n int, err error) {
	w.buffer.Reset()
	return w.buffer.Write(s)
}

func (w *ResponseWriter) OverwriteString(s string) (n int, err error) {
	w.buffer.Reset()
	return w.buffer.WriteString(s)
}

// WriteHeader implements the interface of http.ResponseWriter.WriteHeader.
func (w *ResponseWriter) WriteHeader(status int) {
	w.status = status
}

// Hijack implements the interface function of http.Hijacker.Hijack.
func (w *ResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	w.hijacked = true
	return w.writer.(http.Hijacker).Hijack()
}

// Buffer returns the buffered content as []byte.
func (w *ResponseWriter) Buffer() []byte {
	return w.buffer.Bytes()
}

func (w *ResponseWriter) Clear() {
	w.buffer.Reset()
}

// BufferString returns the buffered content as string.
func (w *ResponseWriter) BufferString() string {
	return w.buffer.String()
}

// OutputBuffer outputs the buffer to client and clears the buffer.
func (w *ResponseWriter) OutputBuffer() {
	if w.hijacked {
		return
	}
	if w.status != 0 && !w.wroteHeader {
		w.writer.WriteHeader(w.status)
	}
	// Default status text output.
	if w.status != http.StatusOK && w.buffer.Len() == 0 {
		w.buffer.WriteString(http.StatusText(w.status))
	}
	if w.buffer.Len() > 0 {
		_, _ = w.writer.Write(w.buffer.Bytes())
		w.buffer.Reset()
	}
}
