package pan123

type apiHttpResp struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	TraceID string                 `json:"x-traceID"`
	Data    map[string]interface{} `json:"data"`
}

type callApiResp struct {
	TokenRefresh bool
	Data         map[string]interface{}
}

type LoginRespData struct {
	AccessToken string `json:"accessToken"`
	ExpiredAt   string `json:"expiredAt"`
}

type CreateShareRespData struct {
	ShareID  int64  `json:"shareID"`
	ShareKey string `json:"shareKey"`
}

type MkDirRespData struct {
	DirID int64 `json:"dirID"`
}

type FileUploadRespData struct {
	PreuploadID string
	Reuse       bool
	FileID      int64
	Async       bool
}

type UploadAsyncResultRespData struct {
	Completed bool  `json:"completed"`
	FileID    int64 `json:"fileID"`
}

type GetFileListRespData struct {
	FileList []FileListInfoRespData `json:"fileList"`
}

type FileListInfoRespData struct {
	FileID       int64  `json:"fileID"`
	Filename     string `json:"filename"`
	Type         int    `json:"type"`
	Size         int64  `json:"size"`
	Etag         string `json:"etag"`
	Status       int    `json:"status"`
	ParentFileID int64  `json:"parentFileID"`
	ParentName   string `json:"parentName"`
	Category     int    `json:"category"`
	ContentType  string `json:"contentType"`
}

type GetUserInfoRespData struct {
	Uid            int64  `json:"uid"`
	Nickname       string `json:"nickname"`
	HeadImage      string `json:"headImage"`
	Passport       string `json:"passport"`
	Mail           string `json:"mail"`
	SpaceUsed      int64  `json:"spaceUsed"`
	SpacePermanent int64  `json:"spacePermanent"`
	SpaceTemp      int64  `json:"spaceTemp"`
	SpaceTempExpr  string `json:"spaceTempExpr"`
}

type QueryDirectLinkTranscodeRespData struct {
	NoneList  []int64                                     `json:"noneList"`
	ErrorList []QueryDirectLinkTranscodeErrorListRespData `json:"errorList"`
	Success   []int64                                     `json:"success"`
	Running   []int64                                     `json:"running"`
}

type QueryDirectLinkTranscodeErrorListRespData struct {
	ID          []int64 `json:"id"`
	ErrorReason string  `json:"errorReason"`
}

type GetDirectLinkM3u8RespData struct {
	List []GetDirectLinkM3u8InfoRespData `json:"list"`
}

type GetDirectLinkM3u8InfoRespData struct {
	Resolutions string `json:"resolutions"`
	Address     string `json:"address"`
}

type EnableDirectLinkRespData struct {
	Filename string `json:"filename"`
}

type DisableDirectLinkRespData struct {
	Filename string `json:"filename"`
}

type GetDirectLinkUrlRespData struct {
	Url string `json:"url"`
}

type fileUploadCreateFileRespData struct {
	FileID      int64  `json:"fileID"`
	PreuploadID string `json:"preuploadID"`
	Reuse       bool   `json:"reuse"`
	SliceSize   int64  `json:"sliceSize"`
}

type fileUploadGetChunkUploadUrlRespData struct {
	PresignedURL string `json:"presignedURL"`
}

type fileUploadChunkUploadRespData struct {
	fileSliceSizes map[int64]int64
}

type fileUploadListUploadPartsRespData struct {
	Parts []fileUploadListUploadPartsInfoRespData `json:"parts"`
}

type fileUploadListUploadPartsInfoRespData struct {
	PartNumber string `json:"partNumber"`
	Size       int64  `json:"size"`
	Etag       string `json:"etag"`
}

type fileUploadUploadCompleteRespData struct {
	FileID    int64 `json:"fileID"`
	Async     bool  `json:"async"`
	Completed bool  `json:"completed"`
}
