<?php

namespace App\Connector;

class Indico
{
    public function __construct(protected string $auth)
    {

    }

    private function fetch($url): string
    {
        $bearerToken = $this->auth;
        $ch = curl_init();
        curl_setopt($ch, CURLOPT_URL, $url);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, 1);
        curl_setopt($ch, CURLOPT_HTTPHEADER, array(
            'Authorization: Bearer ' . $bearerToken
        ));
        $output = curl_exec($ch);
        curl_close($ch);
        return $output;
    }

    // This has whether the author is a primary author or a co-author
    public function getSessions($eventId) {
        return $this->fetch("https://indico.jacow.org/export/event/$eventId.json?detail=sessions");
    }

    // This has the display order and emails of the authors
    public function getTimetable($eventId) {
        return $this->fetch("https://indico.jacow.org/export/timetable/$eventId.json");
    }

    // This has whether or not the talk should be included
    public function getContribution($eventId, $contributionId) {
        return $this->fetch("https://indico.jacow.org/event/$eventId/contributions/$contributionId.json");
    }
}