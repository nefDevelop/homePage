package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
)

type Container struct {
	ID     string   `json:"Id"`
	Image  string   `json:"Image"`
	Names  []string `json:"Names"`
	State  string   `json:"State"`
	Status string   `json:"Status"`
}

// Conectar al socket de Docker
func dockerRequest(method, endpoint string) ([]byte, error) {
	conn, err := net.Dial("unix", "/var/run/docker.sock")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// Enviar la petición
	fmt.Fprintf(conn, "%s %s HTTP/1.1\r\nHost: docker\r\n\r\n", method, endpoint)

	// Leer la respuesta
	response, err := io.ReadAll(conn)
	if err != nil {
		return nil, err
	}

	// Buscar el cuerpo JSON de la respuesta
	start := string(response)
	startIndex := len("HTTP/1.1 200 OK\r\n")
	return []byte(start[startIndex:]), nil
}

// Obtener lista de contenedores
func getContainers(w http.ResponseWriter, r *http.Request) {
	data, err := dockerRequest("GET", "/containers/json?all=true")
	if err != nil {
		http.Error(w, "Error obteniendo contenedores", http.StatusInternalServerError)
		return
	}

	var containers []Container
	if err := json.Unmarshal(data, &containers); err != nil {
		http.Error(w, "Error procesando JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(containers)
}

// Controlar contenedores (start, stop, remove)
func controlContainer(w http.ResponseWriter, r *http.Request) {
	action := r.URL.Query().Get("action") // start, stop, remove
	id := r.URL.Query().Get("id")         // ID del contenedor

	if action == "" || id == "" {
		http.Error(w, "Faltan parámetros", http.StatusBadRequest)
		return
	}

	endpoint := fmt.Sprintf("/containers/%s/%s", id, action)
	_, err := dockerRequest("POST", endpoint)
	if err != nil {
		http.Error(w, "Error ejecutando acción", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Contenedor %s: %s exitosamente", id, action)
}

func main() {
	http.HandleFunc("/containers", getContainers)
	http.HandleFunc("/control", controlContainer)

	log.Println("Servidor en http://localhost:3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
