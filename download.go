package main

import (
    "fmt"
    "flag"
    "io/ioutil"
    "net/http"
    "os"
    "io"
    "encoding/xml"
    "compress/gzip"
)

type Query struct {
    RepomdList []DataElement `xml:"data"`
    Revision string `xml:"revision"`
}

type DataElement struct {
    Type string `xml:"type,attr"`
    Location Href `xml:"location"`
    Size int `xml:"size"`
}

type Href struct {
    Href string `xml:"href,attr"̀`
}

type Version struct {
    Epoch string `xml:"epoch,attr"̀`
    Ver string `xml:"ver,attr"̀`
    Rel string `xml:"rel,attr"̀`
}

type Primary struct{
    RpmList []Rpm `xml:"package"`
}

type Rpm struct {
    Type string `xml:"type,attr"`
    Location Href `xml:"location"`
    Name string `xml:"name"`
    Version Version `xml:"version"`
}


//func (s DataElement) String() string {
//    return fmt.Sprintf("%s %s (%d)",s.Type, s.Location.Href, s.Size)
//}

//func (s Query) String() string {
//    return fmt.Sprintf("revision=%s ", s.Revision)
//}

func downloadURL(url string) http.Response {
    resp,  err := http.Get(url)
    if(err != nil){
        fmt.Fprintf(os.Stderr, "fetch %s failed: %v\n",url,err)
        os.Exit(1)
    }
    if((resp.StatusCode < 200) || (resp.StatusCode  > 299)){
        fmt.Fprintf(os.Stderr, "received %d exit code downloading %s\n",resp.StatusCode ,url)
        os.Exit(1)
    }
    return *resp
}


func getGzipBytes(r io.Reader) []byte {
    var unzipped io.Reader
    //if(url[len(url)-3:]==".gz"){
    unzipped,err := gzip.NewReader(r)
    if(err != nil){
        fmt.Fprintf(os.Stderr, "invalid gzip: %v\n",err)
        os.Exit(1)
    }
    return getBytes(unzipped)
}


func getBytes(r io.Reader) []byte {
    //content, err := ioutil.ReadAll(resp.Body)
    content, err := ioutil.ReadAll(r)
    if err != nil {
        fmt.Fprintf(os.Stderr, "reading failed: %v\n",err)
        os.Exit(1)
    }

    return content
}


func downloadGzip(url string)  []byte {
    resp := downloadURL(url)
    return getGzipBytes(resp.Body)
}


func download(url string)  []byte {
    resp := downloadURL(url)
    return getBytes(resp.Body)
}


func downloadFile(filepath string, url string) (err error) {

  // Create the file
  out, err := os.Create(filepath)
  if err != nil  {
    return err
  }
  defer out.Close()

  resp := downloadURL(url)
  defer resp.Body.Close()

  // Writer the body to file
  _, err = io.Copy(out, resp.Body)
  if err != nil  {
    return err
  }

  return nil
}


func exists(path string) bool {
    if _, err := os.Stat(path); os.IsNotExist(err) {
        return false
    }
    return true
}



func main (){
    destPtr := flag.String("dest", "", "Destination to write to")
    //srcPtr := flag.String("src", "", "Repo to read from")

    flag.Parse()

    if(*destPtr == ""){
        fmt.Fprintf(os.Stderr,"--dest argument is required\n")
        os.Exit(1)
    }


    for _,url := range flag.Args() {
        var repomd string = url + "/repodata/repomd.xml"
        var content []byte = download(repomd)

        var q Query
        xml.Unmarshal(content, &q)
        for _, d := range q.RepomdList {
            if d.Type == "primary" {
                var primarydata []byte =  downloadGzip(url + d.Location.Href)
                var primary Primary
                xml.Unmarshal(primarydata,&primary)
                for _,r := range primary.RpmList {
                    fmt.Printf("%s %s\t%s\n",r.Type,r.Name,r.Location.Href)
                    if(exists(*destPtr + r.Location.Href) == false){
                        fmt.Printf("%s from %s\n",*destPtr + r.Location.Href, url + r.Location.Href)
                        downloadFile(*destPtr + r.Location.Href, url + r.Location.Href)
                        os.Exit(0)
                    }
                }

            }
        }
    }
}

