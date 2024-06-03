package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gocolly/colly/v2"
)

const (
	baseURL   = "http://localhost/song"
	localDir  = "downloaded"
	try       = 12
	longWait  = 1 * time.Minute
	shortWait = 5 * time.Second
	timeout   = 10 * time.Minute
)

func downloadFile(filename, fileURL string) error {

	startTime := time.Now()

	attempts := 0
	for {
		resp, err := http.Get(fileURL)
		if err == nil && resp.StatusCode == http.StatusOK {
			file, err := os.Create(filepath.Join(localDir, filename)) //создание файла
			if err != nil {
				resp.Body.Close()
				return err
			}
			_, err = io.Copy(file, resp.Body) //ctrl+V ответа в файл
			file.Close()
			resp.Body.Close()
			if err == nil {
				return nil
			}
		}
		if resp != nil {
			resp.Body.Close() //закрытие
		}

		attempts++
		if time.Since(startTime) > timeout {
			return fmt.Errorf("не удалось загрузить файл %s после истечения времени ожидания", filename)
		}
		if attempts == try {
			fmt.Printf("Ожидание долгой паузы перед следующей попыткой...\n")
			time.Sleep(longWait)
		} else {
			fmt.Printf("Повторная попытка через 5 секунд...\n")
			time.Sleep(shortWait)
		}
	}
}

func scrapeLinks() []string {
	c := colly.NewCollector(colly.AllowedDomains("localhost")) //коллектор локалхоста

	var links []string //список ссылок
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		if len(filepath.Ext(link)) > 1 { //файл имеет расширение?
			links = append(links, link)
		}
	})
	c.Visit(baseURL)
	return links
}

func logSuccess(filename string) {
	file, err := os.OpenFile("successful.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Ошибка при открытии successful.txt", err)
		return
	}
	defer file.Close()

	if _, err := file.WriteString(filename + "\n"); err != nil {
		fmt.Println("Ошибка при записи в successful.txt", err)
	}
}

func logFailure(filename string) {
	file, err := os.OpenFile("not-successful.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Ошибка при открытии not-successful.txt", err)
		return
	}
	defer file.Close()

	if _, err := file.WriteString(filename + "\n"); err != nil {
		fmt.Println("Ошибка при записи в not-successful.txt", err)
	}
}

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Ошибка при получении текущего рабочего каталога:", err)
		return
	}
	fmt.Println("Текущий рабочий каталог:", cwd)

	if err := os.MkdirAll(localDir, 0755); err != nil {
		fmt.Printf("Ошибка при создании каталога: %v\n", err)
		return
	}

	fileLinks := scrapeLinks()

	for _, fileURL := range fileLinks {
		filename := filepath.Base(fileURL)
		fmt.Printf("Загружаем %s...\n", filename)
		if err := downloadFile(filename, fileURL); err != nil {
			fmt.Printf("Ошибка загрузки %s: %v\n", filename, err)
			fmt.Println("Запись ошибки загрузки в лог", filename)
			logFailure(filename)
			continue
		}
		fmt.Printf("Загрузка %s\n", filename)
		fmt.Println("Запись успешной загрузки в лог", filename)
		logSuccess(filename)
	}
}
