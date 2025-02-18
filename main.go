package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"

	"blog/api"
	"blog/config"
	"blog/web"

	_ "blog/docs"

	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	ip := flag.String("ip", "localhost", "IP address to bind to")
	port := flag.String("port", "8080", "Port to listen on")
	secret := flag.String("secret", "secret", "Secret key for authentication")
	dbfile := flag.String("dbfile", "blog.db", "Path to the database file")
	init := flag.Bool("init", false, "Initialize the application")

	flag.Parse()

	config.IP = *ip
	config.Port = *port
	config.SecretStr = *secret
	config.DBFile = *dbfile

	err := config.Setup()
	if err != nil {
		log.Fatal(err)
	}
	defer config.DB.Close()

	if *init {
		err = config.Reset()
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Application is successfully initialized")
		os.Exit(0)
	}

	var srv http.Server

	rootMux := http.NewServeMux()
	apiMux := api.ServeMux()
	webMux := web.ServeMux()
	rootMux.Handle("/", http.RedirectHandler("/web/posts/get", http.StatusSeeOther))
	rootMux.Handle("/api/", http.StripPrefix("/api", apiMux))
	rootMux.Handle("/web/", http.StripPrefix("/web", webMux))
	rootMux.Handle("/swagger/", httpSwagger.WrapHandler)

	idleClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		log.Println("Server is running at", config.Host)
		<-sigint
		err = srv.Shutdown(config.Ctx)
		if err != nil {
			log.Println(err)
		}
		close(idleClosed)
		log.Println("Server is shutting down")
	}()

	srv.Addr = config.Addr
	srv.Handler = rootMux
	err = srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
	<-idleClosed
}
