package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// loadDotEnv lit un fichier .env ligne par ligne et définit les variables
// d'environnement correspondantes (uniquement si pas déjà définies).
// Format supporté : CLE=valeur, avec guillemets simples ou doubles optionnels.
// Les lignes vides, spaces-only, et commentaires (#) sont ignorés.
// Retourne une erreur si le fichier existe mais n'est pas lisible.
func loadDotEnv(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lnum := 0
	for scanner.Scan() {
		lnum++
		line := strings.TrimSpace(scanner.Text())

		// Ignorer les lignes vides et les commentaires
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Retirer le préfixe `export ` si présent
		line = strings.TrimPrefix(line, "export ")

		// Split sur le premier '=' uniquement
		idx := strings.Index(line, "=")
		if idx <= 0 {
			return fmt.Errorf("%s:%d: ligne invalide (attendu CLE=valeur)", path, lnum)
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])

		// Retirer les guillemets qui entourent la valeur entière
		if len(val) >= 2 {
			if (strings.HasPrefix(val, `"`) && strings.HasSuffix(val, `"`)) ||
				(strings.HasPrefix(val, `'`) && strings.HasSuffix(val, `'`)) {
				val = val[1 : len(val)-1]
			}
		}

		// Ne pas écraser une variable déjà définie dans l'environnement
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		os.Setenv(key, val)
	}
	return scanner.Err()
}
