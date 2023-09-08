<?php

namespace App\Model;

class Author implements \JsonSerializable {
    protected int $id;
    protected Affiliation $affiliation;
    protected string $first_name;
    protected string $last_name;

    public function __construct($data)
    {
        $this->id = $data["_id"];
        $this->affiliation = $data["affiliation"];
        $this->first_name = $data["first_name"];
        $this->last_name = $data["last_name"];
    }

    public function jsonSerialize()
    {
        return [
            "_id" => $this->id,
            "affiliation" => $this->affiliation,
            "first_name" => $this->first_name,
            "last_name" => $this->last_name
        ];
    }

    public function getId(): int
    {
        return $this->id;
    }

    public function setId(int $id): void
    {
        $this->id = $id;
    }

    public function getAffiliation(): Affiliation
    {
        return $this->affiliation;
    }

    public function setAffiliation(Affiliation $affiliation): void
    {
        $this->affiliation = $affiliation;
    }

    public function getFirstName(): string
    {
        return $this->first_name;
    }

    public function setFirstName(string $first_name): void
    {
        $this->first_name = $first_name;
    }

    public function getLastName(): string
    {
        return $this->last_name;
    }

    public function setLastName(string $last_name): void
    {
        $this->last_name = $last_name;
    }
}