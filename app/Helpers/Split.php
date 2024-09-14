<?php

namespace App\Helpers;

class Split
{
    public static function text($text)
    {

        $text = str_replace("\n", "\n\n", $text);
        $text = wordwrap($text, 74, "\n");

        return $text;

    }
}
