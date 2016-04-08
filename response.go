package gbeta

import (
	"net/http"
)

type _Response struct {
	standard_writer http.ResponseWriter
	status_code     int
	byte_writed     int64
}

//implement writer interface
func (r *_Response) Write(data []byte) (int, error) {
	r.byte_writed += int64(len(data))
	if r.status_code == 0 {
		r.WriteHeader(http.StatusOK)
	}
	return r.standard_writer.Write(data)
}

//like the standard Header
func (r *_Response) Header() http.Header {
	return r.standard_writer.Header()
}

//like the standard WriteHeader
func (r *_Response) WriteHeader(status int) {
	if r.status_code == 0 {
		r.status_code = status //record the status for use in future
		r.standard_writer.WriteHeader(status)
	}
}

// return the status code
func (r *_Response) Code() int {
	return r.status_code
}

// return the bytes written
func (r *_Response) BytesWritten() int64 {
	return r.byte_writed
}
