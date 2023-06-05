package main

import (
  "io"
	"fmt"
  "time"
  "sync"
  "strconv"
  "strings"
	"net/http"
  "encoding/csv"
  "github.com/gin-gonic/gin"
)

type Promotion struct {
	Id string `json:"id"`
  Price float32 `json:"price"`
  ExpirationDate time.Time `json:"expiration_date"`
}

var promotionList = make([]Promotion, 0, 0)
var mutex = sync.Mutex{}

func parseExpirationDate(dateString string) (time.Time, error) {
  str := dateString
  str = strings.Replace(str, " ", "T", 1)
  str = strings.Replace(str, " +", "+", 1)
  parts := strings.Split(str, " ")
  str = parts[0]
  str = str[:len(str)-2] + ":" + str[len(str)-2:]

	date, err := time.Parse("2006-01-02T15:04:05-07:00", str)

  return date, err
}

func createPromotion(serializedPromotion []string) (Promotion, error) {
  id := serializedPromotion[0]
  priceFloat, err := strconv.ParseFloat(serializedPromotion[1], 32)

  if err != nil {
		fmt.Println("Failed to convert string to float32:", err)
		return Promotion{}, err
	}

	price := float32(priceFloat)
	expirationDate, err := parseExpirationDate(serializedPromotion[2])

  if err != nil {
		fmt.Println("Failed to convert string to time:", err)
		return Promotion{}, err
	}

  return Promotion{id, price, expirationDate}, nil
}

func readCSV(file io.Reader, rawData chan<- []string, c *gin.Context) {
  reader := csv.NewReader(file)

  for {
    record, err := reader.Read()

    if err == io.EOF {
      break
    }

    if err != nil {
      fmt.Println("Failed to read the attached file:", err)

      promotionList = make([]Promotion, 0, 0)

      c.JSON(http.StatusBadRequest, gin.H{"error": "invalid CSV file"})
      return
    }

    rawData <- record
  }

  close(rawData)
}

func processRows(rawData <-chan []string, promotionsChan chan<- Promotion) {
  for row := range rawData {
    promotion, err := createPromotion(row)

    if err != nil {
      continue
    }

    promotionsChan <- promotion
  }

  close(promotionsChan)
}

func storePromotions(promotionsChan <-chan Promotion) {
  for promotion := range promotionsChan {
    promotionList = append(promotionList, promotion)
  }
}

func handleFileUpload(c *gin.Context) {
  file, _, err := c.Request.FormFile("file")
  invalidFileError := gin.H{"error": "invalid CSV file"}

  if err != nil {
    c.JSON(http.StatusBadRequest, invalidFileError)
    return
  }

  // as store is immutable that means only one upload thread should run at a time
  mutex.Lock()

  // empty the list
  promotionList = make([]Promotion, 0, 0)

  rawData := make(chan []string)
  promotionsChan := make(chan Promotion)

  go readCSV(file, rawData, c)
  go processRows(rawData, promotionsChan)
  go storePromotions(promotionsChan)

  mutex.Unlock()
}

func getPromotionById(c *gin.Context) {
  indexStr := c.Param("id")
	index, err := strconv.Atoi(indexStr)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

  if index <= 0 || index > len(promotionList) {
		c.JSON(http.StatusNotFound, gin.H{"error": "promotion not found"})
		return
	}

	promotion := promotionList[index - 1]

	c.JSON(http.StatusOK, promotion)
}

func main() {
  router := gin.Default()
  router.POST("/upload", handleFileUpload)
  router.GET("/promotions/:id", getPromotionById)

	router.Run(":8080")
}
