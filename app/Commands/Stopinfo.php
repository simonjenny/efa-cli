<?php

namespace App\Commands;

use App\Helpers\Efa;
use Illuminate\Contracts\Console\PromptsForMissingInput;
use Illuminate\Support\Facades\Config;
use LaravelZero\Framework\Commands\Command;

use function Laravel\Prompts\search;
use function Laravel\Prompts\table;

class Stopinfo extends Command implements PromptsForMissingInput
{
    /**
     * The name and signature of the console command.
     *
     * @var string
     */
    protected $signature = 'stopinfo

                            {stop : Show Information for this stop (optional)}

                            {--json : Shows data as JSON (optional)}';

    /**
     * The console command description.
     *
     * @var string
     */
    protected $description = 'Show Information for a stop.';

    /**
     * Aks for missing Information.
     *
     * @var array
     */
    protected function promptForMissingArgumentsUsing(): array
    {
        Config::set('json', true);

        return [
            'stop' => fn () => search(
                label: 'Witch stop do you want to show information for?',
                placeholder: 'Basel, Basel SBB',
                options: fn ($value) => strlen($value) > 0
                    ? Efa::haltestellen("%{$value}%")
                    : []
            ),
        ];
    }

    /**
     * Execute the console command.
     */
    public function handle()
    {

        Config::set('json', $this->option('json'));

        $haltestelle = Efa::haltestelle($this->arguments()['stop']);

        if (config('json')) {

            echo json_encode($haltestelle, JSON_PRETTY_PRINT).PHP_EOL;

        } else {

            $geo = explode(',', $haltestelle->ref->coords);
            $table = [
                ['EFA Stop ID', $haltestelle->ref->id],
                ['GID', $haltestelle->ref->gid],
                ['Coordinates', $haltestelle->ref->coords],
                ['Google Maps', "https://www.google.com/maps/search/?api=1&query={$geo[1]},{$geo[0]}"],
                ['Web Departure Monitor', 'https://dfi.bvb.ch/?point='.$haltestelle->ref->id],
            ];

            $table[] = ['',''];

            if ($haltestelle->infos) {
                $text = '';
                foreach ($haltestelle->infos as $info) {
                    $text .= $info->infoLinkText.PHP_EOL;
                }

                $table[] = [
                    'Info', $text.PHP_EOL.'Detailed information available at http://info.bvb.ch (German only)',
                ];
            }

            table(
                headers: [$haltestelle->name, ''],
                rows: $table
            );

        }
    }
}
