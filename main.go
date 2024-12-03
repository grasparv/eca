package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/grasparv/xflag/v2"
)

type EncryptCmd struct {
	XFlag     string `xflag:"encrypt|Encrypt a directory"`
	Directory string `xflag:"Path to directory"`
	Passfile  string `xflag:"Secret passphrase file"`
	Remove    *bool  `xflag:"true|Remove source directory on success"`
}

type DecryptCmd struct {
	XFlag    string `xflag:"decrypt|Decrypt a file"`
	File     string `xflag:"Path to encrypted file"`
	Passfile string `xflag:"Secret passphrase file"`
	Remove   *bool  `xflag:"true|Remove source file on success"`
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

	var dst string
	var src string
	var passfile string
	var fn func(string, string, []byte) error

	switch cmd := cmd.(type) {
	case *EncryptCmd:
		fn = encrypt
		src = cmd.Directory
		dst = strings.TrimSuffix(cmd.Directory, "/") + ".bin"
		dst = filepath.Clean(dst)
		passfile = cmd.Passfile
	case *DecryptCmd:
		fn = decrypt
		src = cmd.File
		dst = strings.TrimSuffix(filepath.Base(cmd.File), ".bin")
		dst = filepath.Clean(dst)
		passfile = cmd.Passfile
	default:
		os.Stderr.WriteString(xflag.GetUsage(commands))
		os.Exit(1)
	}

	_, err = os.Stat(src)
	if errors.Is(err, os.ErrNotExist) {
		logger.Error("does not exist", "src", src)
		os.Exit(1)
	}

	_, err = os.Stat(dst)
	if !errors.Is(err, os.ErrNotExist) {
		logger.Error("will not overwrite", "dst", dst)
		os.Exit(1)
	}

	logger = logger.With("src", src, "dst", dst)

	password, err := os.ReadFile(passfile)
	if err != nil {
		logger.Error("could not read password file", "error", err)
		os.Exit(1)
	}

	err = fn(src, dst, password)
	if err != nil {
		logger.Error("write fail", "error", err)
		err = os.RemoveAll(dst)
		if err == nil {
			logger.Info("cleaned up junk file")
		}
		os.Exit(1)
	}

	err = os.RemoveAll(src)
	if err != nil {
		logger.Error("remove source", "error", err)
	}
}
