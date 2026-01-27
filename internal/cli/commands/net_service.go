package commands

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
)

type netServiceImpl struct{}

func NewNetService() NetService {
	return &netServiceImpl{}
}

func (s *netServiceImpl) GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}

func (s *netServiceImpl) ListenTCP() (string, int, io.Closer, error) {
	listener, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		return "", 0, nil, err
	}
	addr := listener.Addr().(*net.TCPAddr)
	return addr.IP.String(), addr.Port, listener, nil
}

func (s *netServiceImpl) AcceptAndValidate(listener io.Closer, token string) (io.ReadCloser, error) {
	tcpListener, ok := listener.(net.Listener)
	if !ok {
		return nil, fmt.Errorf("invalid listener type")
	}

	conn, err := tcpListener.Accept()
	if err != nil {
		return nil, err
	}

	// Validate Token
	receivedToken := make([]byte, len(token))
	_, err = io.ReadFull(conn, receivedToken)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("error reading token: %w", err)
	}

	if string(receivedToken) != token {
		conn.Close()
		return nil, fmt.Errorf("invalid token received from worker")
	}

	return conn, nil
}

func (s *netServiceImpl) ExtractTarGz(r io.Reader, dest string) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dest, header.Name)

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
				f.Close()
				return err
			}
			f.Close()
		}
	}
	return nil
}
