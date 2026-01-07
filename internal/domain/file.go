package domain

// FileDescriptor describes a stored file for download.
type FileDescriptor struct {
	RelativePath string
	MimeType     string
	ETag         string
}
