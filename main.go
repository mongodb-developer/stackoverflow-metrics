package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type InputData struct {
	Output      string   `json:"output"`
	From        string   `json:"from,omitempty"`
	To          string   `json:"to,omitempty"`
	QuestionIds []string `json:"questions"`
}

type QuestionsResponse struct {
	Items []struct {
		Tags  []string `json:"tags"`
		Owner struct {
			Reputation  int    `json:"reputation"`
			UserId      int64  `json:"user_id"`
			UserType    string `json:"user_type"`
			DisplayName string `json:"display_name"`
			Link        string `json:"link"`
		} `json:"owner"`
		IsAnswered       bool   `json:"is_answered"`
		ViewCount        int    `json:"view_count"`
		AnswerCount      int    `json:"answer_count"`
		Score            int    `json:"score"`
		LastActivityDate int64  `json:"last_activity_date"`
		CreationDate     int64  `json:"creation_date"`
		LastEditDate     int64  `json:"last_edit_date"`
		QuestionId       int64  `json:"question_id"`
		Link             string `json:"link"`
		Title            string `json:"title"`
	} `json:"items"`
	HasMore        bool `json:"has_more"`
	QuotaMax       int  `json:"quota_max"`
	QuotaRemaining int  `json:"quota_remaining"`
}

type ErrorResponse struct {
	ErrorId      int    `json:"error_id"`
	ErrorMessage string `json:"error_message"`
	ErrorName    string `json:"error_name"`
}

func LoadConfiguration(file string) InputData {
	var data InputData
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&data)
	return data
}

func GetQuestions(ids []string, from int64, to int64) (QuestionsResponse, error) {
	endpoint, _ := url.Parse("https://api.stackexchange.com/2.2/questions/" + strings.Join(ids, ";"))
	queryParams := endpoint.Query()
	queryParams.Set("site", "stackoverflow")
	if from > 0 {
		queryParams.Set("fromdate", strconv.FormatInt(from, 10))
	}
	if to > 0 {
		queryParams.Set("todate", strconv.FormatInt(to, 10))
	}
	endpoint.RawQuery = queryParams.Encode()
	response, err := http.Get(endpoint.String())
	if err != nil {
		return QuestionsResponse{}, err
	}
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return QuestionsResponse{}, err
	}
	var errorResponse ErrorResponse
	json.Unmarshal(data, &errorResponse)
	if errorResponse != (ErrorResponse{}) {
		return QuestionsResponse{}, errors.New(errorResponse.ErrorName + ": " + errorResponse.ErrorMessage)
	}
	var questions QuestionsResponse
	json.Unmarshal(data, &questions)
	return questions, nil
}

func main() {
	input := flag.String("input", "input.json", "JSON file path for input data")
	flag.Parse()
	fmt.Println("Starting the application...")
	inputData := LoadConfiguration(*input)
	fromDate, _ := time.Parse("2006-01-02", inputData.From)
	toDate, _ := time.Parse("2006-01-02", inputData.To)
	csvFile, _ := os.Create(inputData.Output)
	questions, err := GetQuestions(inputData.QuestionIds, fromDate.Unix(), toDate.Unix())
	if err != nil {
		panic(err)
	}
	fmt.Println("Questions found: " + strconv.Itoa(len(questions.Items)))
	writer := csv.NewWriter(csvFile)
	writer.Write([]string{"Title", "Link", "Tags", "Answered", "Answer Count", "View Count", "Creation Date"})
	for _, question := range questions.Items {
		data := []string{question.Title, question.Link, strings.Join(question.Tags, "/"), strconv.FormatBool(question.IsAnswered), strconv.Itoa(question.AnswerCount), strconv.Itoa(question.ViewCount), time.Unix(question.CreationDate, 0).Format(time.RFC3339)}
		writer.Write(data)
	}
	writer.Flush()
}
