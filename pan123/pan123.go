package pan123

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	netUrl "net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Pan123 struct {
	accessToken string
	timeout     time.Duration
	debug       bool

	httpCli *http.Client
}

// NewPan123 创建123云盘SDK实例
//
// @param timeout time.Duration HTTP请求超时时间, 默认为0
//
// @param debug bool 是否开启debug
//
// @return *Pan123
func NewPan123(timeout time.Duration, debug bool) *Pan123 {
	p123 := &Pan123{
		timeout: timeout,
		debug:   debug,
	}

	p123.httpCli = &http.Client{
		Transport: &http.Transport{
			Dial: func(network, addr string) (c net.Conn, err error) {
				return net.DialTimeout(network, addr, timeout)
			},
			Proxy: http.ProxyFromEnvironment,
		},
	}

	return p123
}

// GetAccessToken 获取当前accessToken
//
// @return string
func (p123 *Pan123) GetAccessToken() string {
	return p123.accessToken
}

// SetAccessToken 设置当前accessToken
//
// @param accessToken string access_token
func (p123 *Pan123) SetAccessToken(accessToken string) {
	p123.accessToken = accessToken
}

// RequestAccessToken 使用clientID、clientSecret请求accessToken
//
// @param clientID string client_id
//
// @param clientSecret string client_secret
//
// @return string accessToken
//
// @return time.Time accessToken过期时间
//
// @return SDKError
func (p123 *Pan123) RequestAccessToken(clientID, clientSecret string) (string, time.Time, error) {
	bodyData := map[string]interface{}{
		"clientID":     clientID,
		"clientSecret": clientSecret,
	}

	body, err := json.Marshal(bodyData)
	if err != nil {
		return "", time.Time{}, newSDKError(999, fmt.Sprintf("json.Marshal(req) error: %s", err), defaultTraceID)
	}
	resp, err := p123.callApi("/api/v1/access_token", "POST", body, map[string]string{}, false)
	if err != nil {
		return "", time.Time{}, err
	}

	var respData loginRespData
	err = toRespData(resp.Data, &respData)
	if err != nil {
		return "", time.Time{}, err
	}
	_accessTokenExpiredAt, err := time.Parse(time.RFC3339, respData.ExpiredAt)
	if err != nil {
		return "", time.Time{}, err
	}

	return respData.AccessToken, _accessTokenExpiredAt, nil
}

// CreateShare 创建分享链接
//
// 分享码: 分享码拼接至 https://www.123pan.com/s/ 后面访问,即是分享页面
//
// @param shareName string 分享链接
//
// @param fileIDList string 分享文件ID列表, 以逗号分割, 最大只支持拼接100个文件ID, 示例:1,2,3
//
// @param sharePwd string 分享链接提取码, 可为空
//
// @param shareExpire int 分享链接有效期天数, 1 -> 1天、7 -> 7天、30 -> 30天、0 -> 永久
//
// @return CreateShareRespData
//
// @return SDKError
func (p123 *Pan123) CreateShare(shareName, fileIDList, sharePwd string, shareExpire int) (*CreateShareRespData, error) {
	if shareExpire != 1 && shareExpire != 7 && shareExpire != 30 && shareExpire != 0 {
		return nil, newSDKError(999, "shareExpire invalid", defaultTraceID)
	}
	bodyData := map[string]interface{}{
		"shareName":   shareName,
		"fileIDList":  fileIDList,
		"sharePwd":    sharePwd,
		"shareExpire": shareExpire,
	}

	body, err := json.Marshal(bodyData)
	if err != nil {
		return nil, newSDKError(999, fmt.Sprintf("json.Marshal(req) error: %s", err), defaultTraceID)
	}
	resp, err := p123.callApi("/api/v1/share/create", "POST", body, map[string]string{}, true)
	if err != nil {
		return nil, err
	}

	var respData CreateShareRespData
	err = toRespData(resp.Data, &respData)
	if err != nil {
		return nil, err
	}

	return &respData, nil
}

