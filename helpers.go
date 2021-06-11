package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
)

func InterfaceMap(i interface{}) (interface{}, error) {
	t := reflect.TypeOf(i)
	switch t.Kind() {
	case reflect.Map:
		v := reflect.ValueOf(i)
		it := reflect.TypeOf((*interface{})(nil)).Elem()
		m := reflect.MakeMap(reflect.MapOf(t.Key(), it))
		for _, mk := range v.MapKeys() {
			m.SetMapIndex(mk, v.MapIndex(mk))
		}
		return m.Interface(), nil
	}
	return nil, errors.New("Unsupported type")
}

func String(v string) *string {
	return &v
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE")
}

func unmarshallResource(resource *errorPageData) (output []byte) {
	output, _ = json.Marshal(resource)
	return output
}

func fileExists(filename string) (fileExist bool) {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func dirExists(filename string) (dirExist bool) {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func templateFileCheck(template string, code string, filetype string) (filepath string, err error) {
	var filename, extension string
	switch filetype {
	case "css":
		filename = "style"
		extension = "css"
		break
	case "html":
		filename = "index"
		extension = "html"
		break
	default:
		return "", err
	}
	codeBasedFile := fmt.Sprintf("%v/%v/%v-%v.%v", templatePath, template, filename, code, extension)
	genericFile := fmt.Sprintf("/%v/%v/%v.%v", templatePath, template, filename, extension)
	fallbackFile := fmt.Sprintf("/%v/default/%v.%v", templatePath, filename, extension)
	switch {
	case fileExists(codeBasedFile):
		return codeBasedFile, err
	case fileExists(genericFile):
		return genericFile, err
	case fileExists(fallbackFile):
		return fallbackFile, err
	}
	return filepath, err
}
