package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"

	store "github.com/christophberger/fingerprintjs-go/internal/store"
	"golang.org/x/crypto/bcrypt"

	"github.com/joho/godotenv"
)

// embed HTML and CSS files
var (
	//go:embed style.css
	style []byte

	//go:embed home.html
	home []byte

	//go:embed signup.gotpl
	signupTpl string

	//go:embed response.gotpl
	responseTpl string
)

func run() error {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		return fmt.Errorf("load .env: %w", err)
	}

	// Connect to the database
	users, err := store.NewUsers(os.Getenv("FINGERPRINT_DATABASE_PATH"))
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer users.Close()

	// signup.gotpl is a Go template. Map the environment variable FINGERPRINT_PUBLIC_KEY to the "{{ . }}" placeholder in the template.
	tmplSignup := template.Must(template.New("signup").Parse(signupTpl))
	var signup bytes.Buffer
	err = tmplSignup.Execute(&signup, os.Getenv("FINGERPRINT_PUBLIC_KEY"))
	if err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	// parse response.gotpl
	tmplResponse := template.Must(template.New("response").Parse(responseTpl))

	// Define and register handlers for the homepage, signup page, stylesheet, and signup request
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write(home)
		if err != nil {
			log.Printf("serve home page: %s\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	})

	http.HandleFunc("/signupform", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write(signup.Bytes())
		if err != nil {
			log.Printf("serve signup form: %s\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	})

	http.HandleFunc("/css/style.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		_, err := w.Write(style)
		if err != nil {
			log.Printf("serve stylesheet: %s\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	})

	http.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")
		visitorId := r.FormValue("visitorId")

		log.Printf("Email: %s, Visitor ID: %s\n", email, visitorId)

		msg := "Thank you for signing up!"

		// Check if the visitor ID already exists in the database
		visitorExists, err := users.Check(visitorId)
		if err != nil {
			log.Printf("/signup: check visitor ID: %s\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if visitorExists {
			msg = "You already have signed up. Please log in with your existing account."
		} else {

			// Add the user to the database
			pwHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				log.Printf("/signup: hash password: %s\n", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			err = users.Add(email, string(pwHash), visitorId)
			if err != nil {
				log.Printf("/signup: add user: %s\n", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
		}

		// Send the response (either "thank you" or "you already signed up")
		w.Header().Add("Location", "/response")
		var response bytes.Buffer
		err = tmplResponse.ExecuteTemplate(&response, "response", msg)
		if err != nil {
			log.Printf("/signup: execute template: %s\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		_, err = w.Write(response.Bytes())
		if err != nil {
			log.Printf("/signup: write response: %s\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	})

	// Start the server
	port := os.Getenv("FINGERPRINT_LOCAL_PORT")
	log.Printf("Starting server at http://localhost:%s\n", port)
	err = http.ListenAndServe("127.0.0.1:"+port, nil)
	if err != nil {
		return fmt.Errorf("ListenAndServe: %w\n", err)
	}

	return nil
}

func main() {
	err := run()
	if err != nil {
		log.Fatalf("%s\n", err)
	}
}
