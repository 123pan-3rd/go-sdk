// $Env:HTTP_PROXY="http://127.0.0.1:8888"
// $Env:HTTPS_PROXY="http://127.0.0.1:8888"
// $Env:PAN123_ACCESS_TOKEN=""
// $Env:PAN123_CLIENT_ID="YOUR_CLIENT_ID"
// $Env:PAN123_CLIENT_SECRET="YOUR_CLIENT_SECRET"
// go test -v ./pan123
package pan123

import (
	"crypto/rand"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

var pan123TestInstance = NewPan123(0, false)

var pan123TestInstanceDirID int64 = 0
var pan123TestInstanceFileID int64 = 0
var pan123TestInstanceSmallFileID int64 = 0
var pan123TestFilePath string
var pan123TestSmallFilePath string

func _TestRequestAccessToken(t *testing.T) {
	accessToken, accessTokenExpiredAt, err := pan123TestInstance.RequestAccessToken(os.Getenv("PAN123_CLIENT_ID"), os.Getenv("PAN123_CLIENT_SECRET"))
	if err != nil {
		t.Fatal(err)
	}
	pan123TestInstance.SetAccessToken(accessToken)
	t.Logf("accessTokenExpiredAt = %s", accessTokenExpiredAt)
}

func _TestMkDir(t *testing.T) {
	resp, err := pan123TestInstance.MkDir("go_sdk_unit_test", 0)
	if err != nil {
		t.Fatal(err)
	}
	pan123TestInstanceDirID = resp.DirID
}

func _TestUploadFile(t *testing.T) {
	file, err := os.OpenFile(pan123TestFilePath, os.O_RDONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	statusCB := func(info FileUploadCallbackInfo) {
		t.Log(info)
	}
	resp, err := pan123TestInstance.FileUploadWithCallback(pan123TestInstanceDirID, "go_sdk_unit_test_test_upload.txt", file, 23, statusCB)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Async {
		// 需要等待异步上传
		t.Logf("wait for async upload")
		for {
			resp2, err := pan123TestInstance.GetUploadAsyncResult(resp.PreuploadID)
			if err != nil {
				t.Fatal(err)
			}
			if resp2.Completed {
				pan123TestInstanceFileID = resp2.FileID
				break
			}
			time.Sleep(2 * time.Second)
		}
	} else {
		pan123TestInstanceFileID = resp.FileID
	}
}

func _TestUploadSmallFile(t *testing.T) {
	file, err := os.OpenFile(pan123TestSmallFilePath, os.O_RDONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	statusCB := func(info FileUploadCallbackInfo) {
		t.Log(info)
	}
	resp, err := pan123TestInstance.FileUploadWithCallback(pan123TestInstanceDirID, "go_sdk_unit_test_test_upload_small.txt", file, 23, statusCB)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Async {
		// 需要等待异步上传
		t.Logf("wait for async upload")
		for {
			resp2, err := pan123TestInstance.GetUploadAsyncResult(resp.PreuploadID)
			if err != nil {
				t.Fatal(err)
			}
			if resp2.Completed {
				pan123TestInstanceSmallFileID = resp2.FileID
				break
			}
			time.Sleep(2 * time.Second)
		}
	} else {
		pan123TestInstanceSmallFileID = resp.FileID
	}
}

func _TestMoveFile(t *testing.T) {
	err := pan123TestInstance.MoveFile([]int64{pan123TestInstanceFileID}, 0)
	if err != nil {
		t.Fatal(err)
	}
	// 再移回去, 下面要用
	err = pan123TestInstance.MoveFile([]int64{pan123TestInstanceFileID}, pan123TestInstanceDirID)
	if err != nil {
		t.Fatal(err)
	}
}

func _TestRenameFile(t *testing.T) {
	err := pan123TestInstance.RenameFile([]string{strconv.FormatInt(pan123TestInstanceFileID, 10) + "|test_rename"})
	if err != nil {
		t.Fatal(err)
	}
}

func _TestGetFileDetail(t *testing.T) {
	_, err := pan123TestInstance.GetFileDetail(pan123TestInstanceFileID)
	if err != nil {
		t.Fatal(err)
	}
}

func _TestEnableDirectLink(t *testing.T) {
	_, err := pan123TestInstance.EnableDirectLink(pan123TestInstanceDirID)
	if err != nil {
		t.Fatal(err)
	}
}

func _TestGetDirectLinkUrl(t *testing.T) {
	_, err := pan123TestInstance.GetDirectLinkUrl(pan123TestInstanceFileID)
	if err != nil {
		t.Fatal(err)
	}
}

func _TestDisableDirectLink(t *testing.T) {
	_, err := pan123TestInstance.DisableDirectLink(pan123TestInstanceDirID)
	if err != nil {
		t.Fatal(err)
	}
}

func _TestGetFileList(t *testing.T) {
	_, err := pan123TestInstance.GetFileList(pan123TestInstanceDirID, 1, 99, "file_id", "asc", false, "")
	if err != nil {
		t.Fatal(err)
	}
}

func _TestGetFileListV2(t *testing.T) {
	_, err := pan123TestInstance.GetFileListV2(pan123TestInstanceDirID, 100, "", -1, -1)
	if err != nil {
		t.Fatal(err)
	}
}

func _TestTrashFile(t *testing.T) {
	err := pan123TestInstance.TrashFile([]int64{pan123TestInstanceFileID})
	if err != nil {
		t.Fatal(err)
	}
	pan123TestInstanceFileID = 0

	// 根据咨询官方技术人员, 该接口可同时删除文件夹
	err = pan123TestInstance.TrashFile([]int64{pan123TestInstanceDirID})
	if err != nil {
		t.Fatal(err)
	}
	pan123TestInstanceDirID = 0
}

func TestSequential(t *testing.T) {
	pan123TestInstance.SetAccessToken(os.Getenv("PAN123_ACCESS_TOKEN"))

	// 创建测试文件
	setupTestFile := func() error {
		// 123MB
		const fileSize = 123 * 1024 * 1024

		exePath, _err := os.Executable()
		if _err != nil {
			return _err
		}
		pan123TestFilePath = filepath.Join(filepath.Dir(exePath), "test_123mb_file.txt")
		testFile, _err := os.Create(pan123TestFilePath)
		if _err != nil {
			return _err
		}
		defer testFile.Close()

		buffer := make([]byte, 8192)
		var totalWritten int64
		for totalWritten < fileSize {
			// 生成随机数据
			n, _err := rand.Read(buffer)
			if _err != nil {
				return _err
			}

			// 写入文件
			written, _err := testFile.Write(buffer[:n])
			if _err != nil {
				return _err
			}

			totalWritten += int64(written)
		}

		_err = testFile.Sync()
		if _err != nil {
			return _err
		}

		return nil
	}
	setupTestSmallFile := func() error {
		// 8KB
		const fileSize = 8 * 1024

		exePath, _err := os.Executable()
		if _err != nil {
			return _err
		}
		pan123TestSmallFilePath = filepath.Join(filepath.Dir(exePath), "test_8kb_file.txt")
		testFile, _err := os.Create(pan123TestSmallFilePath)
		if _err != nil {
			return _err
		}
		defer testFile.Close()

		buffer := make([]byte, 8192)
		var totalWritten int64
		for totalWritten < fileSize {
			// 生成随机数据
			n, _err := rand.Read(buffer)
			if _err != nil {
				return _err
			}

			// 写入文件
			written, _err := testFile.Write(buffer[:n])
			if _err != nil {
				return _err
			}

			totalWritten += int64(written)
		}

		_err = testFile.Sync()
		if _err != nil {
			return _err
		}

		return nil
	}

	err := setupTestFile()
	if err != nil {
		t.Fatalf("setupTestFile() error: %s", err)
	}
	err = setupTestSmallFile()
	if err != nil {
		t.Fatalf("setupTestSmallFile() error: %s", err)
	}

	defer func() {
		t.Logf("test ended, start clean..")
		if pan123TestFilePath != "" {
			if _, _err := os.Stat(pan123TestFilePath); _err == nil {
				_err := os.Remove(pan123TestFilePath)
				if _err != nil {
					t.Fatalf("clean failed: %s", _err)
				}
			} else if os.IsNotExist(_err) {
				t.Logf("%s not found?", pan123TestFilePath)
			} else {
				t.Fatalf("clean failed: %s", _err)
			}
		}
		if pan123TestSmallFilePath != "" {
			if _, _err := os.Stat(pan123TestSmallFilePath); _err == nil {
				_err := os.Remove(pan123TestSmallFilePath)
				if _err != nil {
					t.Fatalf("clean failed: %s", _err)
				}
			} else if os.IsNotExist(_err) {
				t.Logf("%s not found?", pan123TestSmallFilePath)
			} else {
				t.Fatalf("clean failed: %s", _err)
			}
		}
		if pan123TestInstanceDirID != 0 {
			_, _ = pan123TestInstance.DisableDirectLink(pan123TestInstanceDirID)
			_err := pan123TestInstance.TrashFile([]int64{pan123TestInstanceDirID})
			if _err != nil {
				t.Fatalf("clean failed: %s", _err)
			}
		}
		if pan123TestInstanceFileID != 0 {
			_err := pan123TestInstance.TrashFile([]int64{pan123TestInstanceFileID})
			if _err != nil {
				t.Fatalf("clean failed: %s", _err)
			}
		}
		if pan123TestInstanceSmallFileID != 0 {
			_err := pan123TestInstance.TrashFile([]int64{pan123TestInstanceSmallFileID})
			if _err != nil {
				t.Fatalf("clean failed: %s", _err)
			}
		}
	}()

	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{"TestRequestAccessToken", _TestRequestAccessToken},
		{"TestMkDir", _TestMkDir},
		{"TestUploadFile", _TestUploadFile},
		{"TestUploadSmallFile", _TestUploadSmallFile},
		{"TestMoveFile", _TestMoveFile},
		{"TestRenameFile", _TestRenameFile},
		{"TestEnableDirectLink", _TestEnableDirectLink},
		{"TestGetDirectLinkUrl", _TestGetDirectLinkUrl},
		{"TestDisableDirectLink", _TestDisableDirectLink},
		{"TestGetFileList", _TestGetFileList},
		{"TestGetFileListV2", _TestGetFileListV2},
		{"TestGetFileDetail", _TestGetFileDetail},
		{"TestTrashFile", _TestTrashFile},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.test)
		if t.Failed() {
			t.Logf("Test %s failed, stopping subsequent tests.", tc.name)
			break
		}
	}
}
