<?php

use App\Connector\Spaces;
use App\Parser\Session;
use App\Parser\Timetable;

function main(array $args): array
{
    $event = $args['event'] ?? null;

    if ($event === null) {
        return ['body' => "Not event specified"];
    }

    if (filter_var($event, FILTER_VALIDATE_INT) === false) {
        return ['body' => "Not a valid event id"];
    }

    $spaces = new Spaces($_ENV["SPACES_KEY"], $_ENV["SPACES_SECRET"]);

    $sessions = $spaces->read("sessions/$event.json");
    $sessionParser = new Session();
    $papers = $sessionParser->getContributions($sessions);

    $authors = $sessionParser->getAuthors();
    $institutes = $sessionParser->getInstitutes();
    $papers = $sessionParser->getPapers();
    $contributions = $sessionParser->getContributors();

    $timetable = $spaces->read("timetable/$event.json");
    $timetableParser = new Timetable($authors, $institutes, $papers, $contributions);
    $timetableParser->getContributions($timetable);

//    /** @var Paper $paper */
//    foreach ($papers as $paper) {
//        $paper->setEvent($event);
//    }



    return [
        'body' => [
            "papers" => count($papers),
            "authors" => count($authors),
            "institutes" => count($institutes),
            "contributions" => count($contributions),
        ],
    ];
}