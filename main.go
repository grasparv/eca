package main

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/grasparv/xflag/v2"
	"github.com/phsym/console-slog"
)

type EncryptCmd struct {
	XFlag      string `xflag:"encrypt|Encrypt a directory"`
	Directory  string `xflag:"Path to directory"`
	Passphrase string `xflag:"Secret passphrase"`
	Remove     *bool  `xflag:"true|Remove source directory on success"`
}

type DecryptCmd struct {
	XFlag      string `xflag:"decrypt|Decrypt a file"`
	File       string `xflag:"Path to encrypted file"`
	Passphrase string `xflag:"Secret passphrase"`
	Remove     *bool  `xflag:"true|Remove source file on success"`
}

func main() {
	commands := []interface{}{
		EncryptCmd{},
		DecryptCmd{},
	}

	cmd, err := xflag.Parse(commands, os.Args)
	if err != nil {
		os.Stderr.WriteString(err.Error())
		return
	}

	logger := makeLogger()

	switch cmd := cmd.(type) {
	case *EncryptCmd:
		name := cmd.Directory + ".bin"
		_, err := os.Stat(name)
		if !errors.Is(err, os.ErrNotExist) {
			logger.Error("will not overwrite", "name", name)
			os.Exit(1)
		}
		err = encrypt(cmd.Directory, name, cmd.Passphrase)
		if err != nil {
			logger.Error("write fail", "error", err)
			if os.RemoveAll(name) == nil {
				logger.Info("cleaned up junk file")
			}
			os.Exit(1)
		}
		err = os.RemoveAll(cmd.Directory)
		if err != nil {
			logger.Error("remove source", "error", err)
		}
	case *DecryptCmd:
		name := strings.TrimSuffix(filepath.Base(cmd.File), ".bin")
		_, err := os.Stat(name)
		if !errors.Is(err, os.ErrNotExist) {
			logger.Error("will not overwrite", "name", name)
			os.Exit(1)
		}
		err = decrypt(cmd.File, name, cmd.Passphrase)
		if err != nil {
			logger.Error("write fail", "error", err)
			if os.RemoveAll(name) == nil {
				logger.Info("cleaned up junk directory")
			}
			os.Exit(1)
		}
		err = os.RemoveAll(cmd.File)
		if err != nil {
			logger.Error("remove source", "error", err)
		}
	default:
		os.Stderr.WriteString(xflag.GetUsage(commands))
	}
}

func makeLogger() *slog.Logger {
	logger := slog.New(
		console.NewHandler(os.Stdout,
			&console.HandlerOptions{
				Level:      slog.LevelInfo,
				TimeFormat: time.TimeOnly,
			},
		),
	)

	return logger
}
