# Workflow API — Vue d'ensemble du projet

Backend Go de **Workflow**, une plateforme de construction et d'exécution de workflows HTTP visuels. Les équipes définissent des appels API réutilisables, les composent en graphes, les paramètrent avec des variables et des assertions, puis les exécutent manuellement ou selon un planning — avec collaboration temps réel et facturation SaaS.

Ce dépôt est la **source de vérité** : persistance, authentification, orchestration d'exécution, quotas, événements et API REST.

---

## Proposition de valeur

| Besoin | Réponse produit |
|--------|-----------------|
| Tester / enchaîner des APIs sans code | Builder visuel : steps + connexions |
| Réutiliser des appels HTTP | Bibliothèque d'**endpoints** partagés |
| Chaîner des réponses | **Variables** extraites du body / headers |
| Valider les réponses | **Assertions** sur status, headers, body |
| Automatiser | **Planification** (cron-like) + exécution async |
| Collaborer | Multi-utilisateurs par **projet**, sync **Centrifugo** |
| Monétiser | **Abonnements Stripe** + **quotas** par plan |

---

## Modèle métier

### Hiérarchie multi-tenant

```
User
 └── Project (espace de travail collaboratif)
      ├── Endpoints (bibliothèque HTTP)
      ├── Workflows (graphes d'exécution)
      │    ├── Steps (nœuds = snapshot d'un endpoint)
      │    ├── Connections (arêtes dirigées entre steps)
      │    ├── Variables (statiques ou extraites)
      │    └── Assertions (règles de validation par step)
      └── WorkflowRuns (exécutions)
           └── StepRuns (exécution unitaire d'un step)
                └── Insights (métriques réseau / timing)
```

Toutes les ressources métier sont **scopées au projet actif** de l'utilisateur. Un accès à une ressource d'un autre projet membre renvoie `409 WRONG_ORGANIZATION` pour permettre au client de basculer de projet.

> **Note terminologie** : le domaine utilise `Project` (anciennement « organization » dans certains docs). L'API expose `/api/projects`.

### Concepts clés

#### Project
Espace collaboratif : nom, membres, soft delete. Chaque utilisateur a un **projet actif** (`activeProject`).

#### Endpoint
Template HTTP réutilisable au niveau projet :
- method, URL, headers, query, body
- timeout, retry (count, delay, onFailure)
- import possible depuis **OpenAPI**

Un endpoint est une définition ; il n'est pas exécuté directement.

#### Workflow
Graphe nommé, rattaché à un projet :
- statut `active` / `inactive` (seuls les actifs peuvent être planifiés / démarrés)
- **planification** : `none`, `recurring` (intervalle + unité), `once` (date/heure + timezone)
- **concurrence** : nombre max de runs simultanés par workflow
- **notifications** (flags succès / échec / annulation — côté domaine, handlers à venir ou externes)

#### Step
Instance d'un endpoint sur le canvas d'un workflow :
- snapshot de la config HTTP au moment de la création / mise à jour
- **position** canvas (`x`, `y`) → recalcul de `index`, `executionOrder`, `treeIndex`
- ordre d'exécution dérivé du graphe + position

#### Connection
Arête dirigée `sourceStepId` → `targetStepId`. Définit le parcours et les branches (steps parallèles ou conditionnels via skip de branches).

#### Variable
Donnée disponible pendant l'exécution d'un workflow :
- **`static`** : valeur fixe définie à la conception
- **`extracted`** : valeur extraite d'une réponse HTTP précédente (JSON path sur body, ou header)

Les variables alimentent l'interpolation des URLs, headers, body des steps suivants.

#### Assertion
Règle de validation attachée à un step, évaluée **après** une réponse HTTP 2xx :
- **source** : `status`, `header`, `body`
- **opérateurs** : equals, contains, greater_than, is_null, matches_regex, is_string, etc.
- échec → retry du step puis `failed` si épuisé

#### WorkflowRun
Une exécution complète d'un workflow :
- **statuts** : `pending` → `running` → `success` | `failed` | `cancelled`
- **déclenchement** : `user`, `schedule`, `webhook`, `api`, `cli`
- **contexte** JSON : variables fusionnées au fil de l'exécution

