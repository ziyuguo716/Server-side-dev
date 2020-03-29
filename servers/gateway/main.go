package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/UW-Info-441-Winter-Quarter-2020/homework-ziyuguo716/servers/gateway/handlers"
	"github.com/UW-Info-441-Winter-Quarter-2020/homework-ziyuguo716/servers/gateway/models/users"
	"github.com/UW-Info-441-Winter-Quarter-2020/homework-ziyuguo716/servers/gateway/sessions"
	"github.com/go-redis/redis"
	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"
)

//Director is a middleware
type Director func(r *http.Request)

//main is the main entry point for the server
func main() {
	// 1.Read the ADDR environment variable to get the address
	// the server should listen on. If empty, default to ":80"
	addr := os.Getenv("ADDR")
	if len(addr) == 0 {
		//443 is the defualt port for HTTPS
		addr = ":443"
	}
	rabbitAddr := os.Getenv("RABBITADDR")
	if len(rabbitAddr) == 0 {
		rabbitAddr = "amqp://guest:guest@rabbit:5672/"
	}
	msgAddrs := strings.Split(os.Getenv("MESSAGESADDR"), ",")
	summaryAddrs := strings.Split(os.Getenv("SUMMARYADDR"), ",")
	if len(msgAddrs) == 0 {
		msgAddrs = append(msgAddrs, "http://micro-messaging:4000")
	}
	if len(summaryAddrs) == 0 {
		summaryAddrs = append(summaryAddrs, "http://micro-summary:8080")
	}

	//get the TLS key and cert paths from environment variables
	//this allows us to use a self-signed cert/key during development
	//and the Let's Encrypt cert/key in production
	tlsCertPath, tlsCertExists := os.LookupEnv("TLSCERT")
	if !tlsCertExists {
		log.Fatalf("Environment variable TLSCERT not defined.")
		os.Exit(1)
	}
	tlsKeyPath, tlsKeyExists := os.LookupEnv("TLSKEY")
	if !tlsKeyExists {
		log.Fatalf("Environment variable TLSKEY not defined.")
		os.Exit(1)
	}
	sessionKey, sessionKeyExists := os.LookupEnv("SESSIONKEY")
	if !sessionKeyExists {
		log.Fatalf("Environment variable SESSIONKEY not defined.")
		os.Exit(1)
	}
	redisAddr, redisAddrExists := os.LookupEnv("REDISADDR")
	if !redisAddrExists {
		log.Fatalf("Environment variable REDISADDR not defined.")
		os.Exit(1)
	}
	dsn, dsnExists := os.LookupEnv("DSN")
	if !dsnExists {
		log.Fatalf("Environment variable DSN not defined.")
		os.Exit(1)
	}

	// Init sql database
	userStore, postgressErr := users.ConnectToPostgres(dsn)
	if postgressErr != nil {
		log.Fatalf("Postgress not working")
		os.Exit(1)
	}
	err := userStore.LoadTrie()
	if err != nil {
		log.Fatalf("Failed to load trie struct")
		os.Exit(1)
	}

	// Init Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	sessionDuration, _ := time.ParseDuration("1h")
	redisSession := sessions.NewRedisStore(redisClient, sessionDuration)

	//Init RabbitMQ
	rabbitConn, err := amqp.Dial(rabbitAddr)
	if err != nil {
		log.Fatalf("Rabbit server not available")
		os.Exit(1)
	}
	ch, err := rabbitConn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a rabbit channel")
		os.Exit(1)
	}
	defer func() {
		fmt.Println("Rabbit MQ connection closing")
		rabbitConn.Close()
		ch.Close()
	}()

	handlerContext := &handlers.SessionContext{
		Key:     sessionKey,
		Session: redisSession,
		User:    userStore,
	}
	websocketContext := &handlers.WebsocketContext{
		Context:       handlerContext,
		Connections:   make(map[int]*websocket.Conn),
		Lock:          &sync.Mutex{},
		RabbitChannel: ch,
	}

	websocketContext.StartRabbitConsumer()
	// 2.Create a new mux for the web server.
	mux := http.NewServeMux()

	var msgServerAddrs []*url.URL
	for _, msgAddr := range msgAddrs {
		msgSerAddr, _ := url.Parse(msgAddr)
		msgServerAddrs = append(msgServerAddrs, msgSerAddr)
	}
	var sumServerAddrs []*url.URL
	for _, sumAddr := range summaryAddrs {
		sumServerAddr, _ := url.Parse(sumAddr)
		sumServerAddrs = append(sumServerAddrs, sumServerAddr)
	}
	msgProxy := &httputil.ReverseProxy{Director: CustomDirector(msgServerAddrs, handlerContext)}
	summaryProxy := &httputil.ReverseProxy{Director: CustomDirector(sumServerAddrs, handlerContext)}
	mux.Handle("/v1/channels", msgProxy)
	mux.Handle("/v1/channels/", msgProxy)
	mux.Handle("/v1/messages", msgProxy)
	mux.Handle("/v1/messages/", msgProxy)
	mux.Handle("/v1/summary", summaryProxy)

	// 3.Tell the mux to call your handlers.SummaryHandler function
	// when the "/v1/summary" URL path is requested.
	// mux.HandleFunc("/v1/summary/", handlers.SummaryHandler)
	mux.HandleFunc("/v1/users", handlerContext.UsersHandler)
	mux.HandleFunc("/v1/users/", handlerContext.SpecificUserHandler)
	mux.HandleFunc("/v1/sessions", handlerContext.SessionsHandler)
	mux.HandleFunc("/v1/sessions/", handlerContext.SpecificSessionHandler)
	//Websocket connection
	mux.HandleFunc("/v1/ws", websocketContext.WebSocketHandler)
	//   4.Start a web server listening on the address you read from
	//   the environment variable, using the mux you created as
	//   the root handler. Use log.Fatal() to report any errors
	//   that occur when trying to start the web server.
	corsMux := handlers.NewCORS(mux)
	log.Printf("Server is listening on port %s", addr)
	log.Fatal(http.ListenAndServeTLS(addr, tlsCertPath, tlsKeyPath, corsMux))
}

//CustomDirector takes in session context and do authentication
func CustomDirector(targets []*url.URL, context *handlers.SessionContext) Director {
	var counter int32
	counter = 0

	return func(r *http.Request) {
		targ := targets[int(counter)%len(targets)]
		atomic.AddInt32(&counter, 1)

		//Authenticate user
		sessionState := &handlers.SessionState{}
		sessions.GetState(r, context.Key, context.Session, sessionState)
		//Get user from session state
		user := sessionState.User
		if user != nil {
			encoded, _ := json.Marshal(user)
			encodedUser := base64.StdEncoding.EncodeToString(encoded)
			r.Header.Add("X-User", encodedUser)
		}

		r.Host = targ.Host
		r.URL.Host = targ.Host
		r.URL.Scheme = targ.Scheme
	}
}
