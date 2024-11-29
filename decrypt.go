package main

import (
	"archive/tar"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"io"
	"os"
	"path/filepath"

	"github.com/golang/snappy"
	"golang.org/x/crypto/pbkdf2"
)

func decrypt(inputFile, dstDir, passphrase string) error {
	file, err := os.Open(inputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	salt := make([]byte, 16)
	if _, err := io.ReadFull(file, salt); err != nil {
		return err
	}

	key := pbkdf2.Key([]byte(passphrase), salt, 4096, 32, sha256.New)
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(file, iv); err != nil {
		return err
	}

	stream := cipher.NewCFBDecrypter(block, iv)
	reader := &cipher.StreamReader{S: stream, R: file}

	snappyReader := snappy.NewReader(reader)
	tarReader := tar.NewReader(snappyReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		path := filepath.Join(dstDir, header.Name)
		if header.Typeflag == tar.TypeDir {
			if err := os.MkdirAll(path, os.FileMode(header.Mode)); err != nil {
				return err
			}
		} else {
			if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
				return err
			}
			file, err := os.Create(path)
			if err != nil {
				return err
			}
			defer file.Close()
			if _, err := io.Copy(file, tarReader); err != nil {
				return err
			}
		}
	}

	return nil
}