#### StepRun
Exécution d'un step dans le cadre d'un run :
- snapshot figé (config HTTP + assertions) au moment de l'enqueue
- **statuts** : `queued` → `running` → `success` | `failed` | `skipped` | `cancelled`
- résultats HTTP : status, headers, body, erreur, tentatives

#### Insight
Métriques d'une tentative HTTP (plan Pro+) :
- timings : DNS, TCP, TLS, TTFB, durée totale, queue time
- tailles requête / réponse, code status, message d'erreur

---

## Fonctionnalités par domaine

### Builder & configuration

| Feature | Description |
|---------|-------------|
| CRUD Workflows | Création, édition, activation/désactivation, suppression |
| CRUD Endpoints | + import OpenAPI |
| CRUD Steps | Position canvas, lien vers endpoint |
| CRUD Connections | Graphe dirigé |
| CRUD Variables | Statiques et extraites, recherche de paths JSON |
| CRUD Assertions | Par step, recherche de paths pour le body |
| Recalcul d'ordre | Synchrone à la création / déplacement de step ou connexion |

### Exécution

| Feature | Description |
|---------|-------------|
| Démarrage manuel | `POST /workflows/:id/start` |
| Arrêt | `POST /workflows/:id/stop` (annule le run en cours) |
| Planification | Scheduler binaire : claim des workflows `due`, start automatique |
| Orchestration | Worker : graphe → enqueue step runs racine → avance après succès/échec |
| Exécution HTTP | Executor : consomme `stepRun.queued`, appelle l'API cible, assertions, retries |
| Skip de branches | Steps non atteignables marqués `skipped` |
| Finalisation | Run `success` si tous les steps atteignables OK, `failed` si au moins un échec |

### Observabilité & historique

| Feature | Description |
|---------|-------------|
| Liste des runs | Vue allégée (id, status, dates, stepRuns status) |
| Détail d'un run | Step runs complets, assertions résultats, insights |
| Analytics | Agrégats par workflow (succès, échecs, durées) |

### Collaboration temps réel

Événements **Centrifugo** sur le channel `user:{userId}` :
- entités builder : `workflow.*`, `endpoint.*`, `step.*`, `connection.*`, `variable.*`, `assertion.*`
- exécution : `workflowRun.started|succeeded|failed|cancelled|finished`, `stepRun.started|succeeded|failed`

Le type WS est `entity.action` (sans version). Les events domaine bus sont versionnés `*.v1`.

`workflowRun.finished` est l'event global de fin avec `finishType` : `success` | `failed` | `cancelled`.

### Identité & accès

| Feature | Description |
|---------|-------------|
| Auth | Clerk JWT sur `/api/*` |
| Webhook utilisateurs | `POST /webhooks/clerk` (Svix) |
| Profil | `GET /users/me`, changement de projet actif |

### Facturation & quotas

| Feature | Description |
|---------|-------------|
| Plans | Free, Starter, Pro, Enterprise (quotas associés) |
| Abonnement | Stripe : création, preview changement, portail facturation |
| Webhook billing | `POST /webhooks/stripe` |
| Quotas | Membres, projets, workflows, steps, endpoints, variables, runs/mois, runs concurrents, intervalle min de schedule, rétention historique, tailles body, insights, import OpenAPI, etc. |
| Factures | `GET /invoices` |

---

## Flux d'exécution (simplifié)

```
[API / Scheduler / CLI]
        │
        ▼
  StartWorkflowRun (commande)
        │  transaction: WorkflowRun + outbox
        ▼
  workflowRun.started.v1
        │
        ▼
  Worker — Orchestrator.OnStarted
        │  MarkRunning, crée StepRuns racine
        ▼
  stepRun.queued.v1 (× N)
        │
        ▼
  Executor — ExecuteHandler
        │  HTTP call, assertions, insight
        │  stepRun.succeeded.v1 | failed.v1
        ▼
  Worker — Orchestrator.OnSucceeded / OnFailed
        │  merge variables, enqueue suivants, skip branches
        │  finalizeRun → workflowRun.succeeded|failed.v1 + finished.v1
        ▼
  Worker — PublishRealtime → Centrifugo
```

