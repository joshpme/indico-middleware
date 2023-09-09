<?php

namespace App\Model;

class Paper {
    protected int $id;
    protected int $event;
    protected string $code;
    protected string $title;
    protected string $abstract;

    /**
     * @var array|Contributor[]
     */
    protected array $contributors;
    public function __construct()
    {
        $this->contributors = [];
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

    public function getAbstract(): string
    {
        return $this->abstract;
    }

    public function setAbstract(string $abstract): void
    {
        $this->abstract = $abstract;
    }

    public function getContributors(): array
    {
        return $this->contributors;
    }

    public function setContributors(array $contributors): void
    {
        $this->contributors = $contributors;
    }

    public function addContributor(Contributor $contributor): void
    {
        $this->contributors[] = $contributor;
    }
}