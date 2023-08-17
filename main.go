package main

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/PuerkitoBio/goquery"
)

// Product структура для хранения информации о товаре
type Product struct {
	Name        string
	Description string
	PhotoLinks  []string
}

// fetchPageData функция для получения данных страницы
func fetchPageData(url string) (*goquery.Document, error) {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func main() {
	var baseURL string
	fmt.Print("Введите базовый URL: ")
	fmt.Scanln(&baseURL) // Пользователь вводит базовый URL

	pageNumber := 1
	var products []Product

	for {
		url := fmt.Sprintf("%s?_pgn=%d", baseURL, pageNumber)
		doc, err := fetchPageData(url)

		if err != nil {
			log.Println("Error fetching data:", err)
			break
		}

		fmt.Println("Страница:", pageNumber)

		doc.Find("article").Each(func(_ int, item *goquery.Selection) {
			dataTestIDValue, _ := item.Attr("data-testid")
			modifiedDataTestIDValue := strings.Replace(dataTestIDValue, "ig-", "", -1)
			itemURL := fmt.Sprintf("https://www.ebay.com/itm/%s", modifiedDataTestIDValue)

			itemDoc, err := fetchPageData(itemURL)
			if err != nil {
				log.Println("Error fetching item data:", err)
				return
			}

			product := Product{}

			product.Name = itemDoc.Find(".x-item-title__mainTitle").Text()
			product.Description = itemDoc.Find(".x-price-primary").Text()

			itemDoc.Find("button img").Each(func(_ int, imgTag *goquery.Selection) {
				srcValue, exists := imgTag.Attr("src")
				if exists {
					newSrcValue := strings.Replace(srcValue, "s-l64.jpg", "s-l1600.jpg", -1)
					newSrcValue = strings.Replace(newSrcValue, "l140.jpg", "s-l1600.jpg", -1)
					product.PhotoLinks = append(product.PhotoLinks, newSrcValue)
				}
			})

			products = append(products, product)
		})

		nextPageLink := doc.Find("a.pagination__next")
		if nextPageLink.Length() == 0 {
			break
		}

		pageNumber++
		delay := rand.Intn(6) + 10
		fmt.Printf("Пауза %d секунд перед следующей страницей...\n", delay)
		time.Sleep(time.Duration(delay) * time.Second)
	}

	// Создаем новый XLSX файл
	file := excelize.NewFile()
	sheetName := "Products"
	index := file.NewSheet(sheetName)

	file.SetCellValue(sheetName, "A1", "Name")
	file.SetCellValue(sheetName, "B1", "Description")
	file.SetCellValue(sheetName, "C1", "PhotoLinks")

	for i, product := range products {
		row := i + 2
		photoLinks := strings.Join(product.PhotoLinks, "\n")

		file.SetCellValue(sheetName, fmt.Sprintf("A%d", row), product.Name)
		file.SetCellValue(sheetName, fmt.Sprintf("B%d", row), product.Description)
		file.SetCellValue(sheetName, fmt.Sprintf("C%d", row), photoLinks)
	}

	file.SetActiveSheet(index)
	file.SaveAs("products.xlsx")

	fmt.Println("Парсинг и экспорт в Excel завершены.")
	fmt.Printf("Получено элементов: %d\n", len(products))

	fmt.Println("Нажмите Enter для завершения...")
	fmt.Scanln()
}
