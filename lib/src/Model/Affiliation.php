<?php

namespace App\Model;

class Affiliation implements \JsonSerializable {
    public function __construct(protected string $name)
    {

    }

    public function __toString(): string
    {
        return $this->name;
    }

    public function jsonSerialize()
    {
        return $this->name;
    }
}