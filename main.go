package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

const openaiAPIKey = "your-openai-api-key"

// Structs for the API request and response
type StoryRequest struct {
	Story string `json:"story"`
}

type OpenAIResponse struct {
	Choices []struct {
		Text string `json:"text"`
	} `json:"choices"`
}

type ImageResponse struct {
	Data []struct {
		URL string `json:"url"`
	} `json:"data"`
}

// Function to split the story into parts
func splitStoryIntoParts(story string) ([]string, error) {
	url := "https://api.openai.com/v1/completions"

	requestBody, err := json.Marshal(map[string]interface{}{
		"model":      "text-davinci-003",
		"prompt":     fmt.Sprintf("Split this story into distinct parts suitable for creating a comic book: %s", story),
		"max_tokens": 300,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+openaiAPIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var openAIResponse OpenAIResponse
	err = json.Unmarshal(body, &openAIResponse)
	if err != nil {
		return nil, err
	}

	if len(openAIResponse.Choices) > 0 {
		// Split the text into individual parts by lines or paragraphs
		parts := openAIResponse.Choices[0].Text
		return splitByDelimiter(parts, "\n"), nil
	}

	return nil, fmt.Errorf("no response from GPT-3")
}

// Helper function to split text by delimiter (like newline or sentence separator)
func splitByDelimiter(text, delimiter string) []string {
	var parts []string
	for _, part := range bytes.Split([]byte(text), []byte(delimiter)) {
		trimmed := string(bytes.TrimSpace(part))
		if len(trimmed) > 0 {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

// Function to generate an image from a description
func generateImage(description string) (string, error) {
	url := "https://api.openai.com/v1/images/generations"

	requestBody, err := json.Marshal(map[string]interface{}{
		"prompt": description,
		"n":      1,
		"size":   "1024x1024",
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+openaiAPIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var imageResponse ImageResponse
	err = json.Unmarshal(body, &imageResponse)
	if err != nil {
		return "", err
	}

	if len(imageResponse.Data) > 0 {
		return imageResponse.Data[0].URL, nil
	}

	return "", fmt.Errorf("no image generated")
}

// Main handler for generating multiple images based on story parts
func handleGenerateComic(c *gin.Context) {
	var storyRequest StoryRequest

	// Parse incoming request
	if err := c.BindJSON(&storyRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Step 1: Split story into parts
	parts, err := splitStoryIntoParts(storyRequest.Story)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Step 2: Generate images for each part
	var images []string
	for _, part := range parts {
		imageURL, err := generateImage(part)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		images = append(images, imageURL)
	}

	// Return the generated images as a response
	c.JSON(http.StatusOK, gin.H{"images": images})
}

func main() {
	// Load OpenAI API key from environment variables
	if openaiAPIKey == "" {
		log.Fatal("OpenAI API key not found in environment variables")
	}

	// Set up the Gin router
	r := gin.Default()

	// Route to generate comic book images
	r.POST("/generate-comic", handleGenerateComic)

	// Start the server
	r.Run(":8080")
}
