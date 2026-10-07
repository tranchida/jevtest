package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestLoadDotEnv(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantKey string
		wantVal string
		preSet  map[string]string // env à définir avant l'appel ; valeur attendue après
		wantErr bool
	}{
		{
			name:    "simple",
			content: "FOO=bar\n",
			wantKey: "FOO",
			wantVal: "bar",
		},
		{
			name:    "double quotes",
			content: "FOO=\"bar baz\"\n",
			wantKey: "FOO",
			wantVal: "bar baz",
		},
		{
			name:    "single quotes",
			content: "FOO='bar baz'\n",
			wantKey: "FOO",
			wantVal: "bar baz",
		},
		{
			name:    "commentaires et lignes vides",
			content: "# commentaire\n\nFOO=bar\n  \n# autre\n",
			wantKey: "FOO",
			wantVal: "bar",
		},
		{
			name:    "préfixe export",
			content: "export FOO=bar\n",
			wantKey: "FOO",
			wantVal: "bar",
		},
		{
			name:    "valeur contenant un =",
			content: "FOO=bar=baz\n",
			wantKey: "FOO",
			wantVal: "bar=baz",
		},
		{
			name:    "espace autour du =",
			content: "FOO = bar \n",
			wantKey: "FOO",
			wantVal: "bar",
		},
		{
			name:    "env existant prioritaire",
			content: "FOO=du_fichier\n",
			wantKey: "FOO",
			preSet:  map[string]string{"FOO": "du_shell"},
			wantVal: "du_shell",
		},
		{
			name:    "ligne invalide",
			content: "FOO\n",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Exécuter chaque sous-test dans un process séparé : loadDotEnv
			// ne redéfinit jamais une variable déjà présente, donc les
			// variables définies par un cas pollueraient les suivants
			// dans le même process (c'est le but du loader, pas un bug).
			if os.Getenv("JEV_TEST_SUBPROCESS") != "1" {
				cmd := exec.Command(os.Args[0], append(
					[]string{"-test.run=^TestLoadDotEnv$/" + tc.name},
					os.Args[1:]...)...)
				cmd.Env = append(os.Environ(), "JEV_TEST_SUBPROCESS=1")
				out, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("sous-process échoué : %v\n%s", err, out)
				}
				return
			}

			dir := t.TempDir()
			path := filepath.Join(dir, ".env")
			if err := os.WriteFile(path, []byte(tc.content), 0o600); err != nil {
				t.Fatal(err)
			}

			for k, v := range tc.preSet {
				os.Setenv(k, v)
			}

			err := loadDotEnv(path)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("attendu une erreur, obtenu nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("erreur inattendue : %v", err)
			}

			if got := os.Getenv(tc.wantKey); got != tc.wantVal {
				t.Errorf("os.Getenv(%q) = %q, attendu %q", tc.wantKey, got, tc.wantVal)
			}
		})
	}
}

func TestLoadDotEnvFichierAbsent(t *testing.T) {
	err := loadDotEnv(filepath.Join(t.TempDir(), "inexistant.env"))
	if err == nil {
		t.Fatal("attendu une erreur pour fichier absent")
	}
}

func TestLoadDotEnvValeurVide(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	content := "FOO=\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := loadDotEnv(path); err != nil {
		t.Fatalf("erreur inattendue : %v", err)
	}
	// Une CLE= vide définit FOO="" — l'environnement la contient mais vide.
	if _, ok := os.LookupEnv("FOO"); !ok {
		t.Error("FOO devrait exister (valeur vide) après chargement")
	}
}
