package uploadreq

type UploadFile struct {
	FileType   string `form:"file_type" json:"file_type"`
	Permission uint8  `form:"permission" json:"permission"`
}
