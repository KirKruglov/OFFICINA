<p align="center">
  <a href="README.md">English</a> ·
  <a href="README.de.md">Deutsch</a> ·
  <a href="README.es.md">Español</a> ·
  <strong>Français</strong> ·
  <a href="README.ru.md">Русский</a>
</p>

<h1 align="center">OFFICINA</h1>

<p align="center">
  <em>du latin</em> <strong>officina</strong> — un atelier, une forge où l'on fabrique les outils.
</p>

<p align="center">
  <img alt="Plateforme : macOS" src="https://img.shields.io/badge/platform-macOS-000000?logo=apple&logoColor=white">
  <img alt="Cœur : Claude Code" src="https://img.shields.io/badge/core-Claude%20Code-da7756">
  <img alt="Licence : MIT" src="https://img.shields.io/badge/license-MIT-green">
  <a href="https://github.com/KirKruglov/OFFICINA/commits/main"><img alt="Dernier commit" src="https://img.shields.io/github/last-commit/KirKruglov/OFFICINA?logo=github&label=dernier%20commit"></a>
</p>

<p align="center">
  <strong>Méthode et outils pour un travail systématique avec un assistant de développement IA.</strong>
  Skills, sous-agents, CLI et méthodologie — assemblés et éprouvés sur de vrais projets.
</p>

---

> <sub>Traduit du <a href="README.md">README anglais</a> de référence ; peut être en retard sur l'original.</sub>

## Pourquoi OFFICINA

OFFICINA est un atelier pour travailler avec des outils d'IA dans le développement de produits. En son
cœur, la **méthode** : comment structurer le travail avec un assistant IA pour qu'il reste prévisible
et reproductible. Autour d'elle — skills, sous-agents, CLI et configurations, prêts à réutiliser.

- **La méthode d'abord.** Des règles et guides réutilisables — style rédactionnel, création de CLI,
  création de skills et de sous-agents, ingénierie des boucles — chacun lisible en soi.
- **Skills portables, Claude Code au cœur.** Les skills ne sont pas liés à un seul outil — ils tournent
  dans Claude Code et des environnements compatibles. Les sous-agents et réglages sont ajustés pour
  Claude Code. Les deux sont sélectionnés à la main plutôt que déversés en bloc.
- **L'environnement autour du cœur.** VS Code comme section appliquée qui soutient le cœur au lieu d'en
  être la vitrine.

## Pour qui

- **Vous travaillez avec un assistant IA et voulez un système** — une méthode, pas des prompts ponctuels.
- **Vous écrivez vos propres skills, sous-agents, CLI** — vous prenez des schémas et guides éprouvés.
- **Vous configurez votre environnement de travail** pour le développement IA sur macOS.

Si vous cherchez une distribution plug-and-play à « installer et oublier » — ce n'est pas ça : OFFICINA
parle de méthode et de choix délibéré, pas d'un assemblage clé en main prêt à l'emploi.

## Ce qu'il contient

| Section | Ce qu'elle contient |
|---|---|
| [`skills/`](skills/) | Skills portables — modes et procédures réutilisables (Claude Code et environnements compatibles) |
| [`claude/`](claude/) | Couche Claude Code — sous-agents et réglages |
| [`methodology/`](methodology/) | Règles et guides — *comment travailler* |
| [`cli/`](cli/) | Outils CLI personnels et bibliothèques partagées |
| [`vscode/`](vscode/) | Réglages VS Code, raccourcis clavier, extensions curatées |
| [`install/`](install/) | Scripts d'installation et d'agencement |

## Comment fonctionne la méthode

Le point de départ est une question — avec quoi résoudre la tâche :

- Une action déterministe, sans modèle — **CLI**
- Comprendre le contexte et décider selon la situation — **skill**
- Une sous-tâche isolée ou un rôle-persona — **sous-agent**
- Un cycle récurrent planifié — **loop**

Vient ensuite le guide correspondant dans [`methodology/`](methodology/) : structure, conventions, une
checklist avant le déploiement. La méthode est commune ; l'artefact se transpose d'un projet à l'autre.

## Démarrage rapide

Nécessite **macOS**. Clonez le dépôt et ouvrez la section qui correspond à votre besoin — chacune livre
sa propre README et, le cas échéant, un installateur en une commande sous [`install/`](install/) :

```bash
git clone https://github.com/KirKruglov/OFFICINA.git
cd OFFICINA
```

Le premier installateur prêt met en place l'environnement VS Code :

```bash
./install/vscode.sh   # réglages, raccourcis, police MesloLGS Nerd Font, extensions curatées
```

## Philosophie

- **Le bon outil pour la tâche.** Le déterministe va dans un script, les décisions selon le contexte au
  modèle. Un modèle superflu dans la boucle ajoute coût et incertitude.
- **La fiabilité repose sur les contraintes.** Limites, budgets et vérifications tiennent le système.
  Une boucle autonome a besoin de quelque chose capable de dire « non ».
- **Le savoir vit dans l'artefact.** Un guide, un skill ou un sous-agent se réutilise ; ce qui est
  établi une fois ne l'est pas de nouveau à chaque fois.
- **Sélectionné à la main.** Ce qui entre dans le dépôt est éprouvé en pratique et choisi délibérément.

## Contribuer

Les issues, questions et petites corrections sont les bienvenues — voir [CONTRIBUTING.md](CONTRIBUTING.md).
OFFICINA est un projet ouvert.

Utile ? Mettez une étoile — c'est ainsi que le prochain développeur trouve le projet. Forkez-le, prenez
ce qui vous convient et adaptez-le à votre configuration.

## Licence

Publié sous [licence MIT](LICENSE) — © 2026 Kir Kruglov. Libre d'utilisation, de modification et de
distribution.
