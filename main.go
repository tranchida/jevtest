package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type Criteria map[string]string

type Question struct {
	Type         string   `json:"type"`
	Instructions string   `json:"instructions"`
	Criteria     Criteria `json:"criteria"`
	Options      []string `json:"options,omitempty"`
}

type JevRequest struct {
	Model     string              `json:"model"`
	State     string              `json:"state"`
	Questions map[string]Question `json:"questions"`
}

type JevAnswer struct {
	Type          string             `json:"type"`
	Choice        string             `json:"choice,omitempty"`
	Noul          float64            `json:"noul,omitempty"`
	Confidence    float64            `json:"confidence,omitempty"`
	Probabilities map[string]float64 `json:"probabilities,omitempty"`
}

type JevResponse struct {
	Answers map[string]JevAnswer `json:"answers"`
}

type OpenRouterJevClient struct {
	APIKey     string
	HTTPClient *http.Client
	Endpoint   string
	Model      string
}

func NewOpenRouterClient(apiKey string) *OpenRouterJevClient {
	return &OpenRouterJevClient{
		APIKey: apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		Endpoint: "https://openrouter.ai/api/alpha/decisions",
		Model:    "typesafe/jev-1.13",
	}
}

func (c *OpenRouterJevClient) Ask(ctx context.Context, state string, questions map[string]Question) (map[string]JevAnswer, time.Duration, error) {
	reqBody := JevRequest{
		Model:     c.Model,
		State:     state,
		Questions: questions,
	}

	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequestWithContext(ctx, "POST", c.Endpoint, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("HTTP-Referer", "https://github.com/hermes-agent/jevtest")
	req.Header.Set("X-Title", "Jev Test App")

	start := time.Now()
	resp, err := c.HTTPClient.Do(req)
	elapsed := time.Since(start)

	if err != nil {
		return nil, elapsed, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, elapsed, fmt.Errorf("API error: status %d", resp.StatusCode)
	}

	var result JevResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, elapsed, err
	}

	return result.Answers, elapsed, nil
}

func main() {
	// Charger le .env si présent ; l'environment réel garde la priorité
	if err := loadDotEnv(".env"); err == nil {
		log.Println("fichier .env chargé")
	}

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		log.Fatal("L'environnement OPENROUTER_API_KEY n'est pas défini")
	}

	ctx := context.Background()
	client := NewOpenRouterClient(apiKey)

	// Batterie de cas délicats
	tests := []struct {
		text     string
		expected string
		comment  string
	}{
		{
			"Ah oui, bravo, t'as encore oublié mes clés au bureau. Franchement, chapeau.",
			"Second degré",
			"ironie classique sur un reproche",
		},
		{
			"Franchement, chapeau : tu as géré la panne du serveur à 3h du matin en solo.",
			"Premier degré",
			"ÉLOGE SINCÈRE avec les mêmes mots! Même marqueur ('chapeau'), intention opposée.",
		},
		{
			"Ce n'est pas exactement ce que j'appellerais une réussite, mais ça marche.",
			"Second degré",
			"Litote : critique adoucie, aucun marqueur d'exagération.",
		},
		{
			"Je suis mort de rire, cette vidéo est trop drôle, je pleure !",
			"Premier degré",
			"Hyperbole SINCÈRE : l'exagération est vraie, pas ironique.",
		},
		{
			"Après seulement 9 mois, ma demande de carte d'identité a enfin abouti. L'administration est d'une efficacité redoutable.",
			"Second degré",
			"Ironie par contraste factuel (9 mois = lent).",
		},
		{
			"Bravo, vraiment.",
			"? (dépend du contexte)",
			"Ultra-ambigu sans contexte : à vous de voir où Jev place le curseur.",
		},
	}

	questions := map[string]Question{
		"is_ironic": {
			Type:         "noul",
			Instructions: "Est-ce que ce texte est ironique ?",
			Criteria:     Criteria{"true": "ironique", "false": "sincère"},
		},
		"degree": {
			Type:         "choice",
			Instructions: "Degré ?",
			Options:      []string{"Premier degré", "Second degré"},
			Criteria:     Criteria{"Premier degré": "sincère", "Second degré": "ironique"},
		},
	}

	totalTime := time.Duration(0)

	for i, tc := range tests {
		answers, elapsed, err := client.Ask(ctx, tc.text, questions)
		totalTime += elapsed
		if err != nil {
			fmt.Printf("[%d] Erreur : %v\n", i+1, err)
			continue
		}

		fmt.Printf("\n=== Cas %d : %s ===\n", i+1, tc.comment)
		fmt.Printf("Texte : %q\n", tc.text)

		if deg, ok := answers["degree"]; ok && deg.Type == "choice" {
			// Distillation des probabilités par option
			probs := ""
			for _, opt := range []string{"Premier degré", "Second degré"} {
				probs += fmt.Sprintf("%s: %.0f%%  ", opt, deg.Probabilities[opt]*100)
			}

			verdict := ""
			switch {
			case deg.Confidence >= 0.75:
				verdict = "TRANCHE"
			case deg.Confidence >= 0.5:
				verdict = "DOUTE"
			default:
				verdict = "TREMBLE (considérer comme 1er degré par défaut)"
			}

			fmt.Printf("  CHOICE -> %-15s (%s) | %s\n", deg.Choice, verdict, probs)
		}

		if n, ok := answers["is_ironic"]; ok && n.Type == "noul" {
			fmt.Printf("  NOUL   -> probabilité ironique: %.0f%%\n", n.Noul*100)
		}

		fmt.Printf("  Attendu: %s | Temps: %v\n", tc.expected, elapsed)
	}

	fmt.Printf("\nTemps total pour %d requêtes : %v (moyenne %v)\n",
		len(tests), totalTime, totalTime/time.Duration(len(tests)))
}
