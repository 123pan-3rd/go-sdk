package pan123

import (
	"crypto/rand"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var pan123TestInstance = NewPan123("", "333af92cb8234e28bc62d035b610f2fd", "c20b486adbd040bf925ddd5434f00f82", 0, false)

var pan123TestInstanceDirID int64 = 0
var pan123TestInstanceFileID int64 = 0
var pan123TestFilePath string

func _TestLogin(t *testing.T) {
	err := pan123TestInstance.Login()
	if err != nil {
		t.Fatal(err)
	}
	accessTokenExpiredAt, err := pan123TestInstance.GetAccessTokenExpiredAt()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("accessTokenExpiredAt = %s", accessTokenExpiredAt)
}

func _TestMkDir(t *testing.T) {
	resp, _, err := pan123TestInstance.MkDir("go_sdk_unit_test", 0)
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
	resp, _, err := pan123TestInstance.FileUpload(pan123TestInstanceDirID, "go_sdk_unit_test_test_upload.txt", file)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Async {
		// 需要等待异步上传
		t.Logf("wait for async upload")
		for {
			resp2, _, err := pan123TestInstance.GetUploadAsyncResult(resp.PreuploadID)
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

func _TestMoveFile(t *testing.T) {
	_, err := pan123TestInstance.MoveFile([]int64{pan123TestInstanceFileID}, 0)
	if err != nil {
		t.Fatal(err)
	}
	// 再移回去, 下面要用
	_, err = pan123TestInstance.MoveFile([]int64{pan123TestInstanceFileID}, pan123TestInstanceDirID)
	if err != nil {
		t.Fatal(err)
	}
}

func _TestEnableDirectLink(t *testing.T) {
	_, _, err := pan123TestInstance.EnableDirectLink(pan123TestInstanceDirID)
	if err != nil {
		t.Fatal(err)
	}
}

func _TestGetDirectLinkUrl(t *testing.T) {
	_, _, err := pan123TestInstance.GetDirectLinkUrl(pan123TestInstanceFileID)
	if err != nil {
		t.Fatal(err)
	}
}

func _TestDisableDirectLink(t *testing.T) {
	_, _, err := pan123TestInstance.DisableDirectLink(pan123TestInstanceDirID)
	if err != nil {
		t.Fatal(err)
	}
}

func _TestGetFileList(t *testing.T) {
	_, _, err := pan123TestInstance.GetFileList(pan123TestInstanceDirID, 1, 99, "file_id", "asc", false, "")
	if err != nil {
		t.Fatal(err)
	}
}

func _TestTrashFile(t *testing.T) {
	_, err := pan123TestInstance.TrashFile([]int64{pan123TestInstanceFileID})
	if err != nil {
		t.Fatal(err)
	}
	pan123TestInstanceFileID = 0

	// 根据咨询官方技术人员, 该接口可同时删除文件夹
	_, err = pan123TestInstance.TrashFile([]int64{pan123TestInstanceDirID})
	if err != nil {
		t.Fatal(err)
	}
	pan123TestInstanceDirID = 0
}

func TestSequential(t *testing.T) {

	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{"TestLogin", _TestLogin},
		{"TestMkDir", _TestMkDir},
		{"TestUploadFile", _TestUploadFile},
		{"TestMoveFile", _TestMoveFile},
		{"TestEnableDirectLink", _TestEnableDirectLink},
		{"TestGetDirectLinkUrl", _TestGetDirectLinkUrl},
		{"TestDisableDirectLink", _TestDisableDirectLink},
		{"TestGetFileList", _TestGetFileList},
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

func TestMain(m *testing.M) {
	t := &testing.T{}

	// 创建测试文件
	setupTestFile := func() error {
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
	err := setupTestFile()
	if err != nil {
		t.Fatalf("setupTestFile() error: %s", err)
	}

	code := m.Run()

	clean := func() {
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
		if pan123TestInstanceDirID != 0 {
			_, _, _ = pan123TestInstance.DisableDirectLink(pan123TestInstanceDirID)
			_, _err := pan123TestInstance.TrashFile([]int64{pan123TestInstanceDirID})
			if _err != nil {
				t.Fatalf("clean failed: %s", _err)
			}
		}
		if pan123TestInstanceFileID != 0 {
			_, _err := pan123TestInstance.TrashFile([]int64{pan123TestInstanceFileID})
			if _err != nil {
				t.Fatalf("clean failed: %s", _err)
			}
		}
	}
	clean()

	os.Exit(code)
}
