package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
)

var (
	logErr  = log.New(os.Stderr, "ERROR ", log.LstdFlags)
	logInfo = log.New(os.Stdout, "INFO  ", log.LstdFlags)
)

func main() {
	fmt.Println("TCP file transfer server (receiver)")

	var destPath string
	flag.StringVar(&destPath, "dir", "", "destination path for received files")

	flag.Usage = func() {
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "\nExample: %s 0.0.0.0:8080\n\n", os.Args[0])
	}
	flag.Parse()

	addr := flag.Arg(0)
	if addr == "" {
		fmt.Println("Listen address required.")
		os.Exit(1)
	}

	logInfo.Printf("Listening on %q", addr)
	srv, err := net.Listen("tcp", addr)
	if err != nil {
		logErr.Println("listen:", err)
		os.Exit(2)
	}

	for {
		conn, err := srv.Accept()
		if err != nil {
			logErr.Println("accept connection:", err)
			os.Exit(2)
		}

		logInfo.Printf("Incoming connection from %q", conn.RemoteAddr().String())

		go func() {
			err := handleConn(conn, destPath)
			if err != nil {
				logErr.Println("handle conn:", err)
			}
		}()
	}
}

func handleConn(conn net.Conn, destPath string) error {

	const bufferSize = 8 * 1024

	r := bufio.NewReaderSize(conn, bufferSize)
	defer conn.Close()

	fileSizeB, err := r.ReadBytes('\n')
	if err != nil {
		return fmt.Errorf("read file size: %w", err)
	}

	fileSize := binary.BigEndian.Uint64(fileSizeB)
	fileName, err := r.ReadString('\n')
	if err != nil {
		return fmt.Errorf("read file name: %w", err)
	}

	fileName = strings.TrimSpace(fileName)

	if fileName == "" {
		return errors.New("empty file name")
	}

	logInfo.Printf("Creating file %q with size %d ", fileName, fileSize)
	f, err := os.Create(filepath.Join(destPath, fileName))
	if err != nil {
		return fmt.Errorf("create file %q: %s", fileName, err)
	}
	defer f.Close()

	var b [bufferSize]byte

	var total uint64
	for {
		n, err := r.Read(b[:])
		if n > 0 {
			if total+uint64(n) > fileSize {
				n = int(fileSize - total)
				total = fileSize
			} else {
				total += uint64(n)
			}

			_, err := f.Write(b[:n])
			if err != nil {
				return fmt.Errorf("write file data: %w", err)
			}

			if total == fileSize {
				return nil
			}
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read data from conn: %w", err)
		}
	}
}
