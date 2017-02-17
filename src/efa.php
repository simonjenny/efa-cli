<?php

class Efa
{

  public function __construct($arguments){
    $this->arguments = $arguments;
  }

  public function run(){
    if ($this->arguments['help']) {
       echo $this->arguments->getHelpScreen().PHP_EOL.PHP_EOL;
       die();
    }
    foreach($this->arguments->getArguments() as $key => $value){
      if(method_exists($this, $key)){
        call_user_func(array($this, $key));
        break;
      }
    }
  }

  public function route(){

    if($this->arguments['start'] == false ||$this->arguments['ziel'] == false){
      \cli\err('Fehler! Die Parameter für die Start und/oder Zielhaltestelle fehlen!');
      echo $this->arguments->getHelpScreen().PHP_EOL.PHP_EOL;
      return false;
    }

    if(!is_numeric($this->arguments['start'])){
      $this->getStationList('route','start');
      return false;
    }
    if(!is_numeric($this->arguments['ziel'])){
      $this->getStationList('route','ziel');
      return false;
    }

    $data = json_decode(file_get_contents(sprintf("http://www.efa-bw.de/bvb/XML_TRIP_REQUEST2?itdDate=%s&itdTime=%s&language=de&sessionID=0&outputFormat=JSON&type_origin=stop&name_origin=%s&type_destination=stop&name_destination=%s", date('Ymd') , date('Hi'), $this->arguments['start'], $this->arguments['ziel'])),false);
    \cli\line("Verfügbare Routen von %s nach %s".PHP_EOL, $data->origin->points->point->name, $data->destination->points->point->name);
    $trips = [];
    $i = 0;
    foreach($data->trips as $trip){
      $trips["t-".$i++] = sprintf("Ab %s um %s mit %sx Umsteigen. Dauer %sh",
                              $trip->legs[0]->points[0]->name,
                              $trip->legs[0]->points[0]->dateTime->time,
                              $trip->interchange,
                              $trip->duration
                            );
    }
    $trips['quit'] = 'Abbrechen';
    while (true) {
      $choice = \cli\menu($trips, null, 'Bitte wähle eine Fahrt');
      \cli\line();
      if ($choice == 'quit') {
        break;
      } else {
         $tabledata = [];
         foreach($data->trips[str_replace('t-','',$choice)]->legs as $legs){
           $tabledata[] = [
             (empty($legs->mode->name)) ? "Fussweg" : $legs->mode->name,
             $legs->points[0]->name,
             $legs->points[0]->dateTime->time."h",
             $legs->points[1]->dateTime->time."h",
             $legs->points[1]->name
           ];
         }
         if($this->arguments['nofancy']){
           foreach($tabledata as $line){
             \cli\line("{:0} {:1} {:2} {:3} {:4}", $line);
           }
         } else {
           $table = new \cli\Table();
           $table->setHeaders(['', 'Einsteigen', 'Abfahrt', 'Ankunft', 'Aussteigen/Umsteigen']);
           $table->setRows($tabledata);
           $table->display();
         }
         break;

      }
    }

  }

  public function meldungen(){
    $data = json_decode(file_get_contents("http://www.efa-bw.de/bvb/XML_ADDINFO_REQUEST?filterProviderCode=Basler%20Verkehrs-Betriebe%20(BVB)&outputFormat=JSON"), false);
    if($data->additionalInformation->travelInformations->travelInformation){
      foreach($data->additionalInformation->travelInformations->travelInformation as $info){
        \cli\line(\cli\Colors::colorize("%r".$info->infoLink->infoLinkText."%n"), true);
        $text = strip_tags(html_entity_decode($info->infoLink->content), '<br>');
        $text = str_replace('<br /><br />',PHP_EOL, $text);
        $text = str_replace('<br />',PHP_EOL, $text);
        \cli\line($text);
      }
    } else {
      \cli\line(\cli\Colors::colorize("%g✓ Aktuell sind keine Störungen bekannt.%n", true));
    }

  }

  public function monitor(){

    if($this->arguments['monitor'] == false){
      \cli\err('Fehler! Der Parameter für die Haltestelle fehlt!');
      echo $this->arguments->getHelpScreen().PHP_EOL.PHP_EOL;
      return false;
    }

    if(is_numeric($this->arguments['monitor'])){
      $data   = [];
      $headers= ['Min', '', '#', 'Nach'];
      $departures = json_decode(file_get_contents(sprintf('http://www.efa-bw.de/bvb/XML_DM_REQUEST?laguage=de&typeInfo_dm=stopID&deleteAssignedStops_dm=1&useRealtime=1&mode=direct&excludedMeans=0&excludedMeans=1&excludedMeans=2&limit=10&outputFormat=JSON&nameInfo_dm=%s', urlencode($this->arguments['monitor']))), false);
      foreach($departures->departureList as $departure){
        $data[] = [
          $this->padd($departure->countdown),
          $departure->servingLine->name,
          $departure->servingLine->number,
          $departure->servingLine->direction
        ];
      }
      if($this->arguments['nofancy']){
        foreach($data as $line){
          \cli\line("{:0} {:1} {:2} {:3}", $line);
        }
      } else {
        $table = new \cli\Table();
        $table->setHeaders($headers);
        $table->setRows($data);
        $table->display();
      }
    } else {
      $this->getStationList('monitor','monitor');
    }
  }

  public function getStationList($callback, $target){
    $stations = json_decode(file_get_contents(sprintf('http://www.efa-bw.de/bvb/XSLT_STOPFINDER_REQUEST?language=de&outputFormat=JSON&coordOutputFormat=WGS84[DD.ddddd]&itdLPxx_usage=origin&useLocalityMainStop=true&SpEncId=0&locationServerActive=1&stateless=1&type_sf=any&anyObjFilter_sf=2&anyMaxSizeHitList=50&name_sf=%s', urlencode($this->arguments[$target]))), false);
    if(count($stations->stopFinder->points) == 1){
      $this->arguments[$target] = $stations->stopFinder->points->point->stateless;
      call_user_func(array($this, $callback));
    } else {
      $menu = [];
      foreach($stations->stopFinder->points as $point){
        $menu[$point->stateless] = $point->name;
      }
      $menu['quit'] = 'Abbrechen';
      \cli\line("Parameter --%s hat mehrere Haltestellenmöglichkeiten:".PHP_EOL, $target);
      while (true) {
      	$choice = \cli\menu($menu, null, 'Bitte wähle eine Haltestelle');
      	\cli\line();
      	if ($choice == 'quit') {
      		break;
      	} else {
          $this->arguments[$target] = $choice;
          call_user_func(array($this, $callback));
          break;
        }
      }
    }
  }

  private function padd($t){
   return ($t < 10) ? " ".$t."'" : $t."'";
  }

}
