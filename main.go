package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/russross/blackfriday/v2"
)

type Posts struct {
	UserID int    `json:"userId"`
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

type AllPosts []Posts

type Meta struct {
	Timestamp    time.Time `json:"timestamp"`
	TotalResults int       `json:"totalResults"`
	Start        int       `json:"start"`
	Offset       int       `json:"offset"`
	Limit        int       `json:"limit"`
}

type WebData struct {
	Version                 int    `json:"version"`
	VersionZUID             string `json:"versionZUID"`
	MetaDescription         string `json:"metaDescription"`
	MetaTitle               string `json:"metaTitle"`
	MetaLinkText            string `json:"metaLinkText"`
	MetaKeywords            string `json:"metaKeywords"`
	ParentZUID              string `json:"parentZUID"`
	PathPart                string `json:"pathPart"`
	Path                    string `json:"path"`
	SitemapPriority         int    `json:"sitemapPriority"`
	CanonicalTagMode        int    `json:"canonicalTagMode"`
	CanonicalQueryParamList string `json:"canonicalQueryParamList"`
	CanonicalTagCustomValue string `json:"canonicalTagCustomValue"`
	CreatedByUserZUID       string `json:"createdByUserZUID"`
	CreatedAt               string `json:"createdAt"`
	UpdatedAt               string `json:"updatedAt"`
}

type MetaData struct {
	ZUID             string `json:"ZUID"`
	Zid              int    `json:"zid"`
	MasterZUID       string `json:"masterZUID"`
	ContentModelZUID string `json:"contentModelZUID"`
	ContentModelName string `json:"contentModelName"`
	Sort             int    `json:"sort"`
	Listed           bool   `json:"listed"`
	Version          int    `json:"version"`
	LangID           int    `json:"langID"`
	CreatedAt        string `json:"createdAt"`
	UpdatedAt        string `json:"updatedAt"`
}

type Siblings struct {
	EnUS string `json:"en-US"`
}

type Data struct {
	Web      WebData  `json:"web"`
	Meta     MetaData `json:"meta"`
	Siblings Siblings `json:"siblings"`
	Data     struct {
		Body        string `json:"body"`
		Description string `json:"description"`
		Title       string `json:"title"`
	} `json:"data"`
}

type BodyData struct {
	Body        string `json:"body"`
	Description string `json:"description"`
	Title       string `json:"title"`
}

type Response struct {
	Meta struct {
		Timestamp    time.Time `json:"timestamp"`
		TotalResults int       `json:"totalResults"`
		Start        int       `json:"start"`
		Offset       int       `json:"offset"`
		Limit        int       `json:"limit"`
	} `json:"_meta"`
	Data []Data `json:"data"`
}

func handleData(c *fiber.Ctx) error {
	token := c.Query("token")
	if token == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Token is missing")
	}

	url := "https://8-aaeffee09b-7w6v22.api.zesty.io/v1/content/models/6-e092db91a5-rgfcx2/items?limit=5000&page=1&lang=en-US"

	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.Status(resp.StatusCode).SendString("Failed to fetch data")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	var responseStruct Response
	if err := json.Unmarshal([]byte(body), &responseStruct); err != nil {
		fmt.Println("Error decoding JSON:", err)
	}

	var formattedData []map[string]interface{}
	for _, item := range responseStruct.Data {
		mdFIle := item.Data.Body
		htmlBody := string(blackfriday.Run([]byte(mdFIle)))
		content := strings.ReplaceAll(htmlBody, "\n", "<br/>")

		formattedItem := map[string]interface{}{
			"item": map[string]interface{}{
				"zuid":     item.Meta.ZUID,
				"content":  content,
				"markdown": mdFIle,
			},
		}
		formattedData = append(formattedData, formattedItem)
	}

	// Create a map for the final JSON structure
	result := map[string]interface{}{
		"data": formattedData,
	}

	// Marshal the result into JSON
	resultJSON, err := json.Marshal(result)
	if err != nil {
		fmt.Println("Error marshalling result to JSON:", err)
	}

	// Open the file for writing (overwriting if it exists)
	file, err := os.Create("result.json")
	if err != nil {
		fmt.Println("Error creating file:", err)
	}
	defer file.Close()

	// Write the JSON content to the file
	_, err = file.Write(resultJSON)
	if err != nil {
		fmt.Println("Error writing JSON to file:", err)
	}

	fmt.Println("result.json file overwritten successfully.")
	return c.Status(fiber.StatusOK).Send(resultJSON)
}

func openBrowser(url string) error {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	if err != nil {
		fmt.Println("Error opening browser:", err)
		return err
	}

	return nil
}
func main() {
	app := fiber.New()

	app.Get("/", handleData)

	openBrowser("http://localhost:3000/?token=")
	app.Listen(":3000")

}
