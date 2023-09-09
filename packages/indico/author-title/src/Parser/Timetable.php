<?php

namespace App\Parser;

use App\Enum\ContributionType;
use App\Model\Author;
use App\Model\Contributor;
use App\Model\Institute;
use App\Model\Paper;

class Timetable
{
    public function __construct(
        protected array $papers = [],
        protected array $authors = [],
        protected array $institutes = [],
        protected array $contributors = [])
    {

    }

    private function institute(array $data)
    {
        $institute = $this->institutes[$data["affiliation"]] ?? new Institute();
        $institute->setName($data["affiliation"]);
        $this->institutes[$data["affiliation"]] = $institute;
        return $institute;
    }

    private function contributor(array            $data,
                                 Institute        $institute,
                                 ContributionType $type,
                                 Paper            $paper,
                                 Author           $author)
    {
        $contributor = $this->contributors[$data["db_id"]] ?? new Contributor();
        $contributor->setContributorType($type);
        $contributor->setAuthor($author);
        $contributor->setInstitute($institute);
        $contributor->setEmail($data['email']);
        $contributor->setPaper($paper);
        $this->contributors[$data["db_id"]] = $contributor;
        return $contributor;
    }

    private function author(array $data)
    {
        $author = $this->authors[$data["fullName"]] ?? new Author();
        $author->setId($data["person_id"]);
        $author->setFirstName($data["first_name"]);
        $author->setLastName($data["last_name"]);
        $this->authors[$data["emailHash"]] = $author;
        return $author;
    }

    private function processPaper($data): void
    {
        $paper = $this->papers[$data['contributionId']] ?? new Paper();
        $paper->setCode($data["code"]);
        if ($paper->getTitle() !== $data["title"]) {
            echo "Code: " . $data["code"] . " has different titles: " . $paper->getTitle() . " and " . $data["title"] . "\n";
        }

        $paper->setTitle($data["title"]);


        $paper->setAbstract($data["description"]);
        $contributors = $paper->getContributors() ?? [];
//
//        foreach ($data["presenters"] as $authorData) {
//            $found = false;
//            /** @var Contributor $contributor */
//            foreach ($contributors as $contributor) {
//                if (
//                    $contributor->getAuthor()->getFirstName() == $authorData['first_name'] &&
//                    $contributor->getAuthor()->getLastName() == $authorData['last_name'] &&
//                    $contributor->getInstitute()->getName() == $authorData['affiliation']
//                ) {
//                    $found = true;
//                    $contributor->setDisplayOrder($authorData['displayOrderKey'][0]);
//                }
//                if (!$found) {
//                    echo "Could not find contributor for paper " . $data["code"]. "\n";
//                }
//            }
//        }
//
//        foreach ($data["authors"] as $authorData) {
//            $found = false;
//            /** @var Contributor $contributor */
//            foreach ($contributors as $contributor) {
//                if (
//                    $contributor->getAuthor()->getFirstName() == $authorData['firstName'] &&
//                    $contributor->getAuthor()->getLastName() == $authorData['lastName'] &&
//                    $contributor->getInstitute()->getName() == $authorData['affiliation']
//                ) {
//                    $found = true;
//                    $contributor->setDisplayOrder($authorData['displayOrderKey'][0]);
//
//                }
//                if (!$found) {
//                    echo "Could not find contributor for paper " . $data["code"]. "\n";
//                }
//            }
//        }

        $this->papers[$data["contributionId"]] = $paper;
    }

    public function getContributions($contents): void
    {
        $data = json_decode($contents, true);
        foreach ($data['results'] as $conference) {
            foreach ($conference as $day) {
                foreach ($day as $session) {
                    if (!isset($session['entries'])) {
                        continue;
                    }
                    foreach ($session['entries'] as $entry) {
                        if (isset($entry['code'])) {
                            $this->processPaper($entry);
                        }
                    }

                }
            }
        }
    }

    public function getAuthors(): array
    {
        return $this->authors;
    }

    public function getInstitutes(): array
    {
        return $this->institutes;
    }

    public function getPapers(): array
    {
        return $this->papers;
    }

    public function getContributors(): array
    {
        return $this->contributors;
    }
}