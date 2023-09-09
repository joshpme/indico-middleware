<?php

namespace App\Model;

class Event implements \JsonSerializable {
    protected int $id;

    public function __construct($data)
    {
        $this->id = $data["_id"];
    }

    public function jsonSerialize()
    {
        return [
            "_id" => $this->id,
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
}