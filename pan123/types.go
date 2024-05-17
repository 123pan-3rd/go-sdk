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

type loginRespData struct {
	AccessToken string `json:"accessToken"`
	ExpiredAt   string `json:"expiredAt"`
}

type CreateShareRespData struct {
	// 分享ID
	ShareID int64 `json:"shareID"`
	// 分享码
	ShareKey string `json:"shareKey"`
}

type MkDirRespData struct {
	// 创建的目录ID
	DirID int64 `json:"dirID"`
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

type FileUploadRespData struct {
	// 预上传ID, 仅在需要异步查询上传结果时存在
	PreuploadID string
	// 是否秒传
	Reuse bool
	// 文件ID, 仅在秒传或无需异步查询上传结果时存在
	FileID int64
	// 是否需要异步查询上传结果
	Async bool
}

type UploadAsyncResultRespData struct {
	// 上传合并是否完成
	Completed bool `json:"completed"`
	// 上传成功的文件ID
	FileID int64 `json:"fileID"`
}

type GetFileListRespData struct {
	// 文件列表
	FileList []FileListInfoRespData `json:"fileList"`
}

type FileListInfoRespData struct {
	// 文件ID
	FileID int64 `json:"fileID"`
	// 文件名
	Filename string `json:"filename"`
	// 0-文件  1-文件夹
	Type int `json:"type"`
	// 文件大小
	Size int64 `json:"size"`
	// md5
	Etag string `json:"etag"`
	// 文件审核状态, 大于 100 为审核驳回文件
	Status int `json:"status"`
	// 目录ID
	ParentFileID int64 `json:"parentFileID"`
	// 目录名
	ParentName string `json:"parentName"`
	// 文件分类, 0-未知 1-音频 2-视频 3-图片
	Category int `json:"category"`
	// 文件类型
	ContentType string `json:"contentType"`
}

type GetUserInfoRespData struct {
	// 用户账号ID
	Uid int64 `json:"uid"`
	// 昵称
	Nickname string `json:"nickname"`
	// 头像
	HeadImage string `json:"headImage"`
	// 手机号码
	Passport string `json:"passport"`
	// 邮箱
	Mail string `json:"mail"`
	// 已用空间
	SpaceUsed int64 `json:"spaceUsed"`
	// 永久空间
	SpacePermanent int64 `json:"spacePermanent"`
	// 临时空间
	SpaceTemp int64 `json:"spaceTemp"`
	// 临时空间到期日
	SpaceTempExpr string `json:"spaceTempExpr"`
}

type QueryDirectLinkTranscodeRespData struct {
	// 未发起过转码的ID
	NoneList []int64 `json:"noneList"`
	// 错误文件ID列表, 这些文件ID无法进行转码操作
	ErrorList []QueryDirectLinkTranscodeErrorListRespData `json:"errorList"`
	// 转码成功的文件ID列表
	Success []int64 `json:"success"`
	// 正在转码的文件ID列表
	Running []int64 `json:"running"`
}

type QueryDirectLinkTranscodeErrorListRespData struct {
	// 文件ID
	ID []int64 `json:"id"`
	// 错误信息
	ErrorReason string `json:"errorReason"`
}

type GetDirectLinkM3u8RespData struct {
	// 直链转码列表
	List []GetDirectLinkM3u8InfoRespData `json:"list"`
}

type GetDirectLinkM3u8InfoRespData struct {
	// 分辨率
	Resolutions string `json:"resolutions"`
	// 播放地址
	Address string `json:"address"`
}

type EnableDirectLinkRespData struct {
	// 成功启用直链空间的文件夹的名称
	Filename string `json:"filename"`
}

type DisableDirectLinkRespData struct {
	// 成功禁用直链空间的文件夹的名称
	Filename string `json:"filename"`
}

type GetDirectLinkUrlRespData struct {
	// 文件对应的直链链接
	Url string `json:"url"`
}

type GetFileDetailRespData struct {
	// 文件ID
	FileID int64 `json:"fileID"`
	// 文件名
	Filename string `json:"filename"`
	// 0-文件  1-文件夹
	Type int `json:"type"`
	// 文件大小
	Size int64 `json:"size"`
	// md5
	Etag string `json:"etag"`
	// 文件审核状态, 大于 100 为审核驳回文件
	Status int `json:"status"`
	// 父级ID
	ParentFileID int64 `json:"parentFileID"`
	// 目录名
	ParentName string `json:"parentName"`
	// 文件创建时间
	CreateAt string `json:"createAt"`
	// 该文件是否在回收站, 0-否、1-是
	Trashed int `json:"trashed"`
}

type GetOfflineDownloadProcessRespData struct {
	// 下载进度百分比,当文件下载失败,该进度将会归零
	Process float64 `json:"process"`
	// 下载状态, 0-进行中、1-下载失败、2-下载成功、3-重试中
	Status int `json:"status"`
}

type OfflineDownloadRespData struct {
	// 离线下载任务ID,可通过该ID调用查询任务进度接口获取下载进度
	TaskID int64 `json:"taskID"`
	// 下载状态, 0-进行中、1-下载失败、2-下载成功、3-重试中
	Status int `json:"status"`
}
