package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	openrouter "github.com/OpenRouterTeam/go-sdk"
	"github.com/OpenRouterTeam/go-sdk/models/components"
	"github.com/OpenRouterTeam/go-sdk/optionalnullable"
)

const jevModel = "typesafe/jev-1.13"

func main() {
	// Charger le .env si présent ; l'environnement réel garde la priorité
	if err := loadDotEnv(".env"); err == nil {
		log.Println("fichier .env chargé")
	}

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		log.Fatal("L'environnement OPENROUTER_API_KEY n'est pas défini")
	}

	ctx := context.Background()
	client := openrouter.New(
		openrouter.WithSecurity(apiKey),
		openrouter.WithTimeout(30*time.Second),
		openrouter.WithHTTPReferer("https://github.com/hermes-agent/jevtest"),
		openrouter.WithXTitle("Jev Test App"),
	)

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

	questions := map[string]components.Questions{
		"is_ironic": components.CreateQuestionsNoul(components.DecisionsNoulQuestion{
			Type: components.DecisionsNoulQuestionTypeNoul,
			Instructions: optionalnullable.From(openrouter.Pointer(
				components.CreateDecisionsNoulQuestionInstructionsStr("Est-ce que ce texte est ironique ?"),
			)),
			Criteria: &components.DecisionsNoulQuestionCriteria{
				True:  optionalnullable.From(openrouter.Pointer(components.CreateTrueStr("ironique"))),
				False: optionalnullable.From(openrouter.Pointer(components.CreateFalseStr("sincère"))),
			},
		}),
		"degree": components.CreateQuestionsChoice(components.DecisionsChoiceQuestion{
			Type: components.DecisionsChoiceQuestionTypeChoice,
			Instructions: optionalnullable.From(openrouter.Pointer(
				components.CreateDecisionsChoiceQuestionInstructionsStr("Degré ?"),
			)),
			Criteria: map[string]*components.Criteria{
				"Premier degré": openrouter.Pointer(components.CreateCriteriaStr("sincère")),
				"Second degré":  openrouter.Pointer(components.CreateCriteriaStr("ironique")),
			},
		}),
	}

	totalTime := time.Duration(0)

	for i, tc := range tests {
		start := time.Now()
		resp, err := client.Alpha.Decisions.Create(ctx, components.DecisionsRequest{
			Model:     jevModel,
			State:     components.CreateStateStr(tc.text),
			Questions: questions,
		})
		elapsed := time.Since(start)
		totalTime += elapsed

		if err != nil {
			fmt.Printf("[%d] Erreur : %v\n", i+1, err)
			continue
		}

		fmt.Printf("\n=== Cas %d : %s ===\n", i+1, tc.comment)
		fmt.Printf("Texte : %q\n", tc.text)

		if ans, ok := resp.Answers["degree"]; ok && ans.DecisionsChoiceAnswer != nil {
			deg := ans.DecisionsChoiceAnswer

			probs := ""
			for _, opt := range []string{"Premier degré", "Second degré"} {
				probs += fmt.Sprintf("%s: %.0f%%  ", opt, deg.Probabilities[opt]*100)
			}

			var conf float64
			if deg.Confidence != nil {
				conf = *deg.Confidence
			}

			verdict := ""
			switch {
			case conf >= 0.75:
				verdict = "TRANCHE"
			case conf >= 0.5:
				verdict = "DOUTE"
			default:
				verdict = "TREMBLE (considérer comme 1er degré par défaut)"
			}

			fmt.Printf("  CHOICE -> %-15s (%s) | %s\n", deg.Choice, verdict, probs)
		}

		if ans, ok := resp.Answers["is_ironic"]; ok && ans.DecisionsNoulAnswer != nil {
			fmt.Printf("  NOUL   -> probabilité ironique: %.0f%%\n", ans.DecisionsNoulAnswer.Noul*100)
		}

		fmt.Printf("  Attendu: %s | Temps: %v\n", tc.expected, elapsed)
	}

	fmt.Printf("\nTemps total pour %d requêtes : %v (moyenne %v)\n",
		len(tests), totalTime, totalTime/time.Duration(len(tests)))
}
