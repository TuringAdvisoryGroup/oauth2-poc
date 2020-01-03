package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx"
	pg "github.com/vgarvardt/go-oauth2-pg"
	"github.com/vgarvardt/go-pg-adapter/pgxadapter"

	"gopkg.in/oauth2.v3/errors"
	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/models"
	"gopkg.in/oauth2.v3/server"
)

func main() {
	pgxConnConfig, _ := pgx.ParseURI(os.Getenv("DB_URL"))
	pgxConn, _ := pgx.Connect(pgxConnConfig)

	manager := manage.NewDefaultManager()
	adapter := pgxadapter.NewConn(pgxConn)

	// token pg store
	tokenStore, _ := pg.NewTokenStore(adapter, pg.WithTokenStoreGCInterval(time.Minute))
	defer tokenStore.Close()

	// Manager Mapping
	manager.MapTokenStorage(tokenStore)

	// client memory store
	clientStore, _ := pg.NewClientStore(adapter)
	manager.MapClientStorage(clientStore)

	clientStore.Create(&models.Client{
		ID:     "000000",
		Secret: "999999",
		Domain: "http://localhost:9094/",
	})

	srv := server.NewDefaultServer(manager)
	srv.SetAllowGetAccessRequest(true)
	srv.SetClientInfoHandler(server.ClientFormHandler)

	srv.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		log.Println("Internal Error:", err.Error())
		return
	})

	srv.SetResponseErrorHandler(func(re *errors.Response) {
		log.Println("Response Error:", re.Error.Error())
	})

	http.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
		err := srv.HandleAuthorizeRequest(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	})

	http.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		srv.HandleTokenRequest(w, r)
	})

	log.Fatal(http.ListenAndServe(":9096", nil))
}