// MkDir 创建目录
//
// @param name string 目录名(注:不能重名)
//
// @param parentID int 父目录id，创建到根目录时填写 0
//
// @return MkDirRespData
//
// @return SDKError
func (p123 *Pan123) MkDir(name string, parentID int64) (*MkDirRespData, error) {
	bodyData := map[string]interface{}{
		"name":     name,
		"parentID": parentID,
	}

	body, err := json.Marshal(bodyData)
	if err != nil {
		return nil, newSDKError(999, fmt.Sprintf("json.Marshal(req) error: %s", err), defaultTraceID)
	}
	resp, err := p123.callApi("/upload/v1/file/mkdir", "POST", body, map[string]string{}, true)
	if err != nil {
		return nil, err
	}

	var respData MkDirRespData
	err = toRespData(resp.Data, &respData)
	if err != nil {
		return nil, err
	}

	return &respData, nil
}

func (p123 *Pan123) fileUploadCreateFile(parentFileID int64, filename string, file *os.File, fileSize int64) (*fileUploadCreateFileRespData, error) {
	hash := md5.New()
	hashBuf := make([]byte, 4*1024*1024)
	for {
		n, err := file.Read(hashBuf)
		if err != nil && err != io.EOF {
			return nil, newSDKError(999, fmt.Sprintf("file.Read(hashBuf) error: %s", err), defaultTraceID)
		}
		if n == 0 {
			break
		}
		if _, err := hash.Write(hashBuf[:n]); err != nil {
			return nil, newSDKError(999, fmt.Sprintf("file.Read(hashBuf) error: %s", err), defaultTraceID)
		}
	}
	hashBuf = nil
	md5Sum := fmt.Sprintf("%x", hash.Sum(nil))
	hash.Reset()
	bodyData := map[string]interface{}{
		"parentFileID": parentFileID,
		"filename":     filename,
		"etag":         md5Sum,
		"size":         fileSize,
	}

	body, err := json.Marshal(bodyData)
	if err != nil {
		return nil, newSDKError(999, fmt.Sprintf("json.Marshal(req) error: %s", err), defaultTraceID)
	}
	resp, err := p123.callApi("/upload/v1/file/create", "POST", body, map[string]string{}, true)
	if err != nil {
		return nil, err
	}

	var respData fileUploadCreateFileRespData
	err = toRespData(resp.Data, &respData)
	if err != nil {
		return nil, err
	}
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return nil, newSDKError(999, fmt.Sprintf("file.Seek(io.SeekStart) error: %s", err), defaultTraceID)
	}

	return &respData, nil
}

func (p123 *Pan123) fileUploadGetChunkUploadUrl(preuploadID string, sliceNo int64) (*fileUploadGetChunkUploadUrlRespData, error) {
	bodyData := map[string]interface{}{
		"preuploadID": preuploadID,
		"sliceNo":     sliceNo,
	}

	body, err := json.Marshal(bodyData)
	if err != nil {
		return nil, newSDKError(999, fmt.Sprintf("json.Marshal(req) error: %s", err), defaultTraceID)
	}
	resp, err := p123.callApi("/upload/v1/file/get_upload_url", "POST", body, map[string]string{}, true)
	if err != nil {
		return nil, err
	}

	var respData fileUploadGetChunkUploadUrlRespData
	err = toRespData(resp.Data, &respData)
	if err != nil {
		return nil, err
	}

	return &respData, nil
}

