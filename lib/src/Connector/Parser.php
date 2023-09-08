<?php

namespace App\Connector;

use App\Model\Affiliation;
use App\Model\Author;
use App\Model\Paper;

class Parser
{
    public function normalize(array $papers)
    {
        $institutes = [];
        $authors = [];
        /** @var Paper $paper */
        foreach ($papers as $paper) {
            /** @var Author $author */
            foreach ($paper["authors"] as $author) {
                $authors[$author->getId()] = $author;
                $institutes[(string)$author->getAffiliation()] = $author->getAffiliation();
            }
            /** @var Author $author */
            foreach ($paper["coauthors"] as $author) {
                $authors[$author->getId()] = $author;
                $institutes[(string)$author->getAffiliation()] = $author->getAffiliation();
            }
        }

        return [$papers, $authors, $institutes];
    }

    public function toAuthor($data): Author
    {
        return new Author([
            "_id" => $data["db_id"],
            "affiliation" => new Affiliation($data["affiliation"]),
            "first_name" => $data["first_name"],
            "last_name" => $data["last_name"]
        ]);
    }

    public function toPaper($data): Paper
    {
        return new Paper([
            "_id" => $data["db_id"],
            "code" => $data["code"],
            "title" => $data["title"],
            "authors" => array_map(fn($author) => $this->toAuthor($author), $data["primaryauthors"] ?? []),
            "coauthors" => array_map(fn($author) => $this->toAuthor($author), $data["coauthors"] ?? []),
        ]);
    }

    public function getContributions($contents): array
    {
        $data = json_decode($contents, true);
        $papers = [];
        foreach ($data['results'] as $conference) {
            foreach ($conference['sessions'] as $session) {
                foreach ($session['contributions'] as $entry) {
                    if (isset($entry['code'])
                        && count($entry['primaryauthors']) > 0) {
                        // Ignore if it is on a Sunday????
                        $papers[] = $this->toPaper($entry);
                    }
                }
            }
        }
        return $papers;
    }
}