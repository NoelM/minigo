# Changelog

## v1.0.0

### Fonctionnalités

- Affichage d'un texte depuis un fichier

### Corrections

- Suppression des fonctions `AppendKey` et `Correction` dans `info_page` puisqu'inutiles ; aucune entrée n'est attendue.
- Ajout d'une vérification dans `Form` pour ne pas accéder les entrées lorsqu'il n'y a pas d'`input` déclaré. 
- Problèmes d'affichage sur la page serveur, inversion position ligne/colonne.