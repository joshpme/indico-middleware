<?php

use App\Connector\Indico;
use App\Connector\Spaces;

function main(array $args): array
{
    $indico = new Indico($_ENV["INDICO_AUTH"]);
    $event = $args['event'] ?? null;

    if ($event === null) {
        return ['body' => "Not event specified"];
    }

    if (filter_var($event, FILTER_VALIDATE_INT) === false) {
        return ['body' => "Not a valid event id"];
    }

    $contents = $indico->getSessions($event);

    $spaces = new Spaces($_ENV["SPACES_KEY"], $_ENV["SPACES_SECRET"]);
    $spaces->upload($contents, "sessions/$event.json");

    return [
        'body' => "Event ID: $event: Sessions downloaded",
    ];
}