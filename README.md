#Command Line Tool für die Elektronische Fahrplanauskunft

##Requirements:

PHP 5.3 

##Installation

```
wget https://github.com/simonjenny/efa-cli/blob/master/efa.phar?raw=true
mv efa.phar > efa
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