func (p123 *Pan123) fileUploadChunkUpload(preuploadID string, sliceSize int64, file *os.File, retry int, cb FileUploadCallbackFunc, chunkCount int64) (*fileUploadChunkUploadRespData, error) {
	var currFileSliceNo int64 = 1
	chunkBuf := make([]byte, sliceSize)
	fileSliceSizes := map[int64]int64{}

	for {
		_currFileSliceNo := currFileSliceNo
		cb(FileUploadCallbackInfo{
			Status:     FILE_UPLOAD_CALLBACK_STATUS_FIRST_UPLOAD_CHUNK,
			ChunkID:    _currFileSliceNo,
			ChunkCount: chunkCount,
		})
		// 获取块上传地址
		getChunkUploadUrlResp, err := p123.fileUploadGetChunkUploadUrl(preuploadID, currFileSliceNo)
		if err != nil {
			return nil, err
		}

		// 读取块
		n, err := file.Read(chunkBuf)
		if err != nil && err != io.EOF {
			return nil, newSDKError(999, fmt.Sprintf("file.Read(chunkBuf) error: %s", err), defaultTraceID)
		}
		if n == 0 {
			break
		}
		fileSliceSizes[currFileSliceNo] = int64(n)
		// 在读取块内容后就对块ID进行累加
		currFileSliceNo++

		// 上传块
		var _chunkBuf bytes.Buffer
		_chunkBuf.Write(chunkBuf[:n])
		nowRetry := 0
		var retryErr error
		for {
			if nowRetry > retry {
				break
			}
			if nowRetry != 0 {
				cb(FileUploadCallbackInfo{
					Status:     FILE_UPLOAD_CALLBACK_STATUS_RETRY_UPLOAD_CHUNK,
					ChunkID:    _currFileSliceNo,
					ChunkCount: chunkCount,
				})
			}
			chunkUploadResp, err := p123.doHTTPRequest("PUT", getChunkUploadUrlResp.PresignedURL, map[string]string{}, map[string]string{}, &_chunkBuf)
			if err != nil {
				retryErr = newSDKError(999, fmt.Sprintf("http error: %s", err), defaultTraceID)
				nowRetry++
				continue
			}
			if chunkUploadResp == nil {
				retryErr = newSDKError(999, "p123.doHTTPRequest nil?", defaultTraceID)
				nowRetry++
				continue
			}
			if chunkUploadResp.Body != nil {
				_ = chunkUploadResp.Body.Close()
			}
			if chunkUploadResp.StatusCode != 204 && chunkUploadResp.StatusCode != 200 {
				retryErr = newSDKError(999, fmt.Sprintf("http_code error: %d", chunkUploadResp.StatusCode), defaultTraceID)
				nowRetry++
				continue
			}
			break
		}
		if retryErr != nil {
			// 已经到了retry的次数
			return nil, newSDKError(999, fmt.Sprintf("maxRetry, last error: %s", retryErr), defaultTraceID)
		}
		_chunkBuf.Reset()
	}

	return &fileUploadChunkUploadRespData{fileSliceSizes: fileSliceSizes}, nil
}

func (p123 *Pan123) fileUploadListUploadParts(preuploadID string) (*fileUploadListUploadPartsRespData, error) {
	bodyData := map[string]interface{}{
		"preuploadID": preuploadID,
	}

	body, err := json.Marshal(bodyData)
	if err != nil {
		return nil, newSDKError(999, fmt.Sprintf("json.Marshal(req) error: %s", err), defaultTraceID)
	}
	resp, err := p123.callApi("/upload/v1/file/list_upload_parts", "POST", body, map[string]string{}, true)
	if err != nil {
		return nil, err
	}

	var respData fileUploadListUploadPartsRespData
	err = toRespData(resp.Data, &respData)
	if err != nil {
		return nil, err
	}

	return &respData, nil
}

func (p123 *Pan123) fileUploadUploadComplete(preuploadID string) (*fileUploadUploadCompleteRespData, error) {
	bodyData := map[string]interface{}{
		"preuploadID": preuploadID,
	}

	body, err := json.Marshal(bodyData)
	if err != nil {
		return nil, newSDKError(999, fmt.Sprintf("json.Marshal(req) error: %s", err), defaultTraceID)
	}
	resp, err := p123.callApi("/upload/v1/file/upload_complete", "POST", body, map[string]string{}, true)
	if err != nil {
		return nil, err
	}

	var respData fileUploadUploadCompleteRespData
	err = toRespData(resp.Data, &respData)
	if err != nil {
		return nil, err
	}

	return &respData, nil
}

