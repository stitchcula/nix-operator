package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// AtomicWriteFile 原子性地写入文件
// content: 文件内容
// filename: 目标文件路径
// perm: 文件权限
func AtomicWriteFile(content []byte, filename string, perm os.FileMode) error {
	// 创建临时文件
	swap, err := os.CreateTemp(filepath.Dir(filename), filepath.Base(filename)+".tmp.*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	tmpName := swap.Name()
	defer func() {
		if err != nil {
			os.Remove(tmpName) // 如果有错误发生，清理临时文件
		}
	}()

	// 写入临时文件
	if _, err = swap.Write(content); err != nil {
		return fmt.Errorf("failed to write temp file: %v", err)
	}

	// 同步文件内容到磁盘
	if err = swap.Sync(); err != nil {
		return fmt.Errorf("failed to sync temp file: %v", err)
	}

	// 关闭临时文件
	if err = swap.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %v", err)
	}

	// 设置正确的权限
	if err = os.Chmod(tmpName, perm); err != nil {
		return fmt.Errorf("failed to chmod temp file: %v", err)
	}

	// 原子性地重命名临时文件
	if err = os.Rename(tmpName, filename); err != nil {
		return fmt.Errorf("failed to rename temp file: %v", err)
	}

	return nil
}
