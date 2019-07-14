
<?php
require_once 'common.php';

$strict = in_array('--strict', $_SERVER['argv']);
$arguments = new \cli\Arguments(compact('strict'));

$arguments->addFlag(array('help', 'h'), 'Diese Hilfe');
$arguments->addFlag(array('nofancy', 'n'), 'Tabellen als einfache Listen ausgeben ohne Kopfzeile');
$arguments->addFlag(array('route','r'), 'Zeigt eine Reiseroute von --start nach --ziel');
$arguments->addFlag(array('meldungen'), 'Aktuelle Meldungen ausgeben');

$arguments->addOption(array('monitor', 'm'), array(
	'default'     => '',
	'description' => 'Zeige alle Abfahrten ab dieser Haltestelle')
);
$arguments->addOption(array('start', 's'), array(
	'default'     => '',
	'description' => 'Starthaltestelle für --route')
);
$arguments->addOption(array('ziel', 'z'), array(
	'default'     => '',
	'description' => 'Zielhaltestelle für --route')
);

$arguments->parse();

$efa = new \efa($arguments);
$efa->run();
