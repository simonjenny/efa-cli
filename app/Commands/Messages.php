<?php

namespace App\Commands;

use App\Helpers\Efa;
use App\Helpers\Split;
use Illuminate\Support\Facades\Config;
use LaravelZero\Framework\Commands\Command;
use Stevebauman\Hypertext\Transformer;

use function Laravel\Prompts\select;
use function Laravel\Prompts\table;

class Messages extends Command
{
    /**
     * The name and signature of the console command.
     *
     * @var string
     */
    protected $signature = 'messages
                            {--json : Shows data as JSON (optional)}';

    /**
     * The console command description.
     *
     * @var string
     */
    protected $description = 'Show current information, disruptions and alerts (currently only from the Basler Verkehrs-Betriebe network in german!)';

    /**
     * Execute the console command.
     */
    public function handle()
    {
        Config::set('json', $this->option('json'));
        $meldungen = Efa::meldungen();

        if (config('json')) {
            echo json_encode($meldungen, JSON_PRETTY_PRINT).PHP_EOL;
        } else {

            $tmp = [];

            foreach ($meldungen as $key => $meldung) {
                $tmp['_'.$key] = $meldung->infoLink->infoLinkText;
            }

            $id = substr(select(
                label: 'Which alert would you like to read?',
                options: $tmp,
                default: 0
            ), 1);

            $text = (new Transformer)->keepNewLines()->toText($meldungen[$id]->infoLink->content);
            table(
                headers: [$meldung->infoLink->infoLinkText],
                rows: [[Split::text($text)]]
            );

        }

    }
}