// FileUploadWithCallback 带Callback上传文件
//
// @param parentFileID int64 父目录id, 上传到根目录时填写0
//
// @param filename string 文件名要小于128个字符且不能包含以下任何字符："\/:*?|><。（注：不能重名）
//
// @param file *os.File 要上传的文件句柄
//
// @param retry int 上传单一文件块时的重试次数, 0为不重试
//
// @param cb FileUploadCallbackFunc Callback
//
// @return FileUploadRespData
//
// @return SDKError
func (p123 *Pan123) FileUploadWithCallback(parentFileID int64, filename string, file *os.File, retry int, cb FileUploadCallbackFunc) (*FileUploadRespData, error) {
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, newSDKError(999, fmt.Sprintf("content.Stat error: %s", err), defaultTraceID)
	}
	if fileInfo.Size() <= 0 {
		return nil, newSDKError(999, "file_size <= 0", defaultTraceID)
	}
	if cb == nil {
		cb = func(_ FileUploadCallbackInfo) {}
	}

	// 创建文件
	cb(FileUploadCallbackInfo{
		Status: FILE_UPLOAD_CALLBACK_STATUS_CREATE_FILE,
	})
	createFileResp, err := p123.fileUploadCreateFile(parentFileID, filename, file, fileInfo.Size())
	if err != nil {
		return nil, err
	}
	if createFileResp.Reuse {
		// 秒传
		return &FileUploadRespData{FileID: createFileResp.FileID, Reuse: true}, nil
	}

	// 分块上传
	var chunkCount int64 = fileInfo.Size() / createFileResp.SliceSize
	if fileInfo.Size()%createFileResp.SliceSize != 0 {
		chunkCount++
	}
	chunkUploadResp, err := p123.fileUploadChunkUpload(createFileResp.PreuploadID, createFileResp.SliceSize, file, retry, cb, chunkCount)
	if err != nil {
		return nil, err
	}

	// 上传完毕, 进行校验
	if createFileResp.SliceSize < fileInfo.Size() && len(chunkUploadResp.fileSliceSizes) > 1 {
		cb(FileUploadCallbackInfo{
			Status:     FILE_UPLOAD_CALLBACK_STATUS_VERIFY_CHUNK,
			ChunkCount: chunkCount,
		})
		listUploadPartsResp, err := p123.fileUploadListUploadParts(createFileResp.PreuploadID)
		if err != nil {
			return nil, err
		}
		for _, v := range listUploadPartsResp.Parts {
			_partNumber, err := strconv.ParseInt(v.PartNumber, 10, 0)
			if err != nil {
				return nil, newSDKError(999, fmt.Sprintf("chunk _partNumber convert error: %s", err), defaultTraceID)
			}
			if _v, ok := chunkUploadResp.fileSliceSizes[_partNumber]; ok {
				if _v != v.Size {
					return nil, newSDKError(999, fmt.Sprintf("chunk %d size %d != %d", _partNumber, _v, v.Size), defaultTraceID)
				}
			} else {
				return nil, newSDKError(999, fmt.Sprintf("chunk %d not found", _partNumber), defaultTraceID)
			}
		}
	}

	// 通知上传完成
	cb(FileUploadCallbackInfo{
		Status: FILE_UPLOAD_CALLBACK_STATUS_REPORT_COMPLETE,
	})
	uploadCompleteResp, err := p123.fileUploadUploadComplete(createFileResp.PreuploadID)
	if err != nil {
		return nil, err
	}
	if uploadCompleteResp.Completed {
		// 上传成功
		return &FileUploadRespData{FileID: uploadCompleteResp.FileID}, nil
	}
	if uploadCompleteResp.Async {
		// 需要异步查询上传结果
		return &FileUploadRespData{PreuploadID: createFileResp.PreuploadID, Async: true}, nil
	}

	return nil, newSDKError(999, "upload failed", defaultTraceID)
}

// FileUpload 上传文件
//
// @param parentFileID int64 父目录id, 上传到根目录时填写0
//
// @param filename string 文件名要小于128个字符且不能包含以下任何字符："\/:*?|><。（注：不能重名）
//
// @param retry int 上传单一文件块时的重试次数, 0为不重试
//
// @return FileUploadRespData
//
// @return SDKError
func (p123 *Pan123) FileUpload(parentFileID int64, filename string, file *os.File, retry int) (*FileUploadRespData, error) {
	return p123.FileUploadWithCallback(parentFileID, filename, file, retry, nil)
}

// GetUploadAsyncResult 异步轮询获取上传结果
//
// @param preuploadID string 预上传ID
//
// @return UploadAsyncResultRespData
//
// @return SDKError
func (p123 *Pan123) GetUploadAsyncResult(preuploadID string) (*UploadAsyncResultRespData, error) {
	bodyData := map[string]interface{}{
		"preuploadID": preuploadID,
	}

	body, err := json.Marshal(bodyData)
	if err != nil {
		return nil, newSDKError(999, fmt.Sprintf("json.Marshal(req) error: %s", err), defaultTraceID)
	}
	resp, err := p123.callApi("/upload/v1/file/upload_async_result", "POST", body, map[string]string{}, true)
	if err != nil {
		return nil, err
	}

	var respData UploadAsyncResultRespData
	err = toRespData(resp.Data, &respData)
	if err != nil {
		return nil, err
	}

	return &respData, nil
}

