package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	scp "github.com/bramvdbogaerde/go-scp"
	"golang.org/x/crypto/ssh"
)

func cleanup() {
	log.Println("cleanup.....")
	os.Exit(1)
}

func main() {

	signalChannel := make(chan os.Signal) // channel where interrupt signal is captured

	interruptC := make(chan bool, 1) // channel where true is send if interrupt signa happens

	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	// go routine to catch any signal interrupt, which in turn sends true to done channel
	go func() {
		signal := <-signalChannel
		fmt.Println("signal..", signal)
		interruptC <- true
	}()

	filesttoUpload := []string{"home/atri/Downloads/new3file.txt",
		"home/atri/Downloads/newfile.txt"}

	// Calls the uploadFiles
	fmt.Println("Uplaod files: ", uploadFiles(filesttoUpload, interruptC))

}

func uploadFiles(files []string, interruptC chan bool) bool {
	client := getSCPClient()

	for _, file := range files {
		select {
		case <-interruptC:
			fmt.Println("stopping the next upload if interrupt happens")
			cleanup()
		default:
			fmt.Println("Uploading file: ", file) // once the file upload is started interrupt will only stop the next fileupload
			isSuccess := uploadFile(file, client)
			if isSuccess {
				fmt.Println("Successfully uploaded file: ", file)
			}
		}
	}

	fmt.Println("Uplaoded all files successfully")
	return true
}

func uploadFile(fileLoc string, client scp.Client) bool {
	// Connect to the remote server
	conErr := client.Connect()
	if conErr != nil {
		fmt.Println("Couldn't establish a connection to the remote server ", conErr)
		return false
	}
	// Close client connection after the file has been copied
	defer client.Close()

	// Open a file
	f, _ := os.Open(fileLoc)

	// Close the file after it has been copied
	defer f.Close()
	filename := filepath.Base(fileLoc)
	err := client.CopyFile(f, "/home/ec2-user/testfiles/"+filename, "0655")
	if err != nil {
		log.Fatal("Error while copying file ", err)
	}
	return true
}

func getSCPClient() scp.Client {
	pemBytes, err := ioutil.ReadFile("/home/atri/Downloads/ec2test.pem")
	if err != nil {
		log.Fatal(err)
	}
	signer, err := ssh.ParsePrivateKey(pemBytes)

	if err != nil {
		log.Fatalf("parse key failed:%v", err)
	}

	clientConfig := &ssh.ClientConfig{
		User: "ec2-user",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	// Create a new SCP client
	client := scp.NewClient("ec2-18-116-69-58.us-east-2.compute.amazonaws.com:22", clientConfig)
	return client
}
