<?php

namespace App\Model;

use App\Enum\ContributionType;

class Contributor {

    protected int $id;

    protected Author $author;
    protected Institute $institute;
    protected Paper $paper;
    protected string $email;
    protected int $displayOrder;

    protected ContributionType $contributorType;

    public function getId(): int
    {
        return $this->id;
    }

    public function setId(int $id): void
    {
        $this->id = $id;
    }

    public function getAuthor(): Author
    {
        return $this->author;
    }

    public function setAuthor(Author $author): void
    {
        $this->author = $author;
    }

    public function getPaper(): Paper
    {
        return $this->paper;
    }

    public function setPaper(Paper $paper): void
    {
        $this->paper = $paper;
    }

    public function getContributorType(): ContributionType
    {
        return $this->contributorType;
    }

    public function setContributorType(ContributionType $contributorType): void
    {
        $this->contributorType = $contributorType;
    }

    public function getInstitute(): Institute
    {
        return $this->institute;
    }

    public function setInstitute(Institute $institute): void
    {
        $this->institute = $institute;
    }

    public function getEmail(): string
    {
        return $this->email;
    }

    public function setEmail(string $email): void
    {
        $this->email = $email;
    }

    public function getDisplayOrder(): int
    {
        return $this->displayOrder;
    }

    public function setDisplayOrder(int $displayOrder): void
    {
        $this->displayOrder = $displayOrder;
    }
}