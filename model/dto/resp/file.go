package resp

import "time"

type FileInfo struct {
	UUID      string `json:"uuid"`
	FileName  string `json:"file_name"`
	FileURL   string `json:"file_url"`
	FileType  string `json:"file_type"`
	FileExt   string `json:"file_ext"`
	MimeType  string `json:"mime_type"`
	FileSize  uint64 `json:"file_size"`
	Hash      string `json:"hash"`
	CreatedAt string `json:"created_at"`
}

type CodeWithMsgFile struct {
	Code int      `json:"code"`
	Msg  string   `json:"msg"`
	File FileInfo `json:"file"`
}

type TempFileLink struct {
	UUID      string    `json:"uuid"`
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
}

type CodeWithMsgTempFileLink struct {
	Code int          `json:"code"`
	Msg  string       `json:"msg"`
	Link TempFileLink `json:"link"`
}
