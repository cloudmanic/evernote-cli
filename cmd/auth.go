package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

// Evernote's OAuth2 endpoints would go here. These are placeholders.
var evernoteEndpoint = oauth2.Endpoint{
	AuthURL:  "https://www.evernote.com/oauth2/authorize",
	TokenURL: "https://www.evernote.com/oauth2/token",
}

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with Evernote",
	RunE: func(cmd *cobra.Command, args []string) error {
		clientID := os.Getenv("EVERNOTE_CLIENT_ID")
		clientSecret := os.Getenv("EVERNOTE_CLIENT_SECRET")
		cfg, _ := loadConfig()
		if clientID == "" && cfg != nil {
			clientID = cfg.ClientID
		}
		if clientSecret == "" && cfg != nil {
			clientSecret = cfg.ClientSecret
		}
		if clientID == "" || clientSecret == "" {
			return fmt.Errorf("client ID and secret must be provided (run 'evernote-cli init')")
		}

		ctx := context.Background()
		conf := &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Scopes:       []string{"basic"},
			Endpoint:     evernoteEndpoint,
			RedirectURL:  "http://localhost:8080/callback",
		}

		// Start a local server to receive the callback
		srv := &http.Server{Addr: ":8080"}
		codeCh := make(chan string)
		http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
			code := r.URL.Query().Get("code")
			fmt.Fprintf(w, "Authentication complete. You can close this window.")
			codeCh <- code
		})
		go func() {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("server error: %v", err)
			}
		}()

		url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
		// Open browser if possible
		exec.Command("xdg-open", url).Start()
		fmt.Printf("If the browser did not open, visit: %s\n", url)

		code := <-codeCh
		srv.Shutdown(ctx)

		token, err := conf.Exchange(ctx, code)
		if err != nil {
			return fmt.Errorf("token exchange failed: %w", err)
		}

		if cfg == nil {
			cfg = &Config{}
		}
		cfg.ClientID = clientID
		cfg.ClientSecret = clientSecret
		cfg.Token = token

		if err := saveConfig(cfg); err != nil {
			return err
		}
		fmt.Println("Authentication successful. Token saved.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
}
