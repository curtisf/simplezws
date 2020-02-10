package main

import (
	"fmt"
	"log"
	"net/http"
	"io/ioutil"
	"io"
	"os"
	"strings"
	"strconv"
)

var authString = ""
var baseURL = ""

func stringToBin(s string) (binString string) {
    for _, c := range s {
        binString = fmt.Sprintf("%s%.8b",binString, c)
    }
    return 
}

func binaryToZWS(s string) (zwsString string) {
	for _, c := range s {
		if string(c) == "1" {
			zwsString = zwsString + "\u200d" // zero-width joiner space
		} else if (string(c) == "0") {
			zwsString = zwsString + "\u200c" // zero-width non-joiner space
		}
	}
	return
}

func ChunkString(s string, chunkSize int) []string {
    var chunks []string
    runes := []rune(s)

    if len(runes) == 0 {
        return []string{s}
    }

    for i := 0; i < len(runes); i += chunkSize {
        nn := i + chunkSize
        if nn > len(runes) {
            nn = len(runes)
        }
        chunks = append(chunks, string(runes[i:nn]))
    }
    return chunks
}

func zwsToPath(zws string) (path string) {
	var binaryString = ""
	for _, c := range zws {
		if string(c) == "\u200d" { // zero-width joiner space
			binaryString = binaryString + "1" 
		} else if (string(c) == "\u200c") {// zero-width non-joiner space
			binaryString = binaryString + "0" 
		}
	}
	var chunks = ChunkString(binaryString, 8)
	for _, c := range chunks {
		var decConv, err = strconv.ParseInt(c, 2, 64)
		if err != nil {
			panic(err)
		}
		path += string(decConv)
	}
	return
}

func EntryPoint(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		var zwsStr = strings.Replace(r.URL.Path, "/", "", 1)
		var decodedPath = zwsToPath(zwsStr)
		fmt.Printf("Serving: %s\n", decodedPath)
		image, err := os.Open(decodedPath)
		if err != nil {
			panic(err)
		}
		w.Header().Set("Content-Type", "image/png")
		io.Copy(w, image)
	case http.MethodPost:
		w.Header().Set("Content-Type", "application/text")
		if r.Header.Get("Authorization") == "" {
			log.Println("Missing authorization header")
			http.Error(w, "Missing authorization header.", http.StatusUnauthorized)
			return
		}
		if r.Header.Get("Authorization") != authString {
			log.Println("Bad token: " + r.Header.Get("Authorization"))
			http.Error(w, "Invalid token!", http.StatusForbidden)
			return
		}
		r.ParseMultipartForm(10 << 20) // Accept up to 10mb
		file, handler, err := r.FormFile("data")

		if err != nil {
			fmt.Println("There was an error while getting the image from the request")
			fmt.Println(err)
			http.Error(w, "Missing image named data!", http.StatusBadRequest)
			return
		}
		defer file.Close()

		fmt.Printf("Uploaded File: %+v\n", handler.Filename)
    	fmt.Printf("File Size: %+v\n", handler.Size)
		
		tempFile, err := ioutil.TempFile("images", "upload-*.png")
		fmt.Printf("Saving upload as %s\n", tempFile.Name())
		var zwsRep = binaryToZWS(stringToBin(tempFile.Name()))
		if err != nil {
			fmt.Println("Error while making temporary image file")
			fmt.Println(err)
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		
		fileBytes, err := ioutil.ReadAll(file)
    	if err != nil {
			fmt.Println("Error while reading image bytes from upload")
			fmt.Println(err)
			http.Error(w, "Bad image!", http.StatusBadRequest)
			return
		}
		tempFile.Write(fileBytes)
		fmt.Fprintf(w, "%s%s", baseURL, zwsRep) // return just a string for easy xclip copying
	}
}

func main() {
	authData, err := ioutil.ReadFile(".config")
	if err != nil {
		log.Println("Could not open .auth file, exiting!")
		panic(err)
	}
	lines := strings.Split(string(authData), "\n")
	if len(lines) != 2 {
		fmt.Printf("Invalid config - first line must be the authorization string and second must be the base URL. Found %d lines\n", len(lines))
		os.Exit(1)
	}
	authString = strings.TrimSuffix(lines[0], "\n")
	baseURL = strings.TrimSuffix(lines[1], "\n")
	if baseURL == "" || authString == "" {
		log.Println("Invalid config - first line must be the authorization string and second must be the base URL")
		os.Exit(1)
	}
	log.Println("Populated authentication information - content will be served at " + baseURL)
	log.Println("Serving on port 8080")
	http.HandleFunc("/", EntryPoint)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
