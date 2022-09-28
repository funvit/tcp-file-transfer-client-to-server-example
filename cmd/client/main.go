package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

var (
	logErr  = log.New(os.Stderr, "ERROR ", log.LstdFlags)
	logInfo = log.New(os.Stdout, "INFO  ", log.LstdFlags)
)

func main() {
	fmt.Println("TCP file transfer client (sender)")

	var addr string
	flag.StringVar(&addr, "a", "", "server address (ex: 0.0.0.0:8080)")
	flag.Usage = func() {
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "\nExample: %s -a 0.0.0.0:8080 myfile.txt\n\n", os.Args[0])
	}
	flag.Parse()

	fileName := flag.Arg(0)
	if fileName == "" {
		fmt.Println("File name required.")
		os.Exit(1)
	}
	if addr == "" {
		fmt.Println("Server address required.")
		os.Exit(1)
	}

	fileStat, err := os.Stat(fileName)
	if err != nil {
		logErr.Println("Get file stat:", err)
		os.Exit(1)
	}

	f, err := os.Open(fileName)
	if err != nil {
		logErr.Println("Open file:", err)
		os.Exit(1)
	}
	defer f.Close()

	logInfo.Printf("Sending file %q to server %q\n", fileStat.Name(), addr)
	err = sendFile(fileStat.Name(), fileStat.Size(), f, addr)
	if err != nil {
		logErr.Println("Send file:", err)
		os.Exit(2)
	}

	fmt.Println("Done")
}

func sendFile(fileName string, fileSize int64, r io.Reader, addr string) error {

	const bufferSize = 8 * 1024

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("connect to server: %w", err)
	}

	var sizeB [9]byte
	binary.BigEndian.PutUint64(sizeB[:], uint64(fileSize))

	sizeB[8] = '\n'

	_, err = conn.Write(sizeB[:])
	if err != nil {
		return fmt.Errorf("write file size: %w", err)
	}

	_, err = fmt.Fprintln(conn, fileName)
	if err != nil {
		return fmt.Errorf("write file name: %w", err)
	}

	var b [bufferSize]byte
	for {
		n, err := r.Read(b[:])
		if n > 0 {
			_, err := conn.Write(b[:n])
			if err != nil {
				return fmt.Errorf("write buffer to connection: %w", err)
			}
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read file: %w", err)
		}
	}
}
