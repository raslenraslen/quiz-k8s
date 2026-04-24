package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
)

type Question struct {
	ID      int      `json:"id"`
	Text    string   `json:"text"`
	Options []string `json:"options"`
	Answer  int      `json:"answer"`
}

type VoteRequest struct {
	QuestionID int `json:"question_id"`
	Option     int `json:"option"`
}

type Results struct {
	QuestionID int            `json:"question_id"`
	Question   string         `json:"question"`
	Options    []string       `json:"options"`
	Votes      map[int]int    `json:"votes"`
	Total      int            `json:"total"`
	Pod        string         `json:"pod"`
}

var (
	questions = []Question{
		{
			ID:      1,
			Text:    "What is Kubernetes?",
			Options: []string{"A container orchestration tool", "A programming language for servers", "A cloud hosting platform", "A Linux operating system"},
			Answer:  0,
		},
		{
			ID:      2,
			Text:    "What is a Pod in Kubernetes?",
			Options: []string{"A virtual machine in the cloud", "A YAML configuration file", "The smallest deployable unit", "A physical server in the rack"},
			Answer:  2,
		},
		{
			ID:      3,
			Text:    "What command creates a Deployment?",
			Options: []string{"kubectl start deployment", "kubectl deploy my-app", "kubectl create deployment", "kubectl new deployment"},
			Answer:  2,
		},
		{
			ID:      4,
			Text:    "What happens when you delete a Pod without a Deployment?",
			Options: []string{"Kubernetes recreates it after 30s", "The Pod is deleted permanently", "The cluster restarts the node", "A new Pod appears on another node"},
			Answer:  1,
		},
		{
			ID:      5,
			Text:    "What does 'kubectl get pods' show you?",
			Options: []string{"All the YAML files on disk", "The list of cluster nodes", "The pods running in the cluster", "The Docker images available"},
			Answer:  2,
		},
		{
			ID:      6,
			Text:    "What is a Namespace used for?",
			Options: []string{"To store secret passwords", "To connect pods to the internet", "To isolate resources in the cluster", "To create new Docker images"},
			Answer:  2,
		},
		{
			ID:      7,
			Text:    "What does 'replicas: 3' mean in a Deployment?",
			Options: []string{"3 containers inside one Pod", "3 copies of your app running", "3 nodes added to the cluster", "3 services created for the app"},
			Answer:  1,
		},
		{
			ID:      8,
			Text:    "What does the Control Plane do?",
			Options: []string{"It runs your app containers", "It makes decisions for the cluster", "It stores the Docker images", "It connects users to the internet"},
			Answer:  1,
		},
		{
			ID:      9,
			Text:    "What is k3s?",
			Options: []string{"A monitoring tool for clusters", "A container runtime for Docker", "A lightweight Kubernetes distribution", "A network plugin for Kubernetes"},
			Answer:  2,
		},
		{
			ID:      10,
			Text:    "What does 'kubectl apply -f manifest.yaml' do?",
			Options: []string{"It deletes all resources in the file", "It opens the file in a text editor", "It sends the file to Docker Hub", "It creates resources from the YAML file"},
			Answer:  3,
		},
		{
			ID:      11,
			Text:    "What component decides which node runs a new Pod?",
			Options: []string{"The kube-proxy on each node", "The Scheduler in the control plane", "The kubelet on the worker node", "The etcd key-value store"},
			Answer:  1,
		},
		{
			ID:      12,
			Text:    "What component stores the full state of the cluster?",
			Options: []string{"The API Server component", "The Controller Manager process", "The kubelet on each node", "The etcd key-value store"},
			Answer:  3,
		},
		{
			ID:      13,
			Text:    "What type of Service exposes your app outside the cluster?",
			Options: []string{"ClusterIP Service type", "NodePort Service type", "InternalDNS Service type", "PodConnect Service type"},
			Answer:  1,
		},
		{
			ID:      14,
			Text:    "What happens when you delete a Pod managed by a Deployment?",
			Options: []string{"The Deployment creates a replacement", "The Pod is gone permanently", "You need to restart the cluster", "The Service stops all traffic"},
			Answer:  0,
		},
		{
			ID:      15,
			Text:    "How do services find each other inside the cluster?",
			Options: []string{"By using SSH connections", "By using DNS names internally", "By using MAC addresses directly", "By reading YAML files at runtime"},
			Answer:  1,
		},
		{
			ID:      16,
			Text:    "What is the role of the API Server?",
			Options: []string{"It runs all the containers", "It stores data permanently", "It balances network traffic", "It receives all commands and requests"},
			Answer:  3,
		},
		{
			ID:      17,
			Text:    "What is the DNS format for a Kubernetes service?",
			Options: []string{"pod-ip.namespace.cluster.dns", "node-name.service.k8s.internal", "service.namespace.svc.cluster.local", "container.pod.node.k8s.network"},
			Answer:  2,
		},
		{
			ID:      18,
			Text:    "What does kube-proxy do?",
			Options: []string{"It stores the cluster state", "It schedules pods on nodes", "It runs containers on the node", "It routes traffic to the right pod"},
			Answer:  3,
		},
		{
			ID:      19,
			Text:    "What are the 4 main parts of a manifest?",
			Options: []string{"name, image, port, label", "cpu, memory, disk, network", "pod, service, deploy, namespace", "apiVersion, kind, metadata, spec"},
			Answer:  3,
		},
		{
			ID:      20,
			Text:    "What is the difference between ClusterIP and NodePort?",
			Options: []string{"ClusterIP is internal, NodePort is external", "NodePort is internal, ClusterIP is external", "ClusterIP is faster than NodePort", "NodePort needs more memory to run"},
			Answer:  0,
		},
	}

	votes = make(map[int]map[int]int)
	mu    sync.RWMutex
)

