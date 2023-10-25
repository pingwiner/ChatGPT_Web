package main

import (
	"context"
	"errors"	
	"fmt"
	"io"
	"os"		
	"net/http"
	"encoding/json"		
	openai "github.com/sashabaranov/go-openai"
	"github.com/sevlyar/go-daemon"
	"log"
)

type RequestBody struct {
	message string `json:"message"` 
}

type MsgResponse struct {
    Success bool `json:"success"`
    GptResponse string `json:"response"`
}

var client *openai.Client
var req openai.ChatCompletionRequest


func Init() {
	client = openai.NewClient(os.Getenv("OPENAI_API_KEY"))	
	req = openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "you are a helpful chatbot",
			},
		},
	}	
}



func WriteResponse(writer http.ResponseWriter, response MsgResponse) {
	bytes, err := json.Marshal(response)

   	if err != nil {
    	panic(err)
   	}

	if (response.Success) {
		writer.WriteHeader(http.StatusOK)
	} else {
		writer.WriteHeader(http.StatusInternalServerError)	
  	}

  	writer.Header().Set("Access-Control-Allow-Origin", "*")
	writer.Header().Set("Access-Control-Allow-Methods", "*")
	writer.Header().Set("Content-Type", "application/json")

  	io.WriteString(writer, string(bytes))
}

func gptRequest(writer http.ResponseWriter, message string) {	
	fmt.Printf("request: %v\n", message)
	req.Messages = append(req.Messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: message,
	})

	resp, err := client.CreateChatCompletion(context.Background(), req)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		response := MsgResponse { false, "Internal error" }
		WriteResponse(writer, response)
		return
	}

	response := MsgResponse { true, resp.Choices[0].Message.Content }
	WriteResponse(writer, response)
	fmt.Println(resp.Choices[0].Message.Content)
	req.Messages = append(req.Messages, resp.Choices[0].Message)
}

func getRoot(writer http.ResponseWriter, r *http.Request) {
	io.WriteString(writer, "=^_^=")
}

func chat(writer http.ResponseWriter, r *http.Request) {
	switch r.Method {
		case "POST":
			if err := r.ParseForm(); err != nil {
				fmt.Printf("ParseForm() err: %v", err)
				return
			}
			msgText := r.FormValue("message")
			gptRequest(writer, msgText)
			fmt.Printf("body: %v\n", msgText)

		default:
			fmt.Printf("Sorry, only POST method is supported.")
	}
}


func main() {
	cntxt := &daemon.Context{
		PidFileName: "chatgpt_web.pid",
		PidFilePerm: 0644,
		LogFileName: "chatgpt_web.log",
		LogFilePerm: 0640,
		WorkDir:     "./",
		Umask:       027,
		Args:        []string{"[go-daemon chatgpt_web]"},
	}


	d, err := cntxt.Reborn()
	if err != nil {
		log.Fatal("Unable to run: ", err)
	}
	if d != nil {
		return
	}
	defer cntxt.Release()

	log.Print("- - - - - - - - - - - - - - -")
	log.Print("daemon started")

	Init()
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.HandleFunc("/chat", chat)

	err = http.ListenAndServe(":8081", nil)

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
