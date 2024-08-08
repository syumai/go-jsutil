package jsutil

import (
	"bytes"
	"fmt"
	"io"
	"syscall/js"
)

type RawJSBodyWriter interface {
	WriteRawJSBody(body js.Value)
}

type RawJSBodyGetter interface {
	GetRawJSBody() js.Value
}

// readableStreamToReadCloser implements io.Reader sourced from ReadableStreamDefaultReader.
//   - ReadableStreamDefaultReader: https://developer.mozilla.org/en-US/docs/Web/API/ReadableStreamDefaultReader
//   - This implementation is based on: https://deno.land/std@0.139.0/streams/conversion.ts#L76
type readableStreamToReadCloser struct {
	buf          bytes.Buffer
	stream       js.Value
	streamReader *js.Value
}

var (
	_ io.ReadCloser   = (*readableStreamToReadCloser)(nil)
	_ io.WriterTo     = (*readableStreamToReadCloser)(nil)
	_ RawJSBodyGetter = (*readableStreamToReadCloser)(nil)
)

// Read reads bytes from ReadableStreamDefaultReader.
func (sr *readableStreamToReadCloser) Read(p []byte) (n int, err error) {
	if sr.streamReader == nil {
		r := sr.stream.Call("getReader")
		sr.streamReader = &r
	}
	if sr.buf.Len() == 0 {
		resultCh := make(chan js.Value)
		errCh := make(chan error)
		promise := sr.streamReader.Call("read")
		var then, catch js.Func
		then = js.FuncOf(func(_ js.Value, args []js.Value) any {
			defer then.Release()
			result := args[0]
			if result.Get("done").Bool() {
				errCh <- io.EOF
				return js.Undefined()
			}
			resultCh <- result.Get("value")
			return js.Undefined()
		})
		catch = js.FuncOf(func(_ js.Value, args []js.Value) any {
			defer catch.Release()
			result := args[0]
			errCh <- fmt.Errorf("JavaScript error on read: %s", result.Call("toString").String())
			return js.Undefined()
		})
		promise.Call("then", then).Call("catch", catch)
		select {
		case result := <-resultCh:
			chunk := make([]byte, result.Get("byteLength").Int())
			_ = js.CopyBytesToGo(chunk, result)
			// The length written is always the same as the length of chunk, so it can be discarded.
			//   - https://pkg.go.dev/bytes#Buffer.Write
			_, err := sr.buf.Write(chunk)
			if err != nil {
				return 0, err
			}
		case err := <-errCh:
			return 0, err
		}
	}
	return sr.buf.Read(p)
}

func (sr *readableStreamToReadCloser) Close() error {
	if sr.streamReader == nil {
		return nil
	}
	sr.streamReader.Call("cancel")
	return nil
}

// readerWrapper is wrapper to disable readableStreamToReadCloser's WriteTo method.
type readerWrapper struct {
	io.Reader
}

func (sr *readableStreamToReadCloser) WriteTo(w io.Writer) (n int64, err error) {
	if w, ok := w.(RawJSBodyWriter); ok {
		w.WriteRawJSBody(sr.stream)
		return 0, nil
	}
	return io.Copy(w, &readerWrapper{sr})
}

func (sr *readableStreamToReadCloser) GetRawJSBody() js.Value {
	return sr.stream
}

// ConvertReadableStreamToReadCloser converts ReadableStream to io.ReadCloser.
func ConvertReadableStreamToReadCloser(stream js.Value) io.ReadCloser {
	return &readableStreamToReadCloser{
		stream: stream,
	}
}

// readerToReadableStream implements ReadableStream sourced from io.ReadCloser.
//   - ReadableStream: https://developer.mozilla.org/docs/Web/API/ReadableStream
//   - This implementation is based on: https://deno.land/std@0.139.0/streams/conversion.ts#L230
type readerToReadableStream struct {
	reader   io.ReadCloser
	chunkBuf []byte
}

// Pull implements ReadableStream's pull method.
//   - https://developer.mozilla.org/en-US/docs/Web/API/ReadableStream/ReadableStream#pull
func (rs *readerToReadableStream) Pull(controller js.Value) error {
	n, err := rs.reader.Read(rs.chunkBuf)
	if n != 0 {
		ua := NewUint8Array(n)
		js.CopyBytesToJS(ua, rs.chunkBuf[:n])
		controller.Call("enqueue", ua)
	}
	// Cloudflare Workers sometimes call `pull` to closed ReadableStream.
	// When the call happens, `io.ErrClosedPipe` should be ignored.
	if err == io.EOF || err == io.ErrClosedPipe {
		controller.Call("close")
		if err := rs.reader.Close(); err != nil {
			return err
		}
		return nil
	}
	if err != nil {
		controller.Call("error", Error(err.Error()))
		if err := rs.reader.Close(); err != nil {
			return err
		}
		return err
	}
	return nil
}

// Cancel implements ReadableStream's cancel method.
//   - https://developer.mozilla.org/en-US/docs/Web/API/ReadableStream/ReadableStream#cancel
func (rs *readerToReadableStream) Cancel() error {
	return rs.reader.Close()
}

// https://deno.land/std@0.139.0/streams/conversion.ts#L5
const defaultChunkSize = 16_640

// ConvertReaderToReadableStream converts io.ReadCloser to ReadableStream.
func ConvertReaderToReadableStream(reader io.ReadCloser) js.Value {
	stream := &readerToReadableStream{
		reader:   reader,
		chunkBuf: make([]byte, defaultChunkSize),
	}
	rsInit := NewObject()
	rsInit.Set("pull", js.FuncOf(func(_ js.Value, args []js.Value) any {
		var cb js.Func
		cb = js.FuncOf(func(this js.Value, pArgs []js.Value) any {
			defer cb.Release()
			resolve := pArgs[0]
			reject := pArgs[1]
			controller := args[0]
			go func() {
				err := stream.Pull(controller)
				if err != nil {
					reject.Invoke(Error(err.Error()))
					return
				}
				resolve.Invoke()
			}()
			return js.Undefined()
		})
		return NewPromise(cb)
	}))
	rsInit.Set("cancel", js.FuncOf(func(js.Value, []js.Value) any {
		var cb js.Func
		cb = js.FuncOf(func(this js.Value, pArgs []js.Value) any {
			defer cb.Release()
			resolve := pArgs[0]
			reject := pArgs[1]
			go func() {
				err := stream.Cancel()
				if err != nil {
					reject.Invoke(Error(err.Error()))
					return
				}
				resolve.Invoke()
			}()
			return js.Undefined()
		})
		return NewPromise(cb)
	}))
	return readableStreamClass.New(rsInit)
}
