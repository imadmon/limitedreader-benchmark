package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
)

var (
	serverAddress = "localhost:1238"
)

type readFunc func(io.ReadCloser) (int, error)
type writeFunc func(io.Writer) (int, error)

func receiveOnceTCPServer(rf readFunc) (int, error) {
	ln, err := net.Listen("tcp", serverAddress)
	if err != nil {
		fmt.Println("Server failed to start:", err)
		return 0, err
	}
	defer ln.Close()
	fmt.Println("Server listening on", serverAddress)

	conn, err := ln.Accept()
	if err != nil {
		fmt.Println("Failed to accept connection:", err)
		return 0, err
	}
	defer conn.Close()

	n, err := rf(conn)
	fmt.Printf("Server received %d bytes\n", n)
	return n, err
}

func sendTCPMessage(wf writeFunc) (int, error) {
	fmt.Println("Sending message to", serverAddress)
	conn, err := net.Dial("tcp", serverAddress)
	if err != nil {
		fmt.Println("Failed to connect:", err)
		return 0, err
	}
	defer conn.Close()

	n, err := wf(conn)
	fmt.Printf("Client sent %d bytes\n", n)
	return n, err
}

func saveDataToFile(data AllBenchmarkData, filename string) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Cannot create file: %v\n", err)
		return
	}
	defer file.Close()

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Cannot marshal data: %v\n", err)
		return
	}

	_, err = file.Write(jsonData)
	if err != nil {
		fmt.Printf("Cannot write data: %v\n", err)
	}

	fmt.Printf("Saved data to file: %v\n", filename)
}

func loadDataFromFile(filename string) (AllBenchmarkData, error) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Cannot open file: %v\n", err)
		return nil, err
	}
	defer file.Close()

	var data AllBenchmarkData
	err = json.NewDecoder(file).Decode(&data)
	if err != nil {
		fmt.Printf("Cannot unmarshal data: %v\n", err)
		return nil, err
	}

	fmt.Printf("Loaded data from file: %v\n", filename)
	return data, nil
}

func addNumberToFilename(filename string, number int) string {
	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filename, ext)
	return fmt.Sprintf("%s.%d%s", name, number, ext)
}
