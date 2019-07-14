
<?php
require_once 'common.php';

$strict = in_array('--strict', $_SERVER['argv']);
$arguments = new \cli\Arguments(compact('strict'));

$arguments->addFlag(array('help', 'h'), 'Diese Hilfe');
$arguments->addFlag(array('notfancy', 'n'), 'Tabellen als einfache Listen ausgeben ohne Kopfzeile');
$arguments->addFlag(array('meldungen', 'm'), 'Aktuelle Meldungen ausgeben');

$arguments->addOption(array('info', 'i'), array(
	'default'     => '',
	'description' => 'Zeige Informationen zu einer Haltestelle')
);
$arguments->addOption(array('start', 's'), array(
	'default'     => '',
	'description' => 'Starthaltestelle für Route, zeigt Abfahrtsmonitor für diese Haltestelle an wenn kein Ziel angegeben wurde.')
);
$arguments->addOption(array('ziel', 'z'), array(
	'default'     => '',
	'description' => 'Zielhaltestelle für Route')
);

$arguments->parse();

$efa = new \efa($arguments);
$efa->run();
