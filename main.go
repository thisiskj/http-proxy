package main

import (
	"bufio"
	"io/ioutil"
	"net"
	"strings"

	log "github.com/sirupsen/logrus"
)

func main() {
	log.Println("Starting proxy on 8080...")

	// Frontend
	ln, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		log.Fatal(err)
	}

	for {
		client, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go handleConnection(client)

		// Backend
		// backend, err := net.Dial("tcp", "127.0.0.1:8081")
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// Read from backend, send to frontend
		// for {
		// 	var message []byte
		// 	n, err := bufio.NewReader(backend).Read(message)
		// 	if err == io.EOF && n == 0 {
		// 		// log.Println("F <- B: " + err.Error())
		// 		frontend.Close()
		// 		backend.Close()
		// 		break
		// 	}
		//
		// 	if n > 0 {
		// 		log.Print("F <- B:", string(message))
		// 		frontend.Write(message)
		// 	}
		//
		// }
	}

}

func handleConnection(client net.Conn) {
	log.WithFields(log.Fields{
		"client": client.RemoteAddr().String(),
	}).Info("New connection")

	req := readRequest(client)

	log.WithFields(log.Fields{
		"client": client.RemoteAddr().String(),
		"req":    req,
	}).Info("Got Req")

	res := sendToBackend(client, req)

	log.WithFields(log.Fields{
		"client": client.RemoteAddr().String(),
		"res":    string(res),
	}).Info("Got Res")

	sendToClient(client, res)

}

type Request struct {
	Method      string
	Location    string
	HTTPVersion string
	Headers     []Header
	Raw         []byte
	RawHeaders  []byte
}

type Header struct {
	Key   string
	Value string
}

func readRequest(client net.Conn) Request {
	var req Request

	reader := bufio.NewReader(client)
	lines := 0

	for {
		message, err := reader.ReadBytes('\n')
		if err != nil {
			log.WithFields(log.Fields{"client": client.RemoteAddr().String()}).Info("Disconnect connection")
			client.Close()
			break
		}

		log.WithFields(log.Fields{"client": client.RemoteAddr().String(), "line": string(message)}).Info("Received line")

		req.Raw = append(req.Raw, message...)

		messageString := string(message)
		if lines == 0 {
			parts := strings.Split(messageString, " ")
			req.Method = parts[0]
			req.Location = parts[1]
			req.HTTPVersion = parts[2]
		} else if string(message) == "\r\n" {
			req.RawHeaders = append(req.RawHeaders, message...)
			log.WithFields(log.Fields{"client": client.RemoteAddr().String()}).Info("Completed receiving headers")

			break
		} else {
			req.RawHeaders = append(req.RawHeaders, message...)
			parts := strings.Split(messageString, ":")
			req.Headers = append(req.Headers, Header{
				Key:   parts[0],
				Value: parts[1],
			})
		}

		lines++
	}

	return req
}

// Attempt 2: Byte by byte
// func sendToBackend(client net.Conn, req Request) []byte {
//
// 	res := make([]byte, 0)
//
// 	// Backend
// 	backend, err := net.Dial("tcp", "127.0.0.1:8081")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
//
// 	reader := bufio.NewReader(backend)
//
// 	_, err = backend.Write([]byte(req.Method + " " + req.Location + " HTTP/1.0\r\n"))
// 	if err != nil {
// 		log.Fatal(err)
// 	}
//
// 	_, err = backend.Write(req.RawHeaders)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
//
// 	for {
// 		b, err := reader.ReadByte()
// 		if err != nil {
// 			break
// 		}
//
// 		res = append(res, b)
// 	}
//
// 	return res
// }

// Attempt 1: Readall
func sendToBackend(client net.Conn, req Request) []byte {

	res := make([]byte, 0)

	// Backend
	backend, err := net.Dial("tcp", "www.google.com:80")
	if err != nil {
		log.Fatal(err)
	}

	reader := bufio.NewReader(backend)

	_, err = backend.Write([]byte(req.Method + " " + req.Location + " HTTP/1.0\r\n"))
	if err != nil {
		log.Fatal(err)
	}

	_, err = backend.Write(req.RawHeaders)
	if err != nil {
		log.Fatal(err)
	}

	res, err = ioutil.ReadAll(reader)
	if err != nil {
		log.Fatal(err)
	}

	return res
}

func sendToClient(client net.Conn, res []byte) {
	client.Write(res)
	client.Close()
}
