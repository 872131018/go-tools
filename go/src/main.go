package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/mholt/archiver"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func main() {
	var project string = os.Getenv("PROJECT")
	var archive string = project + ".tgz"

	client, session, err := connectToHost(os.Getenv("USER"), os.Getenv("SERVER"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	fmt.Println("Creating TGZ of file")
	_, err = session.CombinedOutput("tar -C Sites/ -czf " + archive + " --exclude='.git' --exclude='node_modules' " + project)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Opening SFTP client")
	sftp, err := sftp.NewClient(client)
	if err != nil {
		log.Fatal(err)
	}
	defer sftp.Close()

	fmt.Println("Reading file to copy")
	srcFile, err := sftp.Open(archive)
	if err != nil {
		log.Fatal(err)
	}
	defer srcFile.Close()

	fmt.Println("Creating file to save to")
	dstFile, err := os.Create(archive)
	if err != nil {
		log.Fatal(err)
	}
	defer dstFile.Close()

	fmt.Println("Writing file to local")
	srcFile.WriteTo(dstFile)

	fmt.Println("Unarchiving project")
	err = archiver.Unarchive(archive, project)

	fmt.Println("Deleting archive")
	err = os.Remove(archive)
	if err != nil {
		log.Fatal(err)
	}

	client, session, err = connectToHost(os.Getenv("USER"), os.Getenv("SERVER"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	fmt.Println("Delete remote archive")
	_, err = session.CombinedOutput("rm ~/" + archive)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Configure remote-sync config")
	input, err := ioutil.ReadFile(".remote-sync.json")
	if err != nil {
		log.Fatalln(err)
	}
	lines := strings.Split(string(input), "\n")
	for i, line := range lines {
		if strings.Contains(line, "\"target\": \"\",") {
			lines[i] = "  \"target\": \"/home/jhuffman/Sites/" + project + "\","
		}
	}
	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(project+"/"+project+"/.remote-sync.json", []byte(output), 0644)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Moving finished project")
	cmd := exec.Command("mv", project+"/"+project, "/go/src/temp/")
	err = cmd.Run()
	/*
			err = os.Rename(project+"/"+project, "/go/src/temp/"+project)
			if err != nil {
		        log.Fatalln(err)
		    }
	*/
}

func connectToHost(user, host string) (*ssh.Client, *ssh.Session, error) {
	privateKeyBytes, err := ioutil.ReadFile("/go/src/.ssh/id_rsa")
	if err != nil {
		log.Fatal(err)
	}

	key, err := ssh.ParsePrivateKey(privateKeyBytes)
	if err != nil {
		log.Fatal(err)
	}

	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", host, sshConfig)
	if err != nil {
		return nil, nil, err
	}

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, nil, err
	}

	return client, session, nil
}
