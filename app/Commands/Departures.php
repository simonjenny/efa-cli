<?php

namespace App\Commands;

use App\Helpers\Efa;
use Illuminate\Contracts\Console\PromptsForMissingInput;
use Illuminate\Support\Facades\Config;
use LaravelZero\Framework\Commands\Command;

use function Laravel\Prompts\search;
use function Laravel\Prompts\table;

class Departures extends Command implements PromptsForMissingInput
{
    /**
     * The name and signature of the console command.
     *
     * @var string
     */
    protected $signature = 'departures
                           {stop : Displays the departures for this stop (optional)}
                           {--limit= : Limits the number of displayed departures (default is 10, optional)}
                           {--json : Shows data as JSON (optional)}';

    /**
     * The console command description.
     *
     * @var string
     */
    protected $description = 'Create a departure schedule for a specific bus stop.';

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
                label: 'Witch Stop do you want to see the departures for?',
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
        Config::set('limit', $this->option('limit') ?? 10);

        if (config('json')) {

            echo json_encode(Efa::abfahrt($this->arguments()['stop'], config('limit')), JSON_PRETTY_PRINT).PHP_EOL;

        } else {

            table(
                headers: ['', 'Nr.', 'Destination', 'Departure'],
                rows: Efa::abfahrt($this->arguments()['stop'], config('limit'))
            );

        }

    }
}
