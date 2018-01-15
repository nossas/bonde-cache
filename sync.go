package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

var (
	awsSecret = os.Getenv("AWS_SECRET_ACCESS_KEY")
	awsAccess = os.Getenv("AWS_ACCESS_KEY_ID")
)

func execCommand(command []string) {
	if len(command) < 2 {
		log.Println("Invalid Command")
	}
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Env = os.Environ()
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Print(err.Error())
	}
	log.Printf("%s\n", stdoutStderr)
}

func restoreCommand(what string, s Specification) {
	args := fmt.Sprintf("s3cmd sync --secret_key=%v --access_key=%v s3://bonde-cache/%v/%v/ ./data/%v/", awsSecret, awsAccess, s.Env, what, what)
	log.Print(args)

	command := strings.Split(args, " ")
	execCommand(command)
}

func updateCommand(what string, s Specification) {
	args := fmt.Sprintf("s3cmd sync --secret_key=%v --access_key=%v ./data/%v/ s3://bonde-cache/%v/%v/", awsSecret, awsAccess, what, s.Env, what)
	log.Print(args)

	command := strings.Split(args, " ")
	execCommand(command)
}

// Download DB from S3
func syncRestoreDb(s Specification) {
	restoreCommand("db", s)
}

// Upload DB to S3
func syncUpdateDb(s Specification) {
	updateCommand("db", s)
}

// Download certificates from S3
func syncRestoreCertificates(s Specification) {
	restoreCommand("certificates", s)
}

// Upload certificates to S3 before exit
func syncUpdateCertificates(s Specification) {
	updateCommand("certificates", s)
}
