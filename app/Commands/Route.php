<?php

namespace App\Commands;

use App\Helpers\Efa;
use Illuminate\Contracts\Console\PromptsForMissingInput;
use Illuminate\Support\Facades\Config;
use LaravelZero\Framework\Commands\Command;

use function Laravel\Prompts\confirm;
use function Laravel\Prompts\search;
use function Laravel\Prompts\select;
use function Laravel\Prompts\table;
use function Laravel\Prompts\error;

class Route extends Command implements PromptsForMissingInput
{
    /**
     * The name and signature of the console command.
     *
     * @var string
     */
    protected $signature = 'route

                            {start : Starting Stop (optional)}

                            {destination : Destination Stop (optional)}

                            {--json : Shows data as JSON (optional)}';

    /**
     * The console command description.
     *
     * @var string
     */
    protected $description = 'Plan a trip from point A to point B (currently only routes available from the current time and date)';

    /**
     * Aks for missing Information.
     *
     * @var array
     */
    protected function promptForMissingArgumentsUsing(): array
    {
        Config::set('json', true);

        return [
            'start' => fn () => search(
                label: 'Starting Stop:',
                placeholder: 'Basel, Basel SBB',
                options: fn ($value) => strlen($value) > 0
                    ? Efa::haltestellen("%{$value}%")
                    : []
            ),
            'destination' => fn () => search(
                label: 'Destination Stop:',
                placeholder: 'Basel, Claraplatz',
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

        $routes = Efa::route($this->arguments()['start'], $this->arguments()['destination']);

        if (config('json')) {

            echo json_encode($routes, JSON_PRETTY_PRINT).PHP_EOL;

        } else {

            $trips = [];
            $table = [];
            $meldungen = false;

            if($routes->trips) {
                foreach ($routes->trips as $key => $route) {
                    $trips['_'.$key] = sprintf(
                        'Departing from %s at %s with %s transfer(s). Duration: %sh',
                        $route->legs[0]->points[0]->name,
                        $route->legs[0]->points[0]->dateTime->time,
                        $route->interchange,
                        $route->duration);
                }

                $id = substr(select(
                    label: 'Available Trips:',
                    options: $trips,
                    default: 0
                ), 1);

                foreach ($routes->trips[$id]->legs as $leg) {

                    $meldungen = isset($leg->infos->info);

                    $table[] = [
                        explode(' ', $leg->mode->product)[0],
                        $leg->mode->number ?? '',
                        $leg->points[0]->name,
                        $leg->points[0]->dateTime->time.'h',
                        $leg->points[1]->dateTime->time.'h',
                        \Carbon\Carbon::parse($leg->points[0]->dateTime->time)->diffInMinutes(\Carbon\Carbon::parse($leg->points[1]->dateTime->time))."'",
                        $leg->points[1]->name,
                    ];

                }

                table(
                    headers: ['', 'Nr.', 'Boarding', 'Departure', 'Arrival', 'Duration', 'Getting off/Transferring'],
                    rows: $table
                );

                if ($meldungen) {
                    if (confirm('there are alerts for this route? Would you like to see them now?')) {
                        $this->call('messages');
                    }
                }
            } else {
                error('There are no trips available!');
            }



        }

    }
}
