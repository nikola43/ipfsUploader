package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
)

func UnmarshalIPFSUploadResult(data []byte) (IPFSUploadResult, error) {
	var r IPFSUploadResult
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *IPFSUploadResult) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type IPFSUploadResult struct {
	IpfsHash  *string `json:"IpfsHash,omitempty"`
	PinSize   *int64  `json:"PinSize,omitempty"`
	Timestamp *string `json:"Timestamp,omitempty"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
	JWT := os.Getenv("JWT")

	// get filepath from args
	if len(os.Args) < 2 {
		fmt.Println("Please provide a file path")
		return
	}
	filePath := os.Args[1]
	filename := filepath.Base(filePath)

	const baseURL = "https://coffee-magnetic-rattlesnake-502.mypinata.cloud/ipfs/"
	ipfsURL := "https://api.pinata.cloud/pinning/pinFileToIPFS"

	// Upload file to IPFS
	// New multipart writer.
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fw, err := writer.CreateFormFile("file", filename)
	if err != nil {
		panic(err)
	}
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	_, err = io.Copy(fw, file)
	if err != nil {
		panic(err)
	}
	writer.Close()
	req, err := http.NewRequest("POST", ipfsURL, bytes.NewReader(body.Bytes()))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+JWT)
	client := &http.Client{
		Timeout: time.Second * 3600,
	}
	rsp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	if rsp.StatusCode != http.StatusOK {
		log.Printf("Request failed with response code: %d", rsp.StatusCode)
	}
	body2 := &bytes.Buffer{}
	_, err = body2.ReadFrom(rsp.Body)
	if err != nil {
		panic(err)
	}
	rsp.Body.Close()
	uploadResult, err := UnmarshalIPFSUploadResult(body2.Bytes())
	if err != nil {
		panic(err)
	}
	fmt.Println("IPFS URL: ", baseURL+*uploadResult.IpfsHash)
}