**Annulation** : commande `CancelWorkflowRun` → run + step runs non terminaux en `cancelled` → events + `finished`.

---

## Architecture technique

| Couche | Rôle |
|--------|------|
| `interfaces/http` | Handlers Fiber, DTOs, validation, presenters |
| `application/command` | Écritures synchrones (HTTP) |
| `application/query` | Lectures (vues SQL optimisées) |
| `application/event` | Handlers async (orchestration, realtime, billing) |
| `domain` | Agrégats, events, règles métier — sans dépendance infra |
| `infrastructure` | Postgres/GORM, RabbitMQ, Clerk, Centrifugo, Stripe |
| `cmd/*/di` | Composition / câblage |

**CQRS** : commandes ≠ events async. Outbox transactionnelle → worker relay → RabbitMQ → handlers idempotents (`processed_events`).

Conventions détaillées : [`.cursor/rules/architecture.mdc`](../.cursor/rules/architecture.mdc).

---

## Binaires

| Binaire | Rôle |
|---------|------|
| `cmd/api` | Serveur HTTP REST + webhooks |
| `cmd/worker` | Relay outbox + consumer RabbitMQ principal (orchestration, realtime, billing events) |
| `cmd/executor` | Consumer dédié exécution HTTP des step runs |
| `cmd/scheduler` | Tick périodique : démarre les workflows planifiés |
| `cmd/cli` | Migrations Goose + vérification drift schéma / commandes admin |

---

## Stack

- **Go 1.25**, Fiber v3, PostgreSQL, GORM (writes), Goose
- **RabbitMQ** (topic, retry TTL + DLQ)
- **Centrifugo** (WebSocket)
- **Clerk** (auth), **Stripe** (billing)
- **Docker Compose** pour le dev local

---

## Surface API (résumé)

Préfixe `/api`, auth Bearer sauf `/plans` et webhooks.

- **Users** : profil, projet actif
- **Projects** : CRUD, membres, activation
- **Workflows** : CRUD, activate/deactivate
- **Endpoints** : CRUD, import OpenAPI
- **Steps / Connections** : nested sous workflow
- **Variables / Assertions** : nested sous workflow (et step)
- **Workflow runs** : start, stop, list, detail, analytics
- **Subscription / Quota / Invoices** : billing
- **Realtime** : token de connexion Centrifugo

Health : `GET /livez`, `/readyz`, `/startupz`.

---

## Événements domaine (principaux)

| Agrégat | Events |
|---------|--------|
| User | `user.created.v1`, `user.updated.v1` |
| Project | `project.created.v1`, `updated.v1`, `deleted.v1`, `member_added.v1`, … |
| Workflow | `workflow.created.v1`, `updated.v1`, `deleted.v1`, `activated.v1`, … |
| Endpoint | `endpoint.created.v1`, `updated.v1`, `deleted.v1` |
| Step | `step.created.v1`, `updated.v1`, `deleted.v1`, `position_updated.v1` |
| Connection | `connection.created.v1`, `deleted.v1` |
| Variable | `variable.created.v1`, `updated.v1`, `deleted.v1` |
| Assertion | `assertion.created.v1`, `updated.v1`, `deleted.v1` |
| WorkflowRun | `started`, `succeeded`, `failed`, `cancelled`, **`finished`**, `scheduledSkipped` |
| StepRun | `queued`, `started`, `succeeded`, `failed`, `cancelled` |
| Subscription / Invoice | events Stripe synchronisés |

---

## Périmètre hors repo

- **UI builder** (canvas React) : consomme cette API + Centrifugo
- **workflow-executor** peut être documenté séparément ; l'exécution HTTP vit ici dans `cmd/executor`

---

## Démarrage rapide

```bash
cp .env.dist .env   # Clerk, Postgres, RabbitMQ, Centrifugo, Stripe
make dev
make migrate
```

API : `http://localhost:4000` — voir [README.md](../README.md) pour ports et commandes Makefile.
