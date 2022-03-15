package helper

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

type DbDriverConf struct {
	Driver string
	Dsn    string
	Usr    string
	Pwd    string
}

var Config = struct {
	Workdir  string
	Database struct {
		Mysql   DbDriverConf
		Mongodb DbDriverConf
		Elastic struct {
			Host string `required:"true"`
			Auth string
		}
	}
}{}

func init() {
	file := "config.yaml"
	// test 临时处理
	if len(os.Args) >= 2 && strings.Contains(os.Args[1], "-test") {
		file = filepath.Join("../", file)
	}
	for i, s := range os.Args {
		if s == "-c" || s == "--config" {
			file = os.Args[i+1]
			break
		}
	}
	fmt.Println("config file:", file)
	data, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	yaml.Unmarshal(data, &Config)
}

func PrintConfig() {
	fmt.Printf("config: %#v", Config)
}
