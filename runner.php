<?php

include_once __DIR__ . "/lib/vendor/autoload.php";

$dotenv = \Dotenv\Dotenv::createImmutable(__DIR__);
$dotenv->load();

include_once __DIR__ . "/packages/indico/" . $argv[1] . "/index.php";

$args = [];
foreach ($argv as $i=>$arg) {
    if ($i < 2)
        continue;
    $e=explode("=",$arg);
    if(count($e)==2)
        $args[$e[0]]=$e[1];
    else
        $args[$e[0]]=0;
}

$results = main($args);
var_dump($results);

// E.g. php runner.php fetch event=41
