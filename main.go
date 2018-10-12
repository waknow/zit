package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func main() {
	const content = "what is up, doc?"

	filepath.Walk(`.git\objects`, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		t, c, err := readObject(path)
		if err != nil {
			log.Println("Read Object Failed", err)
		} else {
			log.Printf("%s => type: %s, content: '%s'", path, t, sPrint(t, c))
		}
		return nil
	})

	// t, c, err := readObject(`.git\objects\b5\5c4b24b098a2b4890f120b8aae44f5f8e59472`)
	// if err != nil {
	// 	log.Fatal(err)
	// } else {
	// 	log.Printf("type: %s, content: %s", t, c)
	// }
}

func sPrint(objType string, content []byte) string {
	switch objType {
	case "blob":
		if len(content) > 10 {
			return string(content)[:10] + "..."
		} else {
			return string(content)
		}
	case "tree":
		s := ""
		index := 0
		for {
			index = bytes.Index(content, []byte("\000"))
			if index == -1 {
				s = string(content)
				break
			}
			nextIndex := bytes.Index(content[index+1:], []byte("\000")) + index
			if nextIndex != -1 {
				s += string(content[:index]) + " " + fmt.Sprintf("%x", content[index+1:nextIndex])
			} else {
				s += string(content[:index]) + " " + fmt.Sprintf("%x", content[index+1:])
				break
			}
			index = nextIndex
			log.Printf("index", index)
		}
		return s
	case "commit":
		return string(content)
	default:
		return ""
	}
}

func readObject(filename string) (string, []byte, error) {
	f, err := os.OpenFile(filename, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return "", nil, err
	}
	defer f.Close()

	reader, err := zlib.NewReader(f)
	if err != nil {
		return "", nil, err
	}

	bs, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", nil, err
	}

	index := bytes.Index(bs, []byte("\000"))
	parts := bytes.Split(bs[:index], []byte(" "))
	t := string(parts[0])
	return t, bs[index+1:], nil
}

func writeObject(content string) {
	header := fmt.Sprintf("blob %d\000", len(content))
	log.Println("header:", header)

	store := header + content
	id := fmt.Sprintf("%x", sha1.Sum([]byte(store)))
	log.Println("id", id)

	buf := new(bytes.Buffer)
	writer := zlib.NewWriter(buf)
	_, err := writer.Write([]byte(store))
	if err != nil {
		log.Fatal(err)
	}
	writer.Flush()
	writer.Close()

	path := ".zit/objects/" + id[0:2] + "/"
	log.Println("filepath", path)

	os.MkdirAll(path, os.ModePerm)
	ioutil.WriteFile(path+id[2:], buf.Bytes(), os.ModePerm)
}
