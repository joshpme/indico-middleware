<?php

namespace App\Connector;

use Aws\S3\Exception\S3Exception;
use Aws\S3\S3Client;

class Spaces
{
    private S3Client $client;

    public function __construct($key, $secret)
    {
        $this->client = new S3Client([
            'version' => 'latest',
            'region' => 'ap-southeast-2',
            'endpoint' => 'https://syd1.digitaloceanspaces.com',
            'credentials' => [
                'key' => $key,
                'secret' => $secret,
            ],
        ]);
    }

    // Upload file
    public function upload($file, $name, $bucket = 'indico')
    {
        try {
            $result = $this->client->putObject([
                'Bucket' => $bucket,
                'Key' => $name,
                'Body' => $file,
                'ACL' => 'private'
            ]);
        } catch (S3Exception $e) {
            echo "Error uploading string: " . $e->getMessage();
        }
    }
    // Upload file
    public function read($name, $bucket = 'indico')
    {
        return $this->client->getObject([
            'Bucket' => $bucket,
            'Key' => $name
        ]);
    }


}