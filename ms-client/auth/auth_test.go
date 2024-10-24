package auth

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	msgraph "github.com/microsoftgraph/msgraph-sdk-go"
	"net/http"
	"os"
	"testing"
	"time"
)

// Test for MS Client Authentication
func TestLogin(t *testing.T) {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	params := AuthParams{
		clientID:     os.Getenv("CLIENT_ID"),
		clientSecret: os.Getenv("CLIENT_SECRET"),
		tenantID:     os.Getenv("TENANT_ID"),
		authorityURL: os.Getenv("AUTHORITY_URL"),
		redirectURI:  os.Getenv("REDIRECT_URI"),
		scopes:       []string{"User.Read"},
	}

	credential, err := GetCredential(params.clientSecret)
	if err != nil {
		panic(err)
	}

	client, err := GetClient(params.clientID, params.tenantID, params.authorityURL, credential)
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		url, err := GetAuthURL(client, params.clientID, params.redirectURI, params.scopes)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, url, http.StatusFound)
	})

	http.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Authorization code not found", http.StatusBadRequest)
			return
		}

		token, err := client.AcquireTokenByAuthCode(context.TODO(), code, params.redirectURI, params.scopes)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "access_token",
			Value:    token.AccessToken,
			Expires:  time.Now().Add(time.Hour),
			HttpOnly: true,
		})

		http.Redirect(w, r, "/me", http.StatusFound)
	})

	http.HandleFunc("/me", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("access_token")
		if err != nil {
			fmt.Println("No user is logged in")
		}

		userToken := &TokenCredential{accessToken: cookie.Value}

		provider, err := GetAuthProvider(userToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		adapter, err := msgraph.NewGraphRequestAdapter(provider)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		graphClient := msgraph.NewGraphServiceClient(adapter)
		user, err := graphClient.Me().Get(context.TODO(), nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err := fmt.Fprintf(w, "Signed-in User: %s\nUser Principal Name: %s", *user.GetDisplayName(), *user.GetUserPrincipalName()); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	})

	fmt.Println("Listening on port 5000")
	if err := http.ListenAndServe(":5000", nil); err != nil {
		panic(err)
	}
}
