<?php

namespace App\Helpers;

use Illuminate\Support\Facades\Http;

use function Laravel\Prompts\spin;

class Efa
{
    private static function load($url)
    {

        $response = config('json') ? Http::accept('application/json')->get('https://www.efa-bw.de/bvb3/'.$url) :
        spin(
            message: 'Fetching Data...',
            callback: fn () => Http::accept('application/json')->get('https://www.efa-bw.de/bvb3/'.$url)
        );

        return $response->object() ?? [];

    }

    public static function meldungen(): ?array
    {
        return self::load('XML_ADDINFO_REQUEST?filterProviderCode=Basler%20Verkehrs-Betriebe%20(BVB)&outputFormat=JSON')->additionalInformation->travelInformations->travelInformation ?? [];
    }

    public static function haltestelle($haltestelle): ?object
    {
        return self::load('XSLT_STOPFINDER_REQUEST?language=de&outputFormat=JSON&coordOutputFormat=WGS84[DD.ddddd]&itdLPxx_usage=origin&useLocalityMainStop=true&SpEncId=0&locationServerActive=1&stateless=1&type_sf=any&anyObjFilter_sf=2&anyMaxSizeHitList=1&name_sf='.$haltestelle)->stopFinder->points->point ?? false;
    }

    public static function route($start, $ziel): ?object
    {
        return self::load(sprintf('XSLT_TRIP_REQUEST2?itdDate=%s&itdTime=%s&language=de&sessionID=0&outputFormat=JSON&type_origin=stop&name_origin=%s&type_destination=stop&name_destination=%s', date('Ymd'), date('Hi'), $start, $ziel));
    }

    public static function haltestellen($search): array
    {

        $tmp = [];

        foreach (self::load('XSLT_STOPFINDER_REQUEST?language=de&outputFormat=JSON&coordOutputFormat=WGS84[DD.ddddd]&itdLPxx_usage=origin&useLocalityMainStop=true&SpEncId=0&locationServerActive=1&stateless=1&type_sf=any&anyObjFilter_sf=2&anyMaxSizeHitList=50&name_sf='.$search)->stopFinder->points ?? [] as $haltestelle) {
            $tmp[$haltestelle->ref->gid] = $haltestelle->name;
        }

        return $tmp;

    }

    public static function abfahrt($haltestelle, $limit = 10)
    {
        $gid = self::haltestelle($haltestelle)->ref->gid ?? '';
        return self::load('XML_DM_REQUEST?laguage=de&typeInfo_dm=stopID&deleteAssignedStops_dm=1&useRealtime=1&mode=direct&outputFormat=rapidJSON&limit='.$limit.'&nameInfo_dm='.$gid)->stopEvents ?? [];
    }
}
