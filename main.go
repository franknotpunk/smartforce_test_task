package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var logMutex sync.Mutex

func writeLog(message string) {
	logMutex.Lock()
	defer logMutex.Unlock()
	log.Println(message)
}

func main() {
	logFile, err := os.Create("log_" + time.Now().Format("2007-01-01_04-20-00") + ".txt")
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(logFile)

	http.HandleFunc("/replace", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Invalid request method", http.StatusBadRequest)
			return
		}

		dirPath := r.URL.Query().Get("dir")
		oldText := r.URL.Query().Get("old")
		newText := r.URL.Query().Get("new")

		fmt.Println("Выполнен запрос на изменение содержимого директории: ", dirPath,
			" с текстом: ", oldText, " на текст: ", newText, " в ", time.Now().Format("2007-01-01_04-20-00"))

		if dirPath == "" || oldText == "" || newText == "" {
			http.Error(w, "Invalid request parameters", http.StatusBadRequest)
			return
		}

		var wg sync.WaitGroup

		err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				writeLog(err.Error())
				return err
			}
			if !info.IsDir() {
				wg.Add(1)
				go func(path string) {
					defer wg.Done()
					fileContent, err := os.ReadFile(path)
					if err != nil {
						writeLog(err.Error())
						return
					}

					newFileContent := string(fileContent)
					//context := ""

					for {
						index := strings.Index(newFileContent, oldText)
						if index == -1 {
							break
						}

						start := index - 20
						if start < 0 {
							start = 0
						}
						end := index + len(oldText) + 20
						if end > len(newFileContent) {
							end = len(newFileContent)
						}
						newContext := newFileContent[start:end] + "...\n"

						writeLog(fmt.Sprintf("Файл: %s\nЗамена: '%s' -> '%s'\nКонтекст замены: %s",
							info.Name(), oldText, newText, newContext))

						newFileContent = newFileContent[:index] + newText + newFileContent[index+len(oldText):]
					}

					err = os.WriteFile(path, []byte(newFileContent), 0644)
					if err != nil {
						writeLog(err.Error())
						return
					}
				}(path)
			}
			return nil
		})
		if err != nil {
			writeLog(err.Error())
			return
		}

		wg.Wait()

		w.WriteHeader(http.StatusOK)
	})

	fmt.Println("Сервис запущен на порту 8080")
	fmt.Println("GET запрос http://localhost:8080/replace?dir=/path/to/directory&old=text_to_replace&new=new_text")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
