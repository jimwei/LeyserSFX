package api

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	RAR    = `C:\Program Files\WinRAR\WinRAR.exe`
	CSEXE  = `SFX\CS_QuickDeployment.exe`
	ASPEXE = `SFX\ASP_QuickDeployment.exe`
	ARGS0  = `a -sfx -z%s\sfx.txt`
	BAT    = "sfx.bat"
)

//the user setting for quickdepyoment
type Config struct {
	//1: client;2:server
	Location int
	//1:asp;2:cs;3:asp first,then cs
	ClientMode int
}
type ExeConfig struct {
	XMLName     xml.Name    `xml:"configuration"`
	Appsettings Appsettings `xml:"appSettings"`
}

type Appsettings struct {
	Comment string `xml:",comment"`
	Adds    []Add  `xml:"add"`
}

type Add struct {
	Key   string `xml:"key,attr"`
	Value string `xml:"value,attr"`
}

var (
	BasePath string
)

func init() {
	//set the current work directory
	setBasePath()
}

//create the sfx exe
func CreateSFX(cfg *Config) {
	//create the sfx exe
	createSFXCore(cfg)
}

//set the base path
func setBasePath() {
	var file = os.Args[0]
	fmt.Println("the args[0] is", file)
	thePath, _ := exec.LookPath(file)
	BasePath = filepath.Dir(thePath)

}

//modify the config
func writeConfig(cfg *Config) {
	cName := filepath.Join(BasePath, `File\QuickDeployment.exe.config`)
	fmt.Println("the cname is", cName)
	sb, _ := ioutil.ReadFile(cName)
	var execfg ExeConfig
	err := xml.Unmarshal(sb, &execfg)
	if err != nil {
		fmt.Println("unmarshal exe config fail.", err.Error())
		return
	}

	execfg.Appsettings.Adds[0].Value = strconv.Itoa(cfg.Location)
	execfg.Appsettings.Adds[1].Value = strconv.Itoa(cfg.ClientMode)
	fmt.Println("the modifyed app is", execfg)
	//marshal
	newb, err := xml.Marshal(&execfg)
	if err != nil {
		fmt.Println("marshal fail", err.Error())
	}
	ioutil.WriteFile(cName, newb, os.ModePerm)
}

//reate the sfx exe binary
func createSFXCore(cfg *Config) {
	//first delete the sfx directory and the older sfx files
	os.Remove(filepath.Join(BasePath, CSEXE))
	os.Remove(filepath.Join(BasePath, ASPEXE))
	argsList := GetRarArgs(cfg)
	log.Println("the args list length", len(argsList))
	for _, arg := range argsList {
		log.Println("the arg is", arg)

		//write config
		if strings.Index(arg, ASPEXE) > 0 {
			//asp model
			cfg.ClientMode = 1
			log.Println("write asp config")
			writeConfig(cfg)
		} else if strings.Index(arg, CSEXE) > 0 {
			//cs model
			cfg.ClientMode = 2
			log.Println("write cs config")
			writeConfig(cfg)
		}
		args := strings.Split(arg, " ")

		cmd := exec.Command(RAR, args...)

		if err := cmd.Run(); err != nil {
			log.Println(err)
			return
		}
		log.Println("create sfx successfully!")
		time.Sleep(1000)
	}
	time.Sleep(1000)
}

//download the sfx files
func DownloadFile(w http.ResponseWriter) {
	//delete the older zip file
	os.Remove(filepath.Join(BasePath, `SFX.zip`))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Transfer-Encoding", "binary")
	w.Header().Set("Content-Disposition", "attachment;filename='SFX.zip'")

	err := exec.Command(RAR, "a", "SFX.zip", "-ep1", filepath.Join(BasePath, `SFX\*`)).Run()
	if err != nil {
		log.Println("zip files fail.", err.Error())
		return
	}
	f, err := os.Open(filepath.Join(BasePath, `SFX.zip`))
	if err != nil {
		log.Println("read the exe file fail.", err.Error())
		return
	}
	defer f.Close()
	content, err := ioutil.ReadAll(f)
	if err != nil {
		log.Println("*read the exe file fail.", err.Error())
		return
	}

	w.Write(content)
}

