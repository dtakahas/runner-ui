package main

import (
  "html/template"
  "net/http"
  "log"
  "os"
  "bytes"
  "io"
  "io/ioutil"
  "encoding/json"
  "github.com/gorilla/schema"

)

var templates = template.Must(template.ParseFiles("templates/choose.html", "templates/edit.html", "templates/saved.html"))
var decoder = schema.NewDecoder()

// Server

func main() {
  log.Println("Started app on http://localhost:8080...ctrl-c to shut down.")
  http.HandleFunc("/", chooseFileHandler)
  http.HandleFunc("/open/", openFileHandler)
  http.HandleFunc("/edit/", editFileHandler)
  err := http.ListenAndServe(":8080", nil)
  if err != nil {
    log.Fatal(err)
  }
}

// Structs

type Page struct {
  Title string
}

type FileContent struct {
  Filename string
  Map      map[string]interface {}
}

type Config struct {
  ProjectGUID               string   `json:"projectGUID",schema:"projectGUID"`
  OrgId                     string   `json:"orgId",schema:"orgId"`
  ProjectId                 string   `json:"projectId",schema:"projectId"`
  UnityConnectEnvironnement string   `json:"unityConnectEnvironnement",schema:"unityConnectEnvironnement"`
  UnityEditorPath           string   `json:"unityEditorPath",schema:"unityEditorPath"`
  MinimumClientNumber       float64  `json:"minimumClientNumber",schema:"minimumClientNumber"`
  Host                      string   `json:"host",schema:"host"`
  Port                      float64  `json:"port",schema:"port"`
  Clients                   []Client `json:"clients",schema:"clients"`
  Tests                     []Test   `json:"tests",schema:"tests"`
  NewFileName               string   `json:"-",schema:"-"`
}

type Client struct {
  Email                     string    `json:"email",schema:"email"`
  LocalProjectFolder        string    `json:"localProjectFolder",schema:"localProjectFolder"`
  ReadyToJob                bool      `json:"readyToJob",schema:"readyToJob"`
  Socket                    string    `json:"socket",schema:"socket"`
  UserId                    string    `json:"userId",schema:"userId"`
  UserName                  string    `json:"userName",schema:"userName"`
  UserPassword              string    `json:"userPassword",schema:"userPassword"`
}

type Test struct {
  Commands                  []Command `json:"commands",schema:"commands"`
  CurrentCmd                string    `json:"currentCmd",schema:"currentCmd"`
  CurrentStep               float64   `json:"currentStep",schema:"currentStep"`
  Description               string    `json:"description",schema:"description"`
  RequiredClientNumber      float64   `json:"requiredClientNumber",schema:"requiredClientNumber"`
  TestId                    float64   `json:"testId",schema:"testId"`
}

type Command struct {
  CmdId                     float64   `json:"cmdId",schema:"cmdId"`
  AddFile                   string    `json:"addFile",schema:"addFile"`
  ClientId                  string    `json:"clientId",schema:"clientId"`
  Cmd                       string    `json:"cmd",schema:"cmd"`
  Completed                 bool      `json:"completed",schema:"completed"`
  Description               string    `json:"description",schema:"description"`
  ExecutionOrder            string    `json:"executionOrder",schema:"executionOrder"`
  Initialized               bool      `json:"initialized",schema:"initialized"`
  Action                    string    `json:"action",schema:"action"`
  Comment                   string    `json:"comment",schema:"comment"`
}

// Handlers

func chooseFileHandler(w http.ResponseWriter, r *http.Request) {
  p := &Page{Title: "Choose a test config"}
  renderTemplate(w, "choose", p)
}

func openFileHandler(w http.ResponseWriter, r *http.Request) {
  file, header, _ := r.FormFile("user_file")
  buf := bytes.NewBuffer(nil)
  io.Copy(buf, file)
  var i interface{}
  err := json.Unmarshal(buf.Bytes(), &i)
  if err != nil {
  	log.Fatal(err)
  }
  m := i.(map[string]interface{})
  fc := &FileContent{ Filename: header.Filename, Map: m }
  //debugUnmarshalledJson(m)
  file.Close()
  renderEditTemplate(w, "edit", fc)
}

func editFileHandler(w http.ResponseWriter, r *http.Request) {
  err := r.ParseMultipartForm(10000)
  if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

  m := r.MultipartForm.Value
  config := new(Config)
  decode_err := decoder.Decode(config, m)
  if decode_err != nil {
    http.Error(w, decode_err.Error(), http.StatusInternalServerError)
    return
  }

  j, json_err := json.MarshalIndent(config, "", "    ")
  if json_err != nil {
    http.Error(w, json_err.Error(), http.StatusInternalServerError)
    return
  }

  path, _ := os.Getwd()
  path += "/" + m["NewFileName"][0] + ".json"
  ioutil.WriteFile(path, j, 0664)

  p := &Page{Title: "A new json file has been saved to: " + path}
  renderTemplate(w, "saved", p)
}


// Helpers

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
  err := templates.ExecuteTemplate(w, tmpl+".html", p)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
}

func renderEditTemplate(w http.ResponseWriter, tmpl string, fc *FileContent) {
  err := templates.ExecuteTemplate(w, tmpl+".html", fc)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
}

func debugUnmarshalledJson(m map[string]interface{}) {
  for k, v := range m {
    switch vv := v.(type) {
    case string:
        log.Println(k, "is string", vv)
    case float64:
        log.Println(k, "is float64", vv)
    case bool:
        log.Println(k, "is bool", vv)
    case []interface{}:
        log.Println(k, "is an array:")
        for i, u := range vv {
            log.Println(i, u)
        }
    default:
        log.Println(k, "is of a type I don't know how to handle")
    }
  }
  //log.Printf("Port is %v", m["port"])
}
