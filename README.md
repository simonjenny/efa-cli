# Command Line Tool für die Elektronische Fahrplanauskunft


```
$ efa --help

Flags
  --help, -h       Diese Hilfe
  --notfancy, -n   Tabellen als einfache Listen ausgeben ohne Kopfzeile
  --meldungen, -m  Aktuelle Meldungen ausgeben

Options
  --info, -i   Zeige Informationen zu einer Haltestelle
  --start, -s  Starthaltestelle für Route, zeigt Abfahrtsmonitor für diese
               Haltestelle an wenn kein Ziel angegeben wurde.
  --ziel, -z   Zielhaltestelle für Route

```

Abfahrtsmonitor:
```
$ efa --start Basel, Mparc

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
$ efa --start Basel, MParc --ziel Allschwil

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

  1. L 19: Bauarbeiten Bahnhof Liestal, 29.06. -11.08.2019
  2. L 31, 38, 42: Haltestelle Hoffmann-La Roche nicht bedient
  3. L 31, 38, 42: Bauarbeiten Grenzacherstrasse, 15.07. - 25.07.2019
  4. L 8: Bauarbeiten Kleinhüningen, 15.07. - 16.07.2019
  5. L 6: Bauarbeiten Morgartenring - Allschwil Dorf, 20.05 - 25.08.2019
  6. L 1, 2, 15: Bauarbeiten Centralbahnplatz und St.Alban-Graben, 15.07. - ca. 22.07.2019
  7. L 36: Haltestelle Signalstrasse noch nicht bedient
  8. L 31, 38, 48, 64, 608: Bauarbeiten Bachgraben/Hegenheimermattweg, von 14.01. - ca. 31.10.2019
  9. Abbrechen

Bitte wähle eine Meldung: 8

 ✓ L 31, 38, 48, 64, 608: Bauarbeiten Bachgraben/Hegenheimermattweg, von 14.01. - ca. 31.10.2019

Von Mo, 14.01. bis ca. Ende Oktober 2019 finden im Bereich Bachgraben
und Hegenheimermattweg Bauarbeiten statt.
Dadurch kommt es zu den folgenden Änderungen in der Haltestellenbedienung:
L 31, 38, 48, 64, 608:
Die Haltestelle Bachgraben in Richtung Allschwil ist in die Lachenstrasse verschoben.
Die Haltestelle Kreuzstrasse in Richtung Basel ist zur Kreuzung
Kreuzstrasse/Hegenheimermattweg verschoben.
Dauer:Mo, 14.01.2019
bis voraussichtlich ca.
Ende Oktober
Grund:
Bauarbeiten
```

## Requirements

- min. PHP 5.3
- (https://github.com/humbug/box)Box (für PHAR)

## Installation

```
wget https://github.com/simonjenny/efa-cli/blob/master/efa.phar?raw=true -O efa
chmod a+x efa
```

## Build

1. Repository klonen
2. Composer
```
composer install
```
3. Box Build
```
box compile
```
