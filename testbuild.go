package main
 
import (
	"fmt"
	"log"
	"os"
	"time"

	"io/ioutil"
	"encoding/json"
)

type Configuration struct {
    Giturl    string
	Privkey   string
	Jailname  string
}

func connectSSH(username, hostname, port, giturl,
	keypath, jailname string) {
	// SSH client config
	key, err := ioutil.ReadFile(keypath)
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("unable to parse private key: %v", err)
	}
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		// Non-production only
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	

	// Connect to host
	client, err := ssh.Dial("tcp", hostname+":"+port, config)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
 
	// Create sesssion
	sess, err := client.NewSession()
	if err != nil {
		log.Fatal("Failed to create session: ", err)
	}
 
	// StdinPipe for commands
	stdin, err := sess.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	defer sess.Close()

 
	// Enable system stdout
	sess.Stdout = os.Stdout
	sess.Stderr = os.Stderr
 
	// Start remote shell
	err = sess.Shell()
	if err != nil {
		log.Fatal(err)
	}
	
	var commands []string
	// send the commands
	if hostname == "192.168.11.101" {
		commands = []string{
			fmt.Sprintf("yes | poudriere jail -d -j %s", jailname),
			fmt.Sprintf("poudriere jail -c -j %s -m git -b -v master -U %s -K GENERIC", jailname, giturl),
			fmt.Sprintf("poudriere bulk -j %s -p default -f pkglist", jailname),
			fmt.Sprintf("poudriere image -j %s -f pkglist -t tar -p default", jailname),
			"./script/check_charlie",
			"cd /b/tftpboot/FreeBSD/install",
			"chflags noschg lib/*",
			"rm -rf lib",
			"chflags noschg usr/lib32/*",
			"rm -rf usr/lib32",
			"chflags noschg libexec/*",
			"rm -rf libexec",
			"chflags noschg usr/bin/*",
			"rm -rf bin",
			"chflags noschg sbin/*",
			"rm -rf sbin",
			"chflags noschg var/*",
			"rm -rf var",
			"chflags noschg usr/lib/*",
			"rm -rf usr/lib",
			"rm -rf *",
			"rm -rf .*",
			"cp /data/images/poudriereimage.txz /b/tftpboot/FreeBSD/install",
			"tar -xf /data/images/poudriereimage.txz -C /b/tftpboot/FreeBSD/install",
			"mkdir -p /b/tftpboot/FreeBSD/install/conf/",
			"{ mkdir -p conf/base; tar -c -v -f conf/base/etc.cpio.gz --format cpio --gzip etc; tar -c -v -f conf/base/var.cpio.gz --format cpio --gzip var; } | chroot .",
			"cp -r ~/conf/default /b/tftpboot/FreeBSD/install/conf/",
			"mkdir /b/tftpboot/FreeBSD/install/root/.ssh",
			"cp ~/.ssh/id_rsa.pub /b/tftpboot/FreeBSD/install/root/.ssh/authorized_keys",
			"{ echo 1; echo \"on 6\"; sleep 1; } | telnet 192.168.11.99",
			"exit",
		}
	} else {
		commands = []string{
			"pwd_mkdb -p /etc/master.passwd",
			"cd /usr/tests",
			"kyua test",
			"exit",
		}
	}
	for _, cmd := range commands {
		_, err = fmt.Fprintf(stdin, "%s\n", cmd)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Wait for sess to finish
	err = sess.Wait()
}

func LoadConfiguration() Configuration {
	var config Configuration
    configFile, err := os.Open("gotest.json")
    defer configFile.Close()
    if err != nil {
        fmt.Println(err.Error())
    }
    jsonParser := json.NewDecoder(configFile)
    jsonParser.Decode(&config)
    return config
}
 
func main() {
	configuration := LoadConfiguration()
	connectSSH("root", "192.168.11.101", "22", 
		configuration.Giturl, configuration.Privkey, configuration.Jailname)
	print("Waiting for charlie to boot up!\n")
	time.Sleep(150 * time.Second)
	connectSSH("root", "192.168.11.4", "22", "", configuration.Privkey, "")
}
