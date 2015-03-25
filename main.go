package main

import (
	"html/template"
	"leyserSFX/api"
	"log"
	"net/http"
)

func main() {
	//router
	http.HandleFunc("/create", CreateSFXHandler)
	http.Handle("/vendor/", StaticHandler())
	http.HandleFunc("/", IndexHandler)

	//listen and serve
	http.ListenAndServe(":8888", nil)
}
func StaticHandler() http.Handler {
	log.Println("enter static func")
	return http.FileServer(http.Dir("./static/"))

}

//index html
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("enter index func")
	t, err := template.ParseFiles("SFX.html")
	if err != nil {
		log.Println("load template fail.", err.Error())
		return
	}
	t.Execute(w, nil)
}

//create sfx files
func CreateSFXHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20)

	cfg := GetConfig(r)
	log.Println("the cfg is", cfg)

	formdata := r.MultipartForm
	log.Println("the MultipartForm is", formdata)

	files := formdata.File
	log.Println("the files is", files)
	// check files type
	if api.CheckFile(files) == false {
		log.Println("check files type fail.")
		w.Write([]byte("invalid file"))
		return
	}
	//create directory
	thePath := api.CreateDirectory(cfg)
	//copy new dll
	for _, header := range files["fileUploader1"] {
		if api.CopyFile(w, header, cfg, thePath[0]) == false {
			return
		}
	}
	//copy old dll
	for _, header := range files["fileUploader2"] {
		if api.CopyFile(w, header, cfg, thePath[1]) == false {
			return
		}
	}

	log.Println("the origin config is", cfg)
	//create sfx
	api.CreateSFX(cfg)

	//Signature
	api.Signature(cfg)
	//download
	api.DownloadFile(w)
}

//get the config
func GetConfig(r *http.Request) *api.Config {
	var cfg api.Config
	cbxASP := r.FormValue("cbxASP")
	cbxCS := r.FormValue("cbxCS")
	location := r.FormValue("location")
	log.Println("the cbxasp value is", cbxASP)
	log.Println("the cbxcs valus is", cbxCS)
	log.Println("the location value is", location)
	//client mode
	if cbxASP == "on" && cbxCS == "on" {
		cfg.ClientMode = 3
	} else if cbxASP == "on" {
		cfg.ClientMode = 1
	} else if cbxCS == "on" {
		cfg.ClientMode = 2
	}
	//location
	if location == "client" {
		cfg.Location = 1
	} else if location == "server" {
		cfg.Location = 2
	}
	return &cfg
}