//generate the rar sfx args
func GetRarArgs(cfg *Config) []string {
	argsList := []string{}

	argsFileDir := `%s\File`
	//get the exe name
	exeNames := GetExeNames(cfg)
	var args1 = ` %s -ep1 %s`
	for _, v := range exeNames {
		arg := fmt.Sprintf(strings.Join([]string{fmt.Sprintf(ARGS0, BasePath), args1}, ""), v, fmt.Sprintf(argsFileDir, BasePath))
		argsList = append(argsList, arg)
	}

	return argsList
}

// get the exe name depend on the client mode
func GetExeNames(cfg *Config) []string {
	names := []string{}
	csName := filepath.Join(BasePath, CSEXE)
	aspName := filepath.Join(BasePath, ASPEXE)
	if cfg.ClientMode == 3 {
		//need produce cs sfx and asp sfx
		names = append(names, aspName)
		names = append(names, csName)
	} else if cfg.ClientMode == 1 {
		names = append(names, aspName)
	} else if cfg.ClientMode == 2 {
		names = append(names, csName)
	}
	log.Println("the exenames is", names)
	return names
}

//create the dll directory
//return the directory stored new dll and the old path sored origin dll
func CreateDirectory(cfg *Config) []string {
	//delete the directory
	deleteDirectory()

	var thePath string = ""
	var theOldPath string = ""
	if cfg.Location == 1 {
		//client
		thePath = filepath.Join(BasePath, `File\Deployment\New\Kaikei`)
		theOldPath = filepath.Join(BasePath, `File\Deployment\Old\Kaikei`)
		//need delete the server directory
	} else if cfg.Location == 2 {
		//server
		thePath = filepath.Join(BasePath, `File\Deployment\New`)
		theOldPath = filepath.Join(BasePath, `File\Deployment\Old`)
		//need delete the client kaikei directory
	}
	_, err := os.Stat(thePath)
	if err != nil {
		//no the directory,so create the directory
		os.MkdirAll(thePath, os.ModePerm)
	}
	_, err = os.Stat(theOldPath)
	if err != nil {
		//no the directory,so create the directory
		os.MkdirAll(theOldPath, os.ModePerm)
	}
	return []string{thePath, theOldPath}
}

//delete the upload directory
func deleteDirectory() {
	thePath := filepath.Join(BasePath, `File\Deployment`)
	log.Println("the delete path is", thePath)
	err := os.RemoveAll(thePath)
	if err != nil {
		log.Println("delete the directory fail.", err.Error())
		return
	}

	//创建目录
	err = os.MkdirAll(thePath, os.ModePerm)
	if err != nil {

		log.Println("make deplyoment fail.", err.Error())
	}
}

//check the uploaded files whether are dll or not
func CheckFile(files map[string][]*multipart.FileHeader) bool {
	for _, headers := range files {
		for _, header := range headers {
			ext := filepath.Ext(header.Filename)
			if strings.ToLower(ext) != ".dll" {
				log.Println("the ext is", ext)
				log.Println("invalid file", header.Filename)
				return false
			}
		}
	}
	return true
}

//copy the uploaded file content
func CopyFile(w http.ResponseWriter, header *multipart.FileHeader, cfg *Config, thePath string) bool {
	//open the file
	file, err := header.Open()
	if err != nil {
		log.Println("open the file fail", err.Error())

		return false
	}
	defer file.Close()
	log.Println("the uploaded filename is", header.Filename)

	//create the file
	f, _ := os.OpenFile(filepath.Join(thePath, header.Filename), os.O_CREATE|os.O_WRONLY, os.ModePerm)
	defer f.Close()
	_, err = io.Copy(f, file)
	if err != nil {
		log.Println("upload fail.", err.Error())
		return false
	}
	return true
}
