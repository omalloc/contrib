package runtime_test

import (
	"os/user"
	"testing"

	"github.com/omalloc/contrib/x/runtime"
)

func TestSetCurrentUser(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{
			name:     "空用户名应该直接返回nil",
			username: "",
			wantErr:  false,
		},
		{
			name:     "当前用户名相同应该返回nil",
			username: getCurrentUsername(t),
			wantErr:  false,
		},
		{
			name:     "不存在的用户名应该返回错误",
			username: "nonexistentuser123456789",
			wantErr:  true,
		},
		{
			name:     "切换用户为daemon",
			username: "daemon",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := runtime.SetCurrentUser(tt.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetCurrentUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseUser(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantUid  uint
		wantGid  uint
		wantErr  bool
	}{
		{
			name:     "有效用户",
			username: "root", // 使用一个确定存在的用户
			wantUid:  0,      // root用户通常是0
			wantGid:  0,      // root组通常是0
			wantErr:  false,
		},
		{
			name:     "无效用户",
			username: "nonexistentuser123",
			wantUid:  0,
			wantGid:  0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uid, gid, err := runtime.ParseUser(tt.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if uid != tt.wantUid {
					t.Errorf("ParseUser() uid = %v, want %v", uid, tt.wantUid)
				}
				if gid != tt.wantGid {
					t.Errorf("ParseUser() gid = %v, want %v", gid, tt.wantGid)
				}
			}
		})
	}
}

// 辅助函数：获取当前用户名
func getCurrentUsername(t *testing.T) string {
	current, err := user.Current()
	if err != nil {
		t.Fatalf("无法获取当前用户: %v", err)
	}
	return current.Username
}
