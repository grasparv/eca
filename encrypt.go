package main

import (
	"archive/tar"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"io"
	"os"
	"path/filepath"

	"github.com/golang/snappy"
	"golang.org/x/crypto/pbkdf2"
)

func encrypt(srcDir, outputFile, passphrase string) error {
	output, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer output.Close()

	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return err
	}
	if _, err := output.Write(salt); err != nil {
		return err
	}

	key := pbkdf2.Key([]byte(passphrase), salt, 4096, 32, sha256.New)
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return err
	}
	if _, err := output.Write(iv); err != nil {
		return err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	writer := &cipher.StreamWriter{S: stream, W: output}

	snappyWriter := snappy.NewBufferedWriter(writer)
	defer snappyWriter.Close()

	tarWriter := tar.NewWriter(snappyWriter)
	defer tarWriter.Close()

	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}
		header.Name, err = filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			if _, err := io.Copy(tarWriter, file); err != nil {
				return err
			}
		}
		return nil
	})
}
