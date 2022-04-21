package main

import (
	"log"
	"net/http"
)

func serveProfiles(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodDelete {
		deleteProfile(w, r)
	} else {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
}

func deleteProfile(w http.ResponseWriter, r *http.Request) {
	if !verifyRoute(w, r, http.MethodDelete, "/profiles") {
		return
	}

	var user *User = r.Context().Value(userContextKey).(*User)

	lobAddressId, err := postgresClient.getLobAddressId(user.Id)
	if err != nil {
		log.Println(err)
		http.Error(w, "No profile found that corresponds to this user.", http.StatusNotFound)
		return
	}

	if lobAddressId != "" {
		if err := lobClient.DeleteAddress(lobAddressId, true); err != nil {
			log.Println(err)
			http.Error(w, "Error deleting address", http.StatusInternalServerError)
			return
		}
	}

	if err = postgresClient.deleteUser(user.Id); err != nil {
		log.Println(err)
		http.Error(w, "Error setting address in database", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	return
}
