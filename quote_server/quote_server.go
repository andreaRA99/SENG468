package main

import (
	"bufio"
	"bytes"
	"crypto/aes"
	cryptoRand "crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	mathRand "math/rand"
	"net"
	"time"
)

func main() {
	bind := flag.String("bind", "localhost:4444", "host:port to listen on")
	flag.Parse()

	ln, err := net.Listen("tcp", *bind)
	if err != nil {
		log.Fatalln(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatalln(err)
		}

		log.Println("Client connected")
		go interact(conn)
	}
}

func interact(conn net.Conn) {
	defer conn.Close()

	rngSeed := make([]byte, 16)
	if _, err := cryptoRand.Read(rngSeed); err != nil {
		log.Printf("Error generating random seed: %s\n", err)
		return
	}

	rng, err := aes.NewCipher(rngSeed)
	if err != nil {
		// this should never happen
		panic(err)
	}

	counter := uint32(0)
	counterBuf := make([]byte, 16)

	weakRngSeed := make([]byte, 16)

	rng.Encrypt(weakRngSeed, counterBuf)
	counter += 1
	binary.LittleEndian.PutUint32(counterBuf[:4], counter)

	weakRng := mathRand.New(mathRand.NewSource(int64(binary.LittleEndian.Uint64(weakRngSeed[:8]))))

	scanner := bufio.NewScanner(conn)
	responseKeyBuf := make([]byte, 32)

	for scanner.Scan() {
		if counter >= 0xffffff00 {
			log.Println("Counter overflow")
			return
		}

		line := scanner.Bytes()
		words := bytes.SplitN(line, []byte(" "), 2)

		if len(words) != 2 {
			log.Printf("Error parsing client request: %v\n", words)
			return
		}

		sym := words[0]
		username := words[1]

		priceInCents := weakRng.Intn(30000 + 1)

		rng.Encrypt(responseKeyBuf[:16], counterBuf)
		counter += 1
		binary.LittleEndian.PutUint32(counterBuf[:4], counter)

		rng.Encrypt(responseKeyBuf[16:], counterBuf)
		counter += 1
		binary.LittleEndian.PutUint32(counterBuf[:4], counter)

		responseKey := base64.RawURLEncoding.EncodeToString(responseKeyBuf[:20])
		timestamp := time.Now().UnixMilli()

		response := fmt.Sprintf("%d.%02d,%s,%s,%d,%s\n", priceInCents/100, priceInCents%100, sym, username, timestamp, responseKey)

		conn.Write([]byte(response))
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading line from client: %s\n", err)
		return
	}

	log.Println("Client disconnected")
}
