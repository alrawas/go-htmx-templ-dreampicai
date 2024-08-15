package handler

import (
	"dreampicai/view/home"
	"fmt"
	"net/http"
	"time"
)

func HandleLongProcess(w http.ResponseWriter, r *http.Request) error {
	time.Sleep(5 * time.Second)
	return home.UserLikes(1000).Render(r.Context(), w)
}
func HandleHomeIndex(w http.ResponseWriter, r *http.Request) error {
	user := getAuthenticatedUser(r)
	// account, err := db.GetAccountByUserID(user.ID)
	// if err != nil {
	// 	return err
	// }
	// // +v is super verbose, field name and filed values. not only values
	fmt.Printf("%+v\n", user.Account)
	return home.Index().Render(r.Context(), w)
}