// MoveFile 移动文件
//
// 批量移动文件，单级最多支持100个
//
// @param fileIDs []int64 文件id数组
//
// @param toParentFileID int64 要移动到的目标文件夹id，移动到根目录时填写 0
//
// @return SDKError
func (p123 *Pan123) MoveFile(fileIDs []int64, toParentFileID int64) error {
	bodyData := map[string]interface{}{
		"fileIDs":        fileIDs,
		"toParentFileID": toParentFileID,
	}

	body, err := json.Marshal(bodyData)
	if err != nil {
		return newSDKError(999, fmt.Sprintf("json.Marshal(req) error: %s", err), defaultTraceID)
	}
	_, err = p123.callApi("/api/v1/file/move", "POST", body, map[string]string{}, true)

	return err
}

// TrashFile 删除文件至回收站
//
// 删除的文件，会放入回收站中
//
// @param fileIDs []int64 文件id数组,一次性最大不能超过 100 个文件
//
// @return SDKError
func (p123 *Pan123) TrashFile(fileIDs []int64) error {
	bodyData := map[string]interface{}{
		"fileIDs": fileIDs,
	}

	body, err := json.Marshal(bodyData)
	if err != nil {
		return newSDKError(999, fmt.Sprintf("json.Marshal(req) error: %s", err), defaultTraceID)
	}
	_, err = p123.callApi("/api/v1/file/trash", "POST", body, map[string]string{}, true)

	return err
}

// RecoverFile 从回收站恢复文件
//
// 将回收站的文件恢复至删除前的位置
//
// @param fileIDs []int64 文件id数组,一次性最大不能超过 100 个文件
//
// @return SDKError
func (p123 *Pan123) RecoverFile(fileIDs []int64) error {
	bodyData := map[string]interface{}{
		"fileIDs": fileIDs,
	}

	body, err := json.Marshal(bodyData)
	if err != nil {
		return newSDKError(999, fmt.Sprintf("json.Marshal(req) error: %s", err), defaultTraceID)
	}
	_, err = p123.callApi("/api/v1/file/recover", "POST", body, map[string]string{}, true)

	return err
}

// DeleteFile 彻底删除文件
//
// 彻底删除文件前,文件必须要在回收站中,否则无法删除
//
// @param fileIDs []int64 文件id数组,一次性最大不能超过 100 个文件
//
// @return SDKError
func (p123 *Pan123) DeleteFile(fileIDs []int64) error {
	bodyData := map[string]interface{}{
		"fileIDs": fileIDs,
	}

	body, err := json.Marshal(bodyData)
	if err != nil {
		return newSDKError(999, fmt.Sprintf("json.Marshal(req) error: %s", err), defaultTraceID)
	}
	_, err = p123.callApi("/api/v1/file/delete", "POST", body, map[string]string{}, true)

	return err
}

// GetFileList 获取文件列表
//
// @param parentFileId int64 文件夹ID，根目录传 0
//
// @param page int64 页码数
//
// @param limit int64 每页文件数量，最大不超过100
//
// @param orderBy string 排序字段,例如:file_id、size、file_name
//
// @param orderDirection string 排序方向:asc、desc
//
// @param trashed boolean 是否查看回收站的文件
//
// @param searchData string 搜索关键字
//
// @return GetFileListRespData
//
// @return SDKError
func (p123 *Pan123) GetFileList(parentFileId, page, limit int64, orderBy, orderDirection string, trashed bool, searchData string) (*GetFileListRespData, error) {
	if orderBy != "file_id" && orderBy != "size" && orderBy != "file_name" {
		return nil, newSDKError(999, "orderBy invalid", defaultTraceID)
	}
	if orderDirection != "asc" && orderDirection != "desc" {
		return nil, newSDKError(999, "orderDirection invalid", defaultTraceID)
	}
	querys := map[string]string{
		"parentFileId":   strconv.FormatInt(parentFileId, 10),
		"page":           strconv.FormatInt(page, 10),
		"limit":          strconv.FormatInt(limit, 10),
		"orderBy":        orderBy,
		"orderDirection": orderDirection,
		"searchData":     searchData,
	}
	if trashed {
		querys["trashed"] = "true"
	}

	resp, err := p123.callApi("/api/v1/file/list", "GET", nil, querys, true)
	if err != nil {
		return nil, err
	}

	var respData GetFileListRespData
	err = toRespData(resp.Data, &respData)
	if err != nil {
		return nil, err
	}

	return &respData, nil
}

