# jevtest

Petit banc d'essai du modèle de décision [TypeSafe Jev](https://openrouter.ai/typesafe) (via l'API Decisions d'OpenRouter) sur la détection d'ironie / second degré en français.

L'app envoie une batterie de cas délicats à Jev et compare les réponses structurées (probabilités, confiance) aux attentes :

- cas miroirs : mêmes marqueurs de surface (« bravo », « chapeau », hyperboles), intentions opposées ;
- litote sans marqueur d'exagération ;
- hyperbole sincère (« je pleure de rire ») ;
- ironie par contraste factuel ;
- un cas ultra-ambigu sans contexte (« Bravo, vraiment. ») pour lire le prior du modèle.

## Résultats renvoyés par cas

- `degree` (choice) : « Premier degré » / « Second degré », avec confiance et distribution par option.
  Verdict affiché : TRANCHE (confiance ≥ 0.75), DOUTE (≥ 0.5), TREMBLE sinon (défaut : 1er degré).
- `is_ironic` (noul) : probabilité que le texte soit ironique.

Le temps de réponse est affiché par cas, avec total et moyenne en fin de run.

## Prérequis

- Go 1.27+ 
- Une clé API [OpenRouter](https://openrouter.ai/keys)

## Installation

```bash
go build .
```

## Configuration

La clé API est lue dans la variable `OPENROUTER_API_KEY`, avec priorité à l'environnement du shell :

1. `OPENROUTER_API_KEY` déjà définie dans l'environnement → utilisée telle quelle ;
2. sinon, la valeur du fichier `.env` (s'il est présent dans le répertoire courant) est chargée.

Le fichier `.env` est ignoré par git. Format accepté : `CLE=valeur`, guillemets simples ou doubles optionnels, préfixe `export ` optionnel, commentaires `#`. Une variable déjà définie dans l'environnement n'est jamais écrasée.

```bash
cp /dev/null .env   # puis :
echo 'OPENROUTER_API_KEY=sk-or-v1-...' >> .env
```

## Utilisation

```bash
./jevtest
```

Exemple de sortie :

```
=== Cas 1 : ironie classique sur un reproche ===
Texte : "Ah oui, bravo, t'as encore oublié mes clés au bureau. Franchement, chapeau."
  CHOICE -> Second degré    (TRANCHE) | Premier degré: 0%  Second degré: 100%  
  NOUL   -> probabilité ironique: 97%
  Attendu: Second degré | Temps: 447.210103ms
```

## Détails techniques

- SDK : SDK Go officiel OpenRouter (`github.com/OpenRouterTeam/go-sdk`), endpoint `Alpha.Decisions.Create` (`https://openrouter.ai/api/alpha/decisions`).
- Modèle : `typesafe/jev-1.13`.
- Les questions sont typées via `components.Questions` (`noul` et `choice`) et les critères passés sous forme de map / struct dédiée.
- Headers `HTTP-Referer` et `X-Title` configurés via `WithHTTPReferer` et `WithXTitle` pour l'attribution OpenRouter.
- Timeout HTTP : 30 s par requête (`WithTimeout`).

## Structure

```
main.go        client OpenRouter SDK (Alpha Decisions) + batterie de tests
dotenv.go      mini-loader .env (zéro dépendance)
dotenv_test.go tests du loader
```

## Tests

```bash
go test ./...
```