<?php

namespace App\Model;

class Paper implements \JsonSerializable {
    protected int $id;
    protected int $event;
    protected string $code;
    protected string $title;

    /**
     * @var array|Author[]
     */
    protected array $authors;

    /**
     * @var array|Author[]
     */
    protected array $coauthors;
    public function __construct($paper)
    {
        $this->id = $paper["_id"];
        $this->code = $paper["code"];
        $this->title = $paper["title"];
        $this->authors = $paper["authors"];
        $this->coauthors = $paper["coauthors"];
    }


    public function jsonSerialize()
    {
        return [
            "_id" => $this->id,
            "event" => $this->event,
            "paper" => $this->code,
            "title" => $this->title,
            "authors" => $this->authors,
            "coauthors" => $this->coauthors
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

    public function getEvent(): int
    {
        return $this->event;
    }

    public function setEvent(int $event): void
    {
        $this->event = $event;
    }

    public function getTitle(): string
    {
        return $this->title;
    }

    public function setTitle(string $title): void
    {
        $this->title = $title;
    }

    public function getCode(): string
    {
        return $this->code;
    }

    public function setCode(string $code): void
    {
        $this->code = $code;
    }
}