// GetUserInfo 获取用户信息
//
// @return UploadAsyncResultRespData
//
// @return SDKError
func (p123 *Pan123) GetUserInfo() (*GetUserInfoRespData, error) {
	resp, err := p123.callApi("/api/v1/user/info", "GET", nil, map[string]string{}, true)
	if err != nil {
		return nil, err
	}

	var respData GetUserInfoRespData
	err = toRespData(resp.Data, &respData)
	if err != nil {
		return nil, err
	}

	return &respData, nil
}

// OfflineDownload 创建离线下载任务
//
// 离线下载任务仅支持 http/https 任务创建
//
// @param url string 下载资源地址(http/https)
//
// @param fileName string 自定义文件名称
//
// @param callBackUrl string 回调地址, 回调内容请参考: https://123yunpan.yuque.com/org-wiki-123yunpan-muaork/cr6ced/wn77piehmp9t8ut4#jf5bZ
//
// @param dirID int64 下载到的指定目录ID, 不支持下载到根目录, 传0会下载到名为"来自:离线下载"的目录中
//
// @return UploadAsyncResultRespData
//
// @return SDKError
func (p123 *Pan123) OfflineDownload(url, fileName, callBackUrl string, dirID int64) (*OfflineDownloadRespData, error) {
	bodyData := map[string]interface{}{
		"url": url,
	}
	if fileName != "" {
		bodyData["fileName"] = fileName
	}
	if callBackUrl != "" {
		bodyData["callBackUrl"] = callBackUrl
	}
	if dirID != 0 {
		bodyData["dirID"] = dirID
	}

	body, err := json.Marshal(bodyData)
	if err != nil {
		return nil, newSDKError(999, fmt.Sprintf("json.Marshal(req) error: %s", err), defaultTraceID)
	}
	resp, err := p123.callApi("api/v1/offline/download", "POST", body, map[string]string{}, true)
	if err != nil {
		return nil, err
	}

	var respData OfflineDownloadRespData
	err = toRespData(resp.Data, &respData)
	if err != nil {
		return nil, err
	}

	return &respData, nil
}

// QueryDirectLinkTranscode 查询直链转码进度
//
// @param ids []int64 视频文件ID列表
//
// @return QueryDirectLinkTranscodeRespData
//
// @return SDKError
func (p123 *Pan123) QueryDirectLinkTranscode(ids []int64) (*QueryDirectLinkTranscodeRespData, error) {
	bodyData := map[string]interface{}{
		"ids": ids,
	}

	body, err := json.Marshal(bodyData)
	if err != nil {
		return nil, newSDKError(999, fmt.Sprintf("json.Marshal(req) error: %s", err), defaultTraceID)
	}
	resp, err := p123.callApi("/api/v1/direct-link/queryTranscode", "POST", body, map[string]string{}, true)
	if err != nil {
		return nil, err
	}

	var respData QueryDirectLinkTranscodeRespData
	err = toRespData(resp.Data, &respData)
	if err != nil {
		return nil, err
	}

	return &respData, nil
}

// DoDirectLinkTranscode 发起直链转码
//
// 请注意: 文件必须要在直链空间下,且源文件是视频文件才能进行转码操作
//
// @param ids []int64 需要转码的文件ID列表
//
// @return SDKError
func (p123 *Pan123) DoDirectLinkTranscode(ids []int64) error {
	bodyData := map[string]interface{}{
		"ids": ids,
	}

	body, err := json.Marshal(bodyData)
	if err != nil {
		return newSDKError(999, fmt.Sprintf("json.Marshal(req) error: %s", err), defaultTraceID)
	}
	_, err = p123.callApi("/api/v1/direct-link/doTranscode", "POST", body, map[string]string{}, true)

	return err
}

// GetDirectLinkM3u8 获取直链转码链接
//
// @param fileID int64 文件ID
//
// @return GetDirectLinkM3u8RespData
//
// @return SDKError
func (p123 *Pan123) GetDirectLinkM3u8(fileID int64) (*GetDirectLinkM3u8RespData, error) {
	querys := map[string]string{
		"fileID": strconv.FormatInt(fileID, 10),
	}

	resp, err := p123.callApi("/api/v1/direct-link/get/m3u8", "GET", nil, querys, true)
	if err != nil {
		return nil, err
	}

	var respData GetDirectLinkM3u8RespData
	err = toRespData(resp.Data, &respData)
	if err != nil {
		return nil, err
	}

	return &respData, nil
}

