<?php

namespace App\Parser;

use App\Enum\ContributionType;
use App\Model\Author;
use App\Model\Contributor;
use App\Model\Institute;
use App\Model\Paper;

class Session
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
        $author = $this->authors[$data["emailHash"]] ?? new Author();
        $author->setId($data["person_id"]);
        $author->setFirstName($data["first_name"]);
        $author->setLastName($data["last_name"]);
        $this->authors[$data["emailHash"]] = $author;
        return $author;
    }

    private function processPaper($data): void
    {
        $paper = $this->papers[$data['db_id']] ?? new Paper();
        $paper->setId($data["db_id"]);
        $paper->setCode($data["code"]);
        $paper->setTitle($data["title"]);
        $paper->setAbstract($data["description"]);

        foreach ($data["speakers"] as $authorData) {
            $author = $this->author($authorData);
            $institute = $this->institute($authorData);
            $contributor = $this->contributor($authorData,
                $institute,
                ContributionType::Speaker,
                $paper,
                $author);
            $author->addContribution($contributor);
            $paper->addContributor($contributor);
        }

        foreach ($data["primaryauthors"] as $authorData) {
            $author = $this->author($authorData);
            $institute = $this->institute($authorData);
            $contributor = $this->contributor($authorData,
                $institute,
                ContributionType::PrimaryAuthor,
                $paper,
                $author);
            $author->addContribution($contributor);
            $paper->addContributor($contributor);
        }

        foreach ($data["coauthors"] as $authorData) {
            $author = $this->author($authorData);
            $institute = $this->institute($authorData);
            $contributor = $this->contributor($authorData,
                $institute,
                ContributionType::CoAuthor,
                $paper,
                $author);
            $author->addContribution($contributor);
            $paper->addContributor($contributor);
        }
        $this->papers[(string)$data["db_id"]] = $paper;
    }

    public function getContributions($contents): void
    {
        $data = json_decode($contents, true);
        foreach ($data['results'] as $conference) {
            foreach ($conference['sessions'] as $session) {
                foreach ($session['contributions'] as $entry) {
                    if (isset($entry['code'])) {
                        $this->processPaper($entry);
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