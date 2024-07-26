package http

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"tgBot/bot"
	"time"
)

func getToDoList(w http.ResponseWriter, req *http.Request) {

	chatId := req.URL.Query().Get("chatId")
	id, err := strconv.ParseInt(chatId, 10, 64)
	if err != nil {
		fmt.Fprintf(w, "incorrect chatId\n")
		return
	}

	list, err := bot.GetToDoList(id)
	if err != nil {
		fmt.Fprintf(w, "cant get ToDoList\n")
		return
	}

	fmt.Fprintf(w, list)
}

func removeAll(w http.ResponseWriter, req *http.Request) {

	chatId := req.URL.Query().Get("chatId")
	id, err := strconv.ParseInt(chatId, 10, 64)
	if err != nil {
		fmt.Fprintf(w, "incorrect chatId\n")
		return
	}

	err = bot.RemoveToDoList(id)
	if err != nil {
		fmt.Fprintf(w, "cant remove ToDoList\n")
		return
	}

	fmt.Fprintf(w, "list was cleared\n")
}

func hi(w http.ResponseWriter, req *http.Request) {

	fmt.Fprintf(w, "hi\n")
}

func ServerStart() (*http.Server, error) {
	server := &http.Server{Addr: ":8000"}

	http.HandleFunc("/getToDoList", getToDoList)
	http.HandleFunc("/removeAll", removeAll)
	http.HandleFunc("/", hi)

	go func() error {
		if err := server.ListenAndServe(); err != nil {
			return fmt.Errorf("failed to start server %w", err)
		}
		return nil
	}()
	return server, nil
}

func ServerStop(server *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	server.Shutdown(ctx)
}
