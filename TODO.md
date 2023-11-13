# Les choses à faire

## Minitel

* [x] Tout passer en string/rune plutôt que byte
* [x] Gestion complète des accents
* [ ] Faire une pile d'Ack, il peut y en avoir plusieurs
* [ ] Améliorer la gestion des attributs par zone `0x20`
* [x] Wrapping automatique 40 chars
* [ ] Implémenter Hline et Vline
* [ ] Formattage plus intelligent, un objet Zone ? avec les attributs ?
* [ ] Page suite/retour en mode rouleau
* [ ] Toujours plus de loggggs !
* [ ] Page de tests du formattage
* [ ] Simplifier le fonctionnement de la boucle modem, trop de choses, trop de go-routines
* [ ] Réduire le nombre de goroutines dans la gestion d'une connexion modem
  * Il faut que la boucle listen soit synchrone avec le handler modem
  * De fait, on a une boucle applicative uniquement en goroutine

## Grafana

* [x] Labels: source=ws/modem1/modem2, etc...
* [x] Erreurs de connexion, attempts, lost

## Notel

* [ ] Réorganiser le code source pour avoir des packages
* [ ] Histo des départements les plus demandés sur la MTO
* [x] Une seule base de données utilisateurs
* [ ] Un fichier pour changer le message d'accueil, plutôt qu'en dur
* [ ] Refaire la page serveur avec des stats issues de prometheus
* [ ] Ban bad logins, when too many unsucessful
* [ ] Account management page

## Chat

* [ ] Afficher le nombre de connectées en rangée 0

## Actualités

* [ ] Trouver une solution à l'espace nécessaire pour le souslignage
* [x] Format basique avec la pagination

## PiouPiou

* [x] Limiter les premières fonctionalités (pas de profil, pas de notif ?)
* [ ] Ecrire la fonction des erreurs standard
