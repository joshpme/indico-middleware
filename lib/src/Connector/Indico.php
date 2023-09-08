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

    public function getSessions($eventId) {
        return $this->fetch("https://indico.jacow.org/export/event/$eventId.json?detail=sessions");
    }
}