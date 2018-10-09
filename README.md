# Command Line Tool für die Elektronische Fahrplanauskunft

Abfahrtsmonitor:
```
$ efa --monitor Basel, Mparc

+-----+----------+----+------------------------+
| Min |          | #  | Nach                   |
+-----+----------+----+------------------------+
|  1' | Tram BLT | 11 | Basel, St-Louis Grenze |
|  1' | Tram BLT | 11 | Aesch BL, Dorf         |
|  4' | Tram BLT | 10 | Dornach, Bahnhof       |
|  6' | Tram BLT | 10 | Ettingen, Bahnhof      |
|  8' | Tram BLT | 11 | Aesch BL, Dorf         |
|  9' | Tram BLT | 11 | Basel, St-Louis Grenze |
| 11' | Tram BLT | 10 | Dornach, Bahnhof       |
| 14' | Tram BLT | 10 | Rodersdorf, Station    |
| 16' | Tram BLT | 11 | Aesch BL, Dorf         |
| 16' | Tram BLT | 11 | Basel, St-Louis Grenze |
+-----+----------+----+------------------------+
```

Routenplaner:
```
$ efa --route --start Basel, MParc --ziel Allschwil

Parameter --ziel hat mehrere Haltestellenmöglichkeiten:

  1. Allschwil, Dorf
  2. Allschwil, Binningerstrasse
  3. Allschwil, Friedhof
  4. Allschwil, Gartenhof
  5. Allschwil, Gartenstrasse
  6. Allschwil, Grabenring
  7. Allschwil, Hagmattstrasse
  8. Allschwil, Im Brühl
  9. Allschwil, Kirche
  10. Allschwil, Kreuzstrasse
  11. Allschwil, Lindenplatz
  12. Allschwil, Merkurstrasse
  13. Allschwil, Reservoir
  14. Allschwil, Rosenberg
  15. Allschwil, Stegmühleweg
  16. Allschwil, Ziegelei
  17. Allschwil, Ziegelhof
  18. Allschwil, Parkallee
  19. Allschwil, Bettenacker
  20. Allschwil, Letten
  21. Binningen BL, Allschwilerweg
  22. Allschwil, Paradies
  23. Allschwil, Spitzwald
  24. Allschwil, Zum Sporn
  25. Basel, Allschwilerplatz
  26. Basel, Depot Allschwilerstrasse
  27. Abbrechen

Bitte wähle eine Haltestelle: 1

Verfügbare Routen von Basel, MParc nach Allschwil, Dorf

  1. Ab Basel, MParc um 13:14 mit 1x Umsteigen. Dauer 00:28h
  2. Ab Basel, MParc um 13:21 mit 1x Umsteigen. Dauer 00:28h
  3. Ab Basel, MParc um 13:24 mit 2x Umsteigen. Dauer 00:25h
  4. Ab Basel, MParc um 13:29 mit 1x Umsteigen. Dauer 00:28h
  5. Abbrechen

Bitte wähle eine Fahrt: 1

+-------------+-----------------+---------+---------+----------------------+
|             | Einsteigen      | Abfahrt | Ankunft | Aussteigen/Umsteigen |
+-------------+-----------------+---------+---------+----------------------+
| Tram BLT 10 | Basel, MParc    | 13:14h  | 13:26h  | Basel, Heuwaage      |
| Tram BVB 6  | Basel, Heuwaage | 13:27h  | 13:42h  | Allschwil, Dorf      |
+-------------+-----------------+---------+---------+----------------------+
```

Störungsmeldungen:
```
$ efa --meldungen

L 14, 36, 37, 47: Fussballspiel, 19.02.2017,  FC Basel - FC Lausanne-Sport

L 14:
Es verkehren zusätzliche Trams zwischen Aeschenplatz und und St. Jakob.

L 36:
Wird vor Spielende ab ca. 15:15 ca. 17:30
zwischen Breite und Dreispitz wie folgt umgeleitet:
-> Kleinhüningen - Breite - Sevogelplatz*  - Dreispitz - Schifflände.
-> Schifflände - Dreispitz - Aeschenplatz* - Breite - Kleinhüningen.

*ab hier jeweils Umsteigemöglichkeit auf die L 14.

Es verkehren Einsatzbusse zwischen Stadionstrasse und Breite, mit Halt an Redingstrasse, Forellenweg und Nasenweg.
Die Haltestellen rund um das Stadion sind teilweise verschoben. Beachten Sie bitte die Hinweise vor Ort.

L 37, 47:
Bei einer Teil- oder Vollsperrung des Bereichs St. Jakob werden die L 37 und 47 umgeleitet, resp. am St. Jakob gewendet.
Bitte beachten Sie die Durchsagen des BLT-Fahrpersonals.

Beachten Sie bitte die Hinweise vor Ort.

Bemerkungen:
Spielbeginn: 13:45
Das Matchticket ist zur freien Fahrt zum Stadion und zurück gültig.

Es kann zu Abweichungen zum Online-Fahrplan kommen.
```

## Requirements:

PHP 5.3 

##Installation

```
wget https://github.com/simonjenny/efa-cli/blob/master/efa.phar?raw=true -O efa 
chmod a+x efa
```


```
$ efa --help

Flags
  --help, -h     Diese Hilfe
  --nofancy, -n  Tabellen als einfache Listen ausgeben ohne Kopfzeile
  --route, -r    Zeigt eine Reiseroute von --start nach --ziel
  --meldungen    Aktuelle Meldungen ausgeben

Options
  --monitor, -m  Zeige alle Abfahrten ab dieser Haltestelle
  --start, -s    Starthaltestelle für --route
  --ziel, -z     Zielhaltestelle für --route
  
```