// EnableDirectLink 启用直链空间
//
// @param fileID int64 启用直链空间的文件夹的fileID
//
// @return EnableDirectLinkRespData
//
// @return SDKError
func (p123 *Pan123) EnableDirectLink(fileID int64) (*EnableDirectLinkRespData, error) {
	bodyData := map[string]interface{}{
		"fileID": fileID,
	}

	body, err := json.Marshal(bodyData)
	if err != nil {
		return nil, newSDKError(999, fmt.Sprintf("json.Marshal(req) error: %s", err), defaultTraceID)
	}
	resp, err := p123.callApi("/api/v1/direct-link/enable", "POST", body, map[string]string{}, true)
	if err != nil {
		return nil, err
	}

	var respData EnableDirectLinkRespData
	err = toRespData(resp.Data, &respData)
	if err != nil {
		return nil, err
	}

	return &respData, nil
}

// DisableDirectLink 禁用直链空间
//
// @param fileID int64 禁用直链空间的文件夹的fileID
//
// @return DisableDirectLinkRespData
//
// @return SDKError
func (p123 *Pan123) DisableDirectLink(fileID int64) (*DisableDirectLinkRespData, error) {
	bodyData := map[string]interface{}{
		"fileID": fileID,
	}

	body, err := json.Marshal(bodyData)
	if err != nil {
		return nil, newSDKError(999, fmt.Sprintf("json.Marshal(req) error: %s", err), defaultTraceID)
	}
	resp, err := p123.callApi("/api/v1/direct-link/disable", "POST", body, map[string]string{}, true)
	if err != nil {
		return nil, err
	}

	var respData DisableDirectLinkRespData
	err = toRespData(resp.Data, &respData)
	if err != nil {
		return nil, err
	}

	return &respData, nil
}

// GetDirectLinkUrl 获取直链链接
//
// @param fileID int64 需要获取直链链接的文件的fileID
//
// @return GetDirectLinkUrlRespData
//
// @return SDKError
func (p123 *Pan123) GetDirectLinkUrl(fileID int64) (*GetDirectLinkUrlRespData, error) {
	querys := map[string]string{
		"fileID": strconv.FormatInt(fileID, 10),
	}
	resp, err := p123.callApi("/api/v1/direct-link/url", "GET", nil, querys, true)
	if err != nil {
		return nil, err
	}

	var respData GetDirectLinkUrlRespData
	err = toRespData(resp.Data, &respData)
	if err != nil {
		return nil, err
	}

	return &respData, nil
}

// RenameFile 重命名文件
//
// @param renameList []string 数组,每个成员的格式为 文件ID|新的文件名
//
// @return SDKError
func (p123 *Pan123) RenameFile(renameList []string) error {
	bodyData := map[string]interface{}{
		"renameList": renameList,
	}

	body, err := json.Marshal(bodyData)
	if err != nil {
		return newSDKError(999, fmt.Sprintf("json.Marshal(req) error: %s", err), defaultTraceID)
	}
	_, err = p123.callApi("/api/v1/file/rename", "POST", body, map[string]string{}, true)

	return err
}

// GetFileDetail 获取文件详情
//
// @param fileID int64 文件ID
//
// @return GetFileDetailRespData
//
// @return SDKError
func (p123 *Pan123) GetFileDetail(fileID int64) (*GetFileDetailRespData, error) {
	querys := map[string]string{
		"fileID": strconv.FormatInt(fileID, 10),
	}
	resp, err := p123.callApi("/api/v1/file/detail", "GET", nil, querys, true)
	if err != nil {
		return nil, err
	}

	var respData GetFileDetailRespData
	err = toRespData(resp.Data, &respData)
	if err != nil {
		return nil, err
	}

	return &respData, nil
}

// GetOfflineDownloadProcess 获取离线下载进度
//
// @param taskID int64 离线下载任务ID
//
// @return GetFileDetailRespData
//
// @return SDKError
func (p123 *Pan123) GetOfflineDownloadProcess(taskID int64) (*GetOfflineDownloadProcessRespData, error) {
	querys := map[string]string{
		"taskID": strconv.FormatInt(taskID, 10),
	}
	resp, err := p123.callApi("/api/v1/offline/download/process", "GET", nil, querys, true)
	if err != nil {
		return nil, err
	}

	var respData GetOfflineDownloadProcessRespData
	err = toRespData(resp.Data, &respData)
	if err != nil {
		return nil, err
	}

	return &respData, nil
}

