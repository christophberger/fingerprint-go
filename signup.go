package main

import (
	"bytes"
	"log"
	"net/http"
	"text/template"

	"github.com/christophberger/fingerprint-go/internal/fingerprint"
	"github.com/christophberger/fingerprint-go/internal/store"
)

func setSignupHandler(users *store.Users, tmplResponse *template.Template) {
	// Define and register the handler for the signup form
	http.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		email := r.FormValue("email")
		visitorId := r.FormValue("visitorId")
		requestId := r.FormValue("requestId")

		log.Printf("Email: %s, Visitor ID: %s\n", email, visitorId)

		// Check if the visitor ID already exists in the database
		recentSignup, err := users.Check(visitorId)
		if err != nil {
			log.Printf("/signup: check visitor ID: %s\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		msg := ""
		if recentSignup {
			msg = "Someone else has signed up from this device in the last minute! To prevent fraudulent mass signups, we restricted the number of signups per device to one signup per minute. Please try again later."
		} else {

			// Add the user to the database
			msg, err = users.Add(email, visitorId)
			if err != nil {
				log.Printf("/signup: add user: %s\n", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
		}
		// Get additional client information through the Go SDK

		log.Printf("Server-side check for request ID %s\n", requestId)
		fp := fingerprint.New()
		success, err := fp.Validate(requestId, visitorId)
		if err != nil {
			log.Printf("/signup: validate fingerprint: %s\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if !success {
			msg = "Error verifying the signup attempt. Please try again."
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
}
