<?php

namespace App\Model;

class Author {

    protected int $id;
    protected string $firstName;
    protected string $lastName;
    protected array $contributions = [];

    public function __construct()
    {

    }

    public function getId(): int
    {
        return $this->id;
    }

    public function setId(int $id): void
    {
        $this->id = $id;
    }

    public function getFirstName(): string
    {
        return $this->firstName;
    }

    public function setFirstName(string $firstName): void
    {
        $this->firstName = $firstName;
    }

    public function getLastName(): string
    {
        return $this->lastName;
    }

    public function setLastName(string $lastName): void
    {
        $this->lastName = $lastName;
    }

    public function getContributions(): array
    {
        return $this->contributions;
    }

    public function setContributions(array $contributions): void
    {
        $this->contributions = $contributions;
    }

    public function addContribution(Contributor $contribution): void
    {
        $this->contributions[] = $contribution;
    }
}