func (p123 *Pan123) callApi(path, method string, body []byte, querys map[string]string, withAuth bool) (*callApiResp, error) {
	headers := map[string]string{}
	r := &callApiResp{}
	accessToken := ""
	if withAuth {
		accessToken = p123.accessToken
	}

	data, err := p123.doApiRequest(method, path, accessToken, querys, headers, body)
	if err != nil {
		var sdkError *SDKError
		if !errors.As(err, &sdkError) {
			// 不应被触发
			return nil, err
		}
		return nil, err
	} else {
		r.Data = data
	}

	return r, nil
}

func (p123 *Pan123) doApiRequest(method, path, accessToken string, querys map[string]string, headers map[string]string, body []byte) (map[string]interface{}, error) {
	headers["Platform"] = "open_platform"
	headers["User-Agent"] = "123PAN-UNOFFICIAL-GO-SDK"
	if accessToken != "" {
		headers["Authorization"] = "Bearer " + accessToken
	}
	if method == "POST" {
		headers["Content-Type"] = "application/json"
	}

	var buf bytes.Buffer
	if body != nil {
		buf.Write(body)
	}
	resp, err := p123.doHTTPRequest(method, "https://open-api.123pan.com"+path, querys, headers, &buf)
	defer func() {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()
	if err != nil {
		return nil, newSDKError(999, fmt.Sprintf("http error: %s", err), defaultTraceID)
	}
	if resp == nil {
		return nil, newSDKError(999, "p123.doHTTPRequest nil?", defaultTraceID)
	}
	if resp.StatusCode != 200 {
		return nil, newSDKError(999, fmt.Sprintf("http_code error: %d", resp.StatusCode), defaultTraceID)
	}
	respBuf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, newSDKError(999, fmt.Sprintf("http_body read error: %s", err), defaultTraceID)
	}

	var r apiHttpResp
	err = json.Unmarshal(respBuf, &r)
	if err != nil {
		return nil, newSDKError(999, fmt.Sprintf("http_resp not json: %s", err), defaultTraceID)
	}
	if r.Code != 0 {
		// 接口错误响应
		return nil, newSDKError(r.Code, r.Message, r.TraceID)
	}

	return r.Data, nil
}

func (p123 *Pan123) doHTTPRequest(method, url string, querys map[string]string, headers map[string]string, body io.Reader) (resp *http.Response, err error) {
	if len(querys) > 0 {
		_q := netUrl.Values{}
		for k, v := range querys {
			_q.Add(k, v)
		}
		url = url + "?" + _q.Encode()
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		if strings.ToLower(k) == "host" {
			req.Host = v
		} else {
			req.Header.Set(k, v)
		}
	}

	if method == "PUT" || method == "POST" {
		found := false
		length := req.Header.Get("Content-Length")
		if length != "" {
			req.ContentLength, _ = strconv.ParseInt(length, 10, 64)
			found = true
		} else {
			switch v := body.(type) {
			case *os.File:
				if fInfo, err := v.Stat(); err == nil {
					req.ContentLength = fInfo.Size()
					found = true
				}
			case *bytes.Buffer:
				req.ContentLength = int64(v.Len())
				found = true
			case *bytes.Reader:
				req.ContentLength = int64(v.Len())
				found = true
			case *strings.Reader:
				req.ContentLength = int64(v.Len())
				found = true
			case *io.LimitedReader:
				req.ContentLength = v.N
				found = true
			}
		}
		if found && req.ContentLength == 0 {
			req.Body = nil
		}
	}

	if p123.debug {
		fmt.Printf("%+v\n", req)
	}

	resp, err = p123.httpCli.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func toRespData(i map[string]interface{}, o interface{}) error {
	b, err := json.Marshal(i)
	if err != nil {
		return newSDKError(999, fmt.Sprintf("json.Marshal(resp) error: %s", err), defaultTraceID)
	}
	err = json.Unmarshal(b, o)
	if err != nil {
		return newSDKError(999, fmt.Sprintf("json.Unmarshal(resp) error: %s", err), defaultTraceID)
	}

	return nil
}
