package settings

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/spf13/afero"
)

var (
	invalidFilenameChars = regexp.MustCompile(`[^0-9A-Za-z@_\-.]`)

	dashes = regexp.MustCompile(`[\-]+`)
)

// MakeUserDir makes the user directory according to settings.
func (s *Settings) MakeUserDir(username, userScope, serverRoot string) (string, error) {
	// 清理用户作用域路径
	userScope = strings.TrimSpace(userScope)
	// 当用户scope为空且启动 create user dir 的时候, 自动创建用户目录
	if userScope == "" && s.CreateUserDir {
		//1. 安全处理对应的用户名
		username = cleanUsername(username)
		if username == "" || username == "-" || username == "." {
			log.Printf("create user: invalid user for home dir creation: [%s]", username)
			return "", errors.New("invalid user for home dir creation")
		}
		// 2. 设置当前用户的scope 为 userName
		userScope = path.Join(s.UserHomeBasePath, username)
	}

	userScope = path.Join("/", userScope)
	// 创建对应的文件目录
	fs := afero.NewBasePathFs(afero.NewOsFs(), serverRoot)
	if err := fs.MkdirAll(userScope, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create user home dir: [%s]: %w", userScope, err)
	}
	return userScope, nil
}

// cleanUsername 清理用户名, 移除无效字符
func cleanUsername(s string) string {
	// Remove any trailing space to avoid ending on -
	s = strings.Trim(s, " ")
	s = strings.Replace(s, "..", "", -1)

	// Replace all characters which not in the list `0-9A-Za-z@_\-.` with a dash
	s = invalidFilenameChars.ReplaceAllString(s, "-")

	// Remove any multiple dashes caused by replacements above
	s = dashes.ReplaceAllString(s, "-")
	return s
}
