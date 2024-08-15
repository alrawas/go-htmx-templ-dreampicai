package handler

import (
	"dreampicai/db"
	"dreampicai/pkg/kit/validate"
	"dreampicai/pkg/sb"
	"dreampicai/types"
	"dreampicai/view/auth"

	// "fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/nedpals/supabase-go"
)

const (
	sessionUserKey        = "user"
	sessionAccessTokenKey = "accessToken"
)

func HandleAccountSetupIndex(w http.ResponseWriter, r *http.Request) error {
	return render(r, w, auth.AccountSetup())
}
func HandleAccountSetupPost(w http.ResponseWriter, r *http.Request) error {
	params := auth.AccountSetupParams{
		Username: r.FormValue("username"),
	}
	var errors auth.AccountSetupErrors
	ok := validate.New(&params, validate.Fields{
		"Username": validate.Rules(validate.Min(2), validate.Max(50)),
	}).Validate(&errors)
	if !ok {
		return render(r, w, auth.AccountSetup())
	}
	user := getAuthenticatedUser(r)
	account := types.Account{
		UserID:   user.ID,
		Username: params.Username,
	}
	if err := db.CreateAccount(&account); err != nil {
		return err
	}
	return hxRedirect(w, r, "/")
}

// access_token=eyJhbGciOiJIUzI1NiIsImtpZCI6Ik8wVmU5YzVFSm40S24vYXUiLCJ0eXAiOiJKV1QifQ.eyJhdWQiOiJhdXRoZW50aWNhdGVkIiwiZXhwIjoxNzE0NzI2NzgxLCJpYXQiOjE3MTQ3MjMxODEsImlzcyI6Imh0dHBzOi8vanBsYWx3c2J4a2d3dWxoZnpkaXguc3VwYWJhc2UuY28vYXV0aC92MSIsInN1YiI6IjlhOWY2NGY0LWZkNGEtNDFiYi1iYjk4LTY2MGQ5NzBiNTJmOCIsImVtYWlsIjoiYWJvdWRpcmF3YXMrMTBAZ21haWwuY29tIiwicGhvbmUiOiIiLCJhcHBfbWV0YWRhdGEiOnsicHJvdmlkZXIiOiJlbWFpbCIsInByb3ZpZGVycyI6WyJlbWFpbCJdfSwidXNlcl9tZXRhZGF0YSI6eyJlbWFpbCI6ImFib3VkaXJhd2FzKzEwQGdtYWlsLmNvbSIsImVtYWlsX3ZlcmlmaWVkIjpmYWxzZSwicGhvbmVfdmVyaWZpZWQiOmZhbHNlLCJzdWIiOiI5YTlmNjRmNC1mZDRhLTQxYmItYmI5OC02NjBkOTcwYjUyZjgifSwicm9sZSI6ImF1dGhlbnRpY2F0ZWQiLCJhYWwiOiJhYWwxIiwiYW1yIjpbeyJtZXRob2QiOiJvdHAiLCJ0aW1lc3RhbXAiOjE3MTQ3MjMxODF9XSwic2Vzc2lvbl9pZCI6IjFkM2Y2NzJlLTE3MDctNDZkZC05YzI4LWY3NWJiYWJlMzQzOCIsImlzX2Fub255bW91cyI6ZmFsc2V9.cGseMS6DZDSaRrved6iuhQsYHfuJwVXc5RCZ3ql_NbM&expires_at=1714726781&expires_in=3600&refresh_token=HkQscpeUGhHKoisewELKrw&token_type=bearer&type=signup
// func HandleSignupIndex(w http.ResponseWriter, r *http.Request) error {
// 	return render(r, w, auth.Signup())
// }

// func HandleSignupPost(w http.ResponseWriter, r *http.Request) error {

// 	params := auth.SignupParams{
// 		Email:           r.FormValue("email"),
// 		Password:        r.FormValue("password"),
// 		ConfirmPassword: r.FormValue("confirmPassword"),
// 	}

// 	errors := auth.SignupErrors{}
// 	if ok := validate.New(&params, validate.Fields{
// 		"Email":    validate.Rules(validate.Email),
// 		"Password": validate.Rules(validate.Password),
// 		"ConfirmPassword": validate.Rules(
// 			validate.Equal(params.Password),
// 			validate.Message("passwords do not match"),
// 		),
// 	}).Validate(&errors); !ok {
// 		return render(r, w, auth.SignupForm(params, errors))
// 	}

// 	user, err := sb.Client.Auth.SignUp(r.Context(), supabase.UserCredentials{
// 		Email:    params.Email,
// 		Password: params.Password,
// 	})
// 	if err != nil {
// 		return err
// 	}

// 	return render(r, w, auth.SignupSuccess(user.Email))
// }

func HandleLoginIndex(w http.ResponseWriter, r *http.Request) error {
	return render(r, w, auth.Login())
}

func HandleLoginWithGoogle(w http.ResponseWriter, r *http.Request) error {
	resp, err := sb.Client.Auth.SignInWithProvider(supabase.ProviderSignInOptions{
		Provider:   "google",
		RedirectTo: "http://localhost:3000/auth/callback",
	})
	if err != nil {
		return err
	}
	http.Redirect(w, r, resp.URL, http.StatusSeeOther)
	return nil
}

func HandleLoginPost(w http.ResponseWriter, r *http.Request) error {
	credentials := supabase.UserCredentials{
		Email: r.FormValue("email"),
		// Password: r.FormValue("password"),
	}

	// call supabase
	// resp, err := sb.Client.Auth.SignIn(r.Context(), credentials)
	err := sb.Client.Auth.SendMagicLink(r.Context(), credentials.Email)

	if err != nil {
		slog.Error("login error", "err", err)
		return render(r, w, auth.LoginForm(credentials, auth.LoginErrors{
			InvalidCredentials: "The credentials you have entered are invalid",
		}))
	}
	// if err := setAuthSession(w, r, resp.AccessToken); err != nil {
	// 	return err
	// }
	// return hxRedirect(w, r, "/")
	return auth.SigninMagicSuccess(credentials.Email).Render(r.Context(), w)
}

func HandleAuthCallbak(w http.ResponseWriter, r *http.Request) error {
	accessToken := r.URL.Query().Get("access_token")
	if len(accessToken) == 0 {
		return render(r, w, auth.CallbackScript())
	}
	// fmt.Println(accessToken)
	if err := setAuthSession(w, r, accessToken); err != nil {
		return err
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

func HandleLogoutPost(w http.ResponseWriter, r *http.Request) error {
	store := sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))
	session, _ := store.Get(r, sessionUserKey)
	session.Values[sessionAccessTokenKey] = ""
	session.Save(r, w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
	return nil
}

func setAuthSession(w http.ResponseWriter, r *http.Request, accessToken string) error {
	store := sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))
	session, _ := store.Get(r, sessionUserKey)
	session.Values[sessionAccessTokenKey] = accessToken
	return sessions.Save(r, w)
}
