package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

var crypto = []byte{}

var (
	newFile *os.File
	err     error
)

func addDocUnsafe(data map[string]string, name string, dir string) error {
	toJSON := "{"
	for k, v := range data {
		toJSON = toJSON + "\"" + k + "\":\"" + v + "\","
	}
	toJSON = toJSON[:len(toJSON)-1] + "}"
	//create document directory if it doesn't exist
	var document *os.File
	var err error
	if dir == "" {
		_ = os.Mkdir("documents", os.ModePerm)
		document, err = os.Create("documents/" + name + ".json")

	} else {
		_ = os.Mkdir("documents", os.ModePerm)
		_ = os.Mkdir("documents/"+dir, os.ModePerm)
		document, err = os.Create("documents/" + dir + "/" + name + ".json")

	}
	if err != nil {
		return err
	}
	defer document.Close()

	_, err = document.WriteString(toJSON)

	if err != nil {
		return err
	}
	return nil
}
func readDocUnsafe(name string, dir string) (map[string]string, error) {
	if !fileExists("documents/" + dir + "/" + name + ".json") {
		return nil, os.ErrNotExist
	}
	document, err := os.Open("documents/" + name + ".json")
	if err != nil {
		return nil, err
	}
	defer document.Close()
	var data map[string]string
	decoder := json.NewDecoder(document)
	err = decoder.Decode(&data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
func removeDoc(name string, dir string) {
	if fileExists("documents/" + dir + "/" + name + ".json") {
		os.Remove("documents/" + dir + "/" + name + ".json")
	}
}
func readKey(key string, dir string) string {
	if hasKey(key, dir) {
		file, err := os.Open(dir + "/" + key)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		data, err := ioutil.ReadAll(file)
		if err != nil {
			log.Fatal(err)
		}
		ret := (string(data))
		result, err := DecryptGCM(crypto, []byte(ret))
		if err != nil {
			fmt.Println(err)
		}
		return string(result)
	}
	return ""
}

func addKey(key string, value string, dir string) {
	ciphertext, err := EncryptGCM(crypto, []byte(value))
	if err != nil {
		fmt.Println(err)
	}

	if !hasKey(key, dir) {
		os.Mkdir(dir, os.ModePerm)
		newFile, err = os.Create(dir + "/" + key)
		if err != nil {
			log.Fatal(err)
		}
		defer newFile.Close()
		log.Println(newFile)
		err2 := ioutil.WriteFile(dir+"/"+key, []byte(ciphertext), 0666)
		if err2 != nil {
			log.Fatal(err2)
		}
	}

}
func removeKey(key string, dir string) {
	if hasKey(key, dir) {
		err := os.Remove(dir + "/" + key)
		if err != nil {
			log.Fatal(err)
		}
	}

}

func changeKeyUnsafe(key string, value string, dir string) {
	removeKey(key, dir)
	addKeyUnsafe(key, value, dir)
}

func changeKey(key string, value string, dir string) {
	removeKey(key, dir)
	addKey(key, value, dir)
}

func hasKey(key string, dir string) bool {
	_, err := os.Stat(dir + "/" + key)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
func hasDoc(name string, dir string) bool {
	_, err := os.Stat("documents/" + dir + "/" + name + ".json")
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
func addKeyUnsafe(key string, value string, dir string) {
	if err != nil {
		fmt.Println(err)
	}

	if !hasKey(key, dir) {
		os.Mkdir(dir, os.ModePerm)
		newFile, err = os.Create(dir + "/" + key)
		if err != nil {
			log.Fatal(err)
		}
		defer newFile.Close()
		log.Println(newFile)
		err2 := ioutil.WriteFile(dir+"/"+key, []byte(value), 0666)
		if err2 != nil {
			log.Fatal(err2)
		}
	}

}
func readKeyUnsafe(key string, dir string) string {
	if hasKey(key, dir) {
		file, err := os.Open(dir + "/" + key)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		data, err := ioutil.ReadAll(file)
		if err != nil {
			log.Fatal(err)
		}
		ret := (string(data))
		if err != nil {
			fmt.Println(err)
		}
		return ret
	}
	return ""
}
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
