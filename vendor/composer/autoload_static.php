<?php

// autoload_static.php @generated by Composer

namespace Composer\Autoload;

class ComposerStaticInit4aa6aa083250ca6e7063d9f62c092824
{
    public static $files = array (
        '4b3cea27fe61e047be11cdf836f7229f' => __DIR__ . '/../..' . '/lib/cli/cli.php',
    );

    public static $prefixesPsr0 = array (
        'e' => 
        array (
            'efa' => 
            array (
                0 => __DIR__ . '/../..' . '/src',
            ),
        ),
        'c' => 
        array (
            'cli' => 
            array (
                0 => __DIR__ . '/../..' . '/lib',
            ),
        ),
    );

    public static function getInitializer(ClassLoader $loader)
    {
        return \Closure::bind(function () use ($loader) {
            $loader->prefixesPsr0 = ComposerStaticInit4aa6aa083250ca6e7063d9f62c092824::$prefixesPsr0;

        }, null, ClassLoader::class);
    }
}
