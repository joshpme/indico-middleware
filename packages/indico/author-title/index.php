<?php

use App\Connector\Mongo;
use App\Connector\Parser;
use App\Model\Paper;

function main(array $args): array
{
    $indico = new Parser();

    $event = $args['event'] ?? null;

    if ($event === null) {
        return ['body' => "Not event specified"];
    }

    if (filter_var($event, FILTER_VALIDATE_INT) === false) {
        return ['body' => "Not a valid event id"];
    }
    $event = (int)$event;

    $papers = $indico->getContributions($event);

    /** @var Paper $paper */
    foreach ($papers as &$paper) {
        $paper->setEvent($event);
    }

    [$papers, $authors, $institutes] = normalize($papers);

//    $mongo->bulkWrite("authors", $authors);
//    $mongo->bulkWrite("institutes", $institutes);
//    $mongo->bulkReplace(["event" => $event], $papers);

    return [
        'body' => ["Papers downloaded: " . count($papers)],
    ];
}