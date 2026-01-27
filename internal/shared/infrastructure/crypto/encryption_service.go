package crypto

import (
	"archive/tar"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/crypto/hkdf"
)

type EncryptionService interface {
	Encrypt(plaintext []byte) ([]byte, error)
	Decrypt(ciphertext []byte) ([]byte, error)
}

type AESGCMEncryptionService struct {
	key []byte
}

func NewAESGCMEncryptionService(key string) (*AESGCMEncryptionService, error) {
	// Key length must be 16, 24, or 32 bytes for AES-128, AES-192, or AES-256
	k := []byte(key)
	if len(k) != 16 && len(k) != 24 && len(k) != 32 {
		return nil, fmt.Errorf("invalid key length: %d. Must be 16, 24, or 32 bytes", len(k))
	}

	return &AESGCMEncryptionService{
		key: k,
	}, nil
}

func (s *AESGCMEncryptionService) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func (s *AESGCMEncryptionService) Decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("malformed ciphertext")
	}

	nonce, ciphertext := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func DeriveKey(masterKey string, backupID string) ([]byte, error) {
	hash := sha256.New
	hk := hkdf.New(hash, []byte(masterKey), []byte(backupID), nil)
	key := make([]byte, 32) // AES-256
	if _, err := io.ReadFull(hk, key); err != nil {
		return nil, err
	}
	return key, nil
}

func EncryptFile(source string, target string, key []byte) error {
	inFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer func() {
		if err := inFile.Close(); err != nil {
			log.Printf("WARNING: Failed to close input file %s: %v", source, err)
		}
	}()

	outFile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer func() {
		if err := outFile.Close(); err != nil {
			log.Printf("WARNING: Failed to close output file %s: %v", target, err)
		}
	}()

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	if _, err := outFile.Write(nonce); err != nil {
		return err
	}

	plaintext, err := io.ReadAll(inFile)
	if err != nil {
		return err
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
	_, err = outFile.Write(ciphertext)
	return err
}

func DecryptFile(source string, target string, key []byte) error {
	inFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer func() {
		if err := inFile.Close(); err != nil {
			log.Printf("WARNING: Failed to close input file %s: %v", source, err)
		}
	}()

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(inFile, nonce); err != nil {
		return err
	}

	ciphertext, err := io.ReadAll(inFile)
	if err != nil {
		return err
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return err
	}

	return os.WriteFile(target, plaintext, 0644)
}

func CompressDirectory(source string, target string) error {
	tarFile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer func() {
		if err := tarFile.Close(); err != nil {
			log.Printf("WARNING: Failed to close tar file %s: %v", target, err)
		}
	}()

	gw := gzip.NewWriter(tarFile)
	defer func() { _ = gw.Close() }()

	tw := tar.NewWriter(gw)
	defer func() { _ = tw.Close() }()

	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		header.Name = relPath

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer func() { _ = file.Close() }()
			_, err = io.Copy(tw, file)
			return err
		}
		return nil
	})
}

func DecompressTarGz(source string, targetDir string) error {
	file, err := os.Open(source)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("WARNING: Failed to close file %s: %v", source, err)
		}
	}()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer func() { _ = gzr.Close() }()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(targetDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				_ = f.Close()
				return err
			}
			_ = f.Close()
		}
	}
	return nil
}