func init() {
	for _, q := range questions {
		votes[q.ID] = make(map[int]int)
	}
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	pod := os.Getenv("HOSTNAME")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"pod":    pod,
	})
}

func questionsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// Return questions without answers
	type SafeQuestion struct {
		ID      int      `json:"id"`
		Text    string   `json:"text"`
		Options []string `json:"options"`
	}
	safe := make([]SafeQuestion, len(questions))
	for i, q := range questions {
		safe[i] = SafeQuestion{ID: q.ID, Text: q.Text, Options: q.Options}
	}
	json.NewEncoder(w).Encode(safe)
}

func voteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var vote VoteRequest
	if err := json.NewDecoder(r.Body).Decode(&vote); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	mu.Lock()
	if _, ok := votes[vote.QuestionID]; ok {
		votes[vote.QuestionID][vote.Option]++
	}
	mu.Unlock()

	pod := os.Getenv("HOSTNAME")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "voted",
		"pod":    pod,
	})
}

func resultsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	pod := os.Getenv("HOSTNAME")

	mu.RLock()
	defer mu.RUnlock()

	allResults := make([]Results, len(questions))
	for i, q := range questions {
		total := 0
		for _, count := range votes[q.ID] {
			total += count
		}
		allResults[i] = Results{
			QuestionID: q.ID,
			Question:   q.Text,
			Options:    q.Options,
			Votes:      votes[q.ID],
			Total:      total,
			Pod:        pod,
		}
	}
	json.NewEncoder(w).Encode(allResults)
}

func revealHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	answers := make(map[int]int)
	for _, q := range questions {
		answers[q.ID] = q.Answer
	}
	json.NewEncoder(w).Encode(answers)
}

func resetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	mu.Lock()
	for _, q := range questions {
		votes[q.ID] = make(map[int]int)
	}
	mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "reset"})
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/api/health", corsMiddleware(healthHandler))
	http.HandleFunc("/api/questions", corsMiddleware(questionsHandler))
	http.HandleFunc("/api/vote", corsMiddleware(voteHandler))
	http.HandleFunc("/api/results", corsMiddleware(resultsHandler))
	http.HandleFunc("/api/reveal", corsMiddleware(revealHandler))
	http.HandleFunc("/api/reset", corsMiddleware(resetHandler))

	fmt.Printf("Quiz backend running